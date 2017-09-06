// Package server wraps repository and blob store objects of docker/distribution upstream. Most significantly,
// the wrappers cause manifests to be stored in OpenShift's etcd store instead of registry's storage.
// Registry's middleware API is utilized to register the object factories.
//
// Module with quotaRestrictedBlobStore defines a wrapper for upstream blob store that does an image quota and
// limits check before committing image layer to a registry. Master server contains admission check that will
// refuse the manifest if the image exceeds whatever quota or limit set. But the check occurs too late (after
// the layers are written). This addition allows us to refuse the layers and thus keep the storage clean.
//
// *Note*: Here, we take into account just a single layer, not the image as a whole because the layers are
// uploaded before the manifest. This leads to a situation where several layers can be written until a big
// enough layer will be received that exceeds the limit.
package server

import (
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/openshift/origin/pkg/dockerregistry/server/client"
	imageadmission "github.com/openshift/origin/pkg/image/admission"
)

const (
	defaultProjectCacheTTL = time.Minute
)

// newQuotaEnforcingConfig creates caches for quota objects. The objects are stored with given eviction
// timeout. Caches will only be initialized if the given ttl is positive. Options are gathered from
// configuration file and will be overridden by enforceQuota and projectCacheTTL environment variable values.
func newQuotaEnforcingConfig(ctx context.Context, enforceQuota, projectCacheTTL string, options map[string]interface{}) *quotaEnforcingConfig {
	enforce, err := getBoolOption(EnforceQuotaEnvVar, "enforcequota", false, options)
	if err != nil {
		context.GetLogger(ctx).Error(err)
	}

	if !enforce {
		context.GetLogger(ctx).Info("quota enforcement disabled")
		return &quotaEnforcingConfig{}
	}

	ttl, err := getDurationOption(ProjectCacheTTLEnvVar, "projectcachettl", defaultProjectCacheTTL, options)
	if err != nil {
		context.GetLogger(ctx).Error(err)
	}

	if ttl <= 0 {
		context.GetLogger(ctx).Info("not using project caches for quota objects")
		return &quotaEnforcingConfig{
			enforcementEnabled: true,
		}
	}

	context.GetLogger(ctx).Infof("caching project quota objects with TTL %s", ttl.String())
	return &quotaEnforcingConfig{
		enforcementEnabled: true,
		limitRanges:        newProjectObjectListCache(ttl),
	}
}

// quotaEnforcingConfig holds configuration and caches of object lists keyed by project name. Caches are
// thread safe and shall be reused by all middleware layers.
type quotaEnforcingConfig struct {
	// if set, enables quota enforcement
	enforcementEnabled bool
	// if set, enables caching of quota objects per project
	limitRanges projectObjectListStore
}

// quotaRestrictedBlobStore wraps upstream blob store with a guard preventing big layers exceeding image quotas
// from being saved.
type quotaRestrictedBlobStore struct {
	distribution.BlobStore

	repo *repository
}

var _ distribution.BlobStore = &quotaRestrictedBlobStore{}

// Create wraps returned blobWriter with quota guard wrapper.
func (bs *quotaRestrictedBlobStore) Create(ctx context.Context, options ...distribution.BlobCreateOption) (distribution.BlobWriter, error) {
	context.GetLogger(ctx).Debug("(*quotaRestrictedBlobStore).Create: starting")

	bw, err := bs.BlobStore.Create(ctx, options...)
	if err != nil {
		return nil, err
	}

	repo := (*bs.repo)
	repo.ctx = ctx
	return &quotaRestrictedBlobWriter{
		BlobWriter: bw,
		repo:       &repo,
	}, nil
}

// Resume wraps returned blobWriter with quota guard wrapper.
func (bs *quotaRestrictedBlobStore) Resume(ctx context.Context, id string) (distribution.BlobWriter, error) {
	context.GetLogger(ctx).Debug("(*quotaRestrictedBlobStore).Resume: starting")

	bw, err := bs.BlobStore.Resume(ctx, id)
	if err != nil {
		return nil, err
	}

	repo := (*bs.repo)
	repo.ctx = ctx
	return &quotaRestrictedBlobWriter{
		BlobWriter: bw,
		repo:       &repo,
	}, nil
}

// quotaRestrictedBlobWriter wraps upstream blob writer with a guard preventig big layers exceeding image
// quotas from being written.
type quotaRestrictedBlobWriter struct {
	distribution.BlobWriter

	repo *repository
}

func (bw *quotaRestrictedBlobWriter) Commit(ctx context.Context, provisional distribution.Descriptor) (canonical distribution.Descriptor, err error) {
	context.GetLogger(ctx).Debug("(*quotaRestrictedBlobWriter).Commit: starting")

	if err := admitBlobWrite(ctx, bw.repo, provisional.Size); err != nil {
		return distribution.Descriptor{}, err
	}

	return bw.BlobWriter.Commit(ctx, provisional)
}

// getLimitRangeList returns list of limit ranges for repo.
func getLimitRangeList(ctx context.Context, limitClient client.LimitRangesGetter, namespace string) (*kapi.LimitRangeList, error) {
	if quotaEnforcing.limitRanges != nil {
		obj, exists, _ := quotaEnforcing.limitRanges.get(namespace)
		if exists {
			return obj.(*kapi.LimitRangeList), nil
		}
	}

	context.GetLogger(ctx).Debugf("listing limit ranges in namespace %s", namespace)

	lrs, err := limitClient.LimitRanges(namespace).List(metav1.ListOptions{})
	if err != nil {
		context.GetLogger(ctx).Errorf("failed to list limitranges: %v", err)
		return nil, err
	}

	if quotaEnforcing.limitRanges != nil {
		err = quotaEnforcing.limitRanges.add(namespace, lrs)
		if err != nil {
			context.GetLogger(ctx).Errorf("failed to cache limit range list: %v", err)
		}
	}

	return lrs, nil
}

// admitBlobWrite checks whether the blob does not exceed image limit ranges if set. Returns ErrAccessDenied
// error if the limit is exceeded.
func admitBlobWrite(ctx context.Context, repo *repository, size int64) error {
	if size < 1 {
		return nil
	}

	lrs, err := getLimitRangeList(ctx, repo.registryOSClient, repo.namespace)
	if err != nil {
		return err
	}

	for _, limitrange := range lrs.Items {
		context.GetLogger(ctx).Debugf("processing limit range %s/%s", limitrange.Namespace, limitrange.Name)
		for _, limit := range limitrange.Spec.Limits {
			if err := imageadmission.AdmitImage(size, limit); err != nil {
				context.GetLogger(ctx).Errorf("refusing to write blob exceeding limit range %s: %s", limitrange.Name, err.Error())
				return distribution.ErrAccessDenied
			}
		}
	}

	// TODO(1): admit also against openshift.io/ImageStream quota resource when we have image stream cache in the
	// registry
	// TODO(2): admit also against openshift.io/imagestreamimages and openshift.io/imagestreamtags resources once
	// we have image stream cache in the registry

	return nil
}
