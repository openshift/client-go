package cmd

import (
	"time"

	"github.com/golang/glog"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kapi "k8s.io/kubernetes/pkg/api"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/retry"
	"k8s.io/kubernetes/pkg/kubectl"
	kutil "k8s.io/kubernetes/pkg/util"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	appsclient "github.com/openshift/origin/pkg/deploy/generated/internalclientset"
	appsinternal "github.com/openshift/origin/pkg/deploy/generated/internalclientset/typed/apps/internalversion"
	"github.com/openshift/origin/pkg/deploy/util"
)

// NewDeploymentConfigReaper returns a new reaper for deploymentConfigs
func NewDeploymentConfigReaper(appsClient appsclient.Interface, kc kclientset.Interface) kubectl.Reaper {
	return &DeploymentConfigReaper{appsClient: appsClient, kc: kc, pollInterval: kubectl.Interval, timeout: kubectl.Timeout}
}

// DeploymentConfigReaper implements the Reaper interface for deploymentConfigs
type DeploymentConfigReaper struct {
	appsClient            appsclient.Interface
	kc                    kclientset.Interface
	pollInterval, timeout time.Duration
}

type updateConfigFunc func(d *deployapi.DeploymentConfig)

// updateConfigWithRetries will try to update a deployment config and ignore any update conflicts.
func updateConfigWithRetries(dn appsinternal.DeploymentConfigsGetter, namespace, name string, applyUpdate updateConfigFunc) (*deployapi.DeploymentConfig, error) {
	var config *deployapi.DeploymentConfig

	resultErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var err error
		config, err = dn.DeploymentConfigs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		// Apply the update, then attempt to push it to the apiserver.
		applyUpdate(config)
		config, err = dn.DeploymentConfigs(namespace).Update(config)
		return err
	})
	return config, resultErr
}

// pause marks the deployment configuration as paused to avoid triggering new
// deployments.
func (reaper *DeploymentConfigReaper) pause(namespace, name string) (*deployapi.DeploymentConfig, error) {
	return updateConfigWithRetries(reaper.appsClient.Apps(), namespace, name, func(d *deployapi.DeploymentConfig) {
		d.Spec.RevisionHistoryLimit = kutil.Int32Ptr(0)
		d.Spec.Replicas = 0
		d.Spec.Paused = true
	})
}

// Stop scales a replication controller via its deployment configuration down to
// zero replicas, waits for all of them to get deleted and then deletes both the
// replication controller and its deployment configuration.
func (reaper *DeploymentConfigReaper) Stop(namespace, name string, timeout time.Duration, gracePeriod *metav1.DeleteOptions) error {
	// Pause the deployment configuration to prevent the new deployments from
	// being triggered.
	config, err := reaper.pause(namespace, name)
	configNotFound := kerrors.IsNotFound(err)
	if err != nil && !configNotFound {
		return err
	}

	var (
		isPaused bool
		legacy   bool
	)
	// Determine if the deployment config controller noticed the pause.
	if !configNotFound {
		if err := wait.Poll(1*time.Second, 1*time.Minute, func() (bool, error) {
			dc, err := reaper.appsClient.Apps().DeploymentConfigs(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			isPaused = dc.Spec.Paused
			return dc.Status.ObservedGeneration >= config.Generation, nil
		}); err != nil {
			return err
		}

		// If we failed to pause the deployment config, it means we are talking to
		// old API that does not support pausing. In that case, we delete the
		// deployment config to stay backward compatible.
		if !isPaused {
			if err := reaper.appsClient.Apps().DeploymentConfigs(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
			// Setting this to true avoid deleting the config at the end.
			legacy = true
		}
	}

	// Clean up deployments related to the config. Even if the deployment
	// configuration has been deleted, we want to sweep the existing replication
	// controllers and clean them up.
	options := metav1.ListOptions{LabelSelector: util.ConfigSelector(name).String()}
	rcList, err := reaper.kc.Core().ReplicationControllers(namespace).List(options)
	if err != nil {
		return err
	}
	rcReaper, err := kubectl.ReaperFor(kapi.Kind("ReplicationController"), reaper.kc)
	if err != nil {
		return err
	}

	// If there is neither a config nor any deployments, nor any deployer pods, we can return NotFound.
	deployments := rcList.Items

	if configNotFound && len(deployments) == 0 {
		return kerrors.NewNotFound(kapi.Resource("deploymentconfig"), name)
	}

	for _, rc := range deployments {
		if err = rcReaper.Stop(rc.Namespace, rc.Name, timeout, gracePeriod); err != nil {
			// Better not error out here...
			glog.Infof("Cannot delete ReplicationController %s/%s for deployment config %s/%s: %v", rc.Namespace, rc.Name, namespace, name, err)
		}

		// Only remove deployer pods when the deployment was failed. For completed
		// deployment the pods should be already deleted.
		if !util.IsFailedDeployment(&rc) {
			continue
		}

		// Delete all deployer and hook pods
		options = metav1.ListOptions{LabelSelector: util.DeployerPodSelector(rc.Name).String()}
		podList, err := reaper.kc.Core().Pods(rc.Namespace).List(options)
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			err := reaper.kc.Core().Pods(pod.Namespace).Delete(pod.Name, gracePeriod)
			if err != nil {
				// Better not error out here...
				glog.Infof("Cannot delete lifecycle Pod %s/%s for deployment config %s/%s: %v", pod.Namespace, pod.Name, namespace, name, err)
			}
		}
	}

	// Nothing to delete or we already deleted the deployment config because we
	// failed to pause.
	if configNotFound || legacy {
		return nil
	}

	return reaper.appsClient.Apps().DeploymentConfigs(namespace).Delete(name, &metav1.DeleteOptions{})
}
