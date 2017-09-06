package image

import (
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kadmission "k8s.io/apiserver/pkg/admission"
	kapi "k8s.io/kubernetes/pkg/api"
	kquota "k8s.io/kubernetes/pkg/quota"
	"k8s.io/kubernetes/pkg/quota/generic"

	osclient "github.com/openshift/origin/pkg/client"
	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	imageinternalversion "github.com/openshift/origin/pkg/image/generated/listers/image/internalversion"
)

var imageStreamTagResources = []kapi.ResourceName{
	imageapi.ResourceImageStreams,
}

type imageStreamTagEvaluator struct {
	store         imageinternalversion.ImageStreamLister
	istNamespacer osclient.ImageStreamTagsNamespacer
}

// NewImageStreamTagEvaluator computes resource usage of ImageStreamsTags. Its sole purpose is to handle
// UPDATE admission operations on imageStreamTags resource.
func NewImageStreamTagEvaluator(store imageinternalversion.ImageStreamLister, istNamespacer osclient.ImageStreamTagsNamespacer) kquota.Evaluator {
	return &imageStreamTagEvaluator{
		store:         store,
		istNamespacer: istNamespacer,
	}
}

// Constraints checks that given object is an image stream tag
func (i *imageStreamTagEvaluator) Constraints(required []kapi.ResourceName, object runtime.Object) error {
	if _, ok := object.(*imageapi.ImageStreamTag); !ok {
		return fmt.Errorf("unexpected input object %v", object)
	}
	return nil
}

func (i *imageStreamTagEvaluator) GroupKind() schema.GroupKind {
	return imageapi.Kind("ImageStreamTag")
}

func (i *imageStreamTagEvaluator) Handles(operation kadmission.Operation) bool {
	return operation == kadmission.Create || operation == kadmission.Update
}

func (i *imageStreamTagEvaluator) Matches(resourceQuota *kapi.ResourceQuota, item runtime.Object) (bool, error) {
	matchesScopeFunc := func(kapi.ResourceQuotaScope, runtime.Object) (bool, error) { return true, nil }
	return generic.Matches(resourceQuota, item, i.MatchingResources, matchesScopeFunc)
}

func (i *imageStreamTagEvaluator) MatchingResources(input []kapi.ResourceName) []kapi.ResourceName {
	return kquota.Intersection(input, imageStreamTagResources)
}

func (i *imageStreamTagEvaluator) Usage(item runtime.Object) (kapi.ResourceList, error) {
	ist, ok := item.(*imageapi.ImageStreamTag)
	if !ok {
		return kapi.ResourceList{}, nil
	}

	res := map[kapi.ResourceName]resource.Quantity{
		imageapi.ResourceImageStreams: *resource.NewQuantity(0, resource.BinarySI),
	}

	isName, _, err := imageapi.ParseImageStreamTagName(ist.Name)
	if err != nil {
		return kapi.ResourceList{}, err
	}

	is, err := i.store.ImageStreams(ist.Namespace).Get(isName)
	if err != nil && !kerrors.IsNotFound(err) {
		utilruntime.HandleError(fmt.Errorf("failed to get image stream %s/%s: %v", ist.Namespace, isName, err))
	}
	if is == nil || kerrors.IsNotFound(err) {
		res[imageapi.ResourceImageStreams] = *resource.NewQuantity(1, resource.BinarySI)
	}

	return res, nil
}

func (i *imageStreamTagEvaluator) UsageStats(options kquota.UsageStatsOptions) (kquota.UsageStats, error) {
	return kquota.UsageStats{}, nil
}
