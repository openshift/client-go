package cmd

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/kubectl"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	deployutil "github.com/openshift/origin/pkg/deploy/util"

	appsclient "github.com/openshift/origin/pkg/deploy/generated/internalclientset"
	appsinternal "github.com/openshift/origin/pkg/deploy/generated/internalclientset/typed/apps/internalversion"
)

func NewDeploymentConfigStatusViewer(appsClient appsclient.Interface) kubectl.StatusViewer {
	return &DeploymentConfigStatusViewer{dn: appsClient.Apps()}
}

// DeploymentConfigStatusViewer is an implementation of the kubectl StatusViewer interface
// for deployment configs.
type DeploymentConfigStatusViewer struct {
	dn appsinternal.DeploymentConfigsGetter
}

var _ kubectl.StatusViewer = &DeploymentConfigStatusViewer{}

// Status returns a message describing deployment status, and a bool value indicating if the status is considered done
func (s *DeploymentConfigStatusViewer) Status(namespace, name string, desiredRevision int64) (string, bool, error) {
	config, err := s.dn.DeploymentConfigs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", false, err
	}
	latestRevision := config.Status.LatestVersion

	if latestRevision == 0 {
		switch {
		case deployutil.HasImageChangeTrigger(config):
			return fmt.Sprintf("Deployment config %q waiting on image update\n", name), false, nil

		case len(config.Spec.Triggers) == 0:
			return "", true, fmt.Errorf("Deployment config %q waiting on manual update (use 'oc rollout latest %s')", name, name)
		}
	}

	if desiredRevision > 0 && latestRevision != desiredRevision {
		return "", false, fmt.Errorf("desired revision (%d) is different from the running revision (%d)", desiredRevision, latestRevision)
	}

	cond := deployutil.GetDeploymentCondition(config.Status, deployapi.DeploymentProgressing)

	if config.Generation <= config.Status.ObservedGeneration {
		switch {
		case cond != nil && cond.Reason == deployapi.NewRcAvailableReason:
			return fmt.Sprintf("%s\n", cond.Message), true, nil

		case cond != nil && cond.Reason == deployapi.TimedOutReason:
			return "", true, errors.New(cond.Message)

		case cond != nil && cond.Reason == deployapi.CancelledRolloutReason:
			return "", true, errors.New(cond.Message)

		case cond != nil && cond.Reason == deployapi.PausedConfigReason:
			return "", true, fmt.Errorf("Deployment config %q is paused. Resume to continue watching the status of the rollout.\n", config.Name)

		case config.Status.UpdatedReplicas < config.Spec.Replicas:
			return fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...\n", config.Status.UpdatedReplicas, config.Spec.Replicas), false, nil

		case config.Status.Replicas > config.Status.UpdatedReplicas:
			return fmt.Sprintf("Waiting for rollout to finish: %d old replicas are pending termination...\n", config.Status.Replicas-config.Status.UpdatedReplicas), false, nil

		case config.Status.AvailableReplicas < config.Status.UpdatedReplicas:
			return fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...\n", config.Status.AvailableReplicas, config.Status.UpdatedReplicas), false, nil
		}
	}
	return fmt.Sprintf("Waiting for latest deployment config spec to be observed by the controller loop...\n"), false, nil
}
