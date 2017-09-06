package recreate

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	kapi "k8s.io/kubernetes/pkg/api"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	kcoreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/kubectl"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	strat "github.com/openshift/origin/pkg/deploy/strategy"
	stratsupport "github.com/openshift/origin/pkg/deploy/strategy/support"
	stratutil "github.com/openshift/origin/pkg/deploy/strategy/util"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
	imageclient "github.com/openshift/origin/pkg/image/generated/internalclientset/typed/image/internalversion"
)

// RecreateDeploymentStrategy is a simple strategy appropriate as a default.
// Its behavior is to scale down the last deployment to 0, and to scale up the
// new deployment to 1.
//
// A failure to disable any existing deployments will be considered a
// deployment failure.
type RecreateDeploymentStrategy struct {
	// out and errOut control where output is sent during the strategy
	out, errOut io.Writer
	// until is a condition that, if reached, will cause the strategy to exit early
	until string
	// rcClient is a client to access replication controllers
	rcClient kcoreclient.ReplicationControllersGetter
	// podClient is used to list and watch pods.
	podClient kcoreclient.PodsGetter
	// eventClient is a client to access events
	eventClient kcoreclient.EventsGetter
	// getUpdateAcceptor returns an UpdateAcceptor to verify the first replica
	// of the deployment.
	getUpdateAcceptor func(time.Duration, int32) strat.UpdateAcceptor
	// scaler is used to scale replication controllers.
	scaler kubectl.Scaler
	// tagClient is used to tag images
	tagClient imageclient.ImageStreamTagsGetter
	// codec is used to decode DeploymentConfigs contained in deployments.
	decoder runtime.Decoder
	// hookExecutor can execute a lifecycle hook.
	hookExecutor stratsupport.HookExecutor
	// retryPeriod is how often to try updating the replica count.
	retryPeriod time.Duration
	// retryParams encapsulates the retry parameters
	retryParams *kubectl.RetryParams
	// events records the events
	events record.EventSink
	// now returns the current time
	now func() time.Time
}

const (
	// acceptorInterval is how often the UpdateAcceptor should check for
	// readiness.
	acceptorInterval = 1 * time.Second
)

// NewRecreateDeploymentStrategy makes a RecreateDeploymentStrategy backed by
// a real HookExecutor and client.
func NewRecreateDeploymentStrategy(client kclientset.Interface, tagClient imageclient.ImageStreamTagsGetter, events record.EventSink, decoder runtime.Decoder, out, errOut io.Writer, until string) *RecreateDeploymentStrategy {
	if out == nil {
		out = ioutil.Discard
	}
	if errOut == nil {
		errOut = ioutil.Discard
	}
	scaler, _ := kubectl.ScalerFor(kapi.Kind("ReplicationController"), client)
	return &RecreateDeploymentStrategy{
		out:         out,
		errOut:      errOut,
		events:      events,
		until:       until,
		rcClient:    client.Core(),
		eventClient: client.Core(),
		podClient:   client.Core(),
		getUpdateAcceptor: func(timeout time.Duration, minReadySeconds int32) strat.UpdateAcceptor {
			return stratsupport.NewAcceptAvailablePods(out, client.Core(), timeout)
		},
		scaler:       scaler,
		decoder:      decoder,
		hookExecutor: stratsupport.NewHookExecutor(client.Core(), tagClient, client.Core(), os.Stdout, decoder),
		retryPeriod:  1 * time.Second,
	}
}

// Deploy makes deployment active and disables oldDeployments.
func (s *RecreateDeploymentStrategy) Deploy(from *kapi.ReplicationController, to *kapi.ReplicationController, desiredReplicas int) error {
	return s.DeployWithAcceptor(from, to, desiredReplicas, nil)
}

// DeployWithAcceptor scales down from and then scales up to. If
// updateAcceptor is provided and the desired replica count is >1, the first
// replica of to is rolled out and validated before performing the full scale
// up.
//
// This is currently only used in conjunction with the rolling update strategy
// for initial deployments.
func (s *RecreateDeploymentStrategy) DeployWithAcceptor(from *kapi.ReplicationController, to *kapi.ReplicationController, desiredReplicas int, updateAcceptor strat.UpdateAcceptor) error {
	config, err := deployutil.DecodeDeploymentConfig(to, s.decoder)
	if err != nil {
		return fmt.Errorf("couldn't decode config from deployment %s: %v", to.Name, err)
	}

	retryTimeout := time.Duration(deployapi.DefaultRecreateTimeoutSeconds) * time.Second
	params := config.Spec.Strategy.RecreateParams
	rollingParams := config.Spec.Strategy.RollingParams

	if params != nil && params.TimeoutSeconds != nil {
		retryTimeout = time.Duration(*params.TimeoutSeconds) * time.Second
	}

	// When doing the initial rollout for rolling strategy we use recreate and for that we
	// have to set the TimeoutSecond based on the rollling strategy parameters.
	if rollingParams != nil && rollingParams.TimeoutSeconds != nil {
		retryTimeout = time.Duration(*rollingParams.TimeoutSeconds) * time.Second
	}

	s.retryParams = kubectl.NewRetryParams(s.retryPeriod, retryTimeout)
	waitParams := kubectl.NewRetryParams(s.retryPeriod, retryTimeout)

	if updateAcceptor == nil {
		updateAcceptor = s.getUpdateAcceptor(retryTimeout, config.Spec.MinReadySeconds)
	}

	// Execute any pre-hook.
	if params != nil && params.Pre != nil {
		if err := s.hookExecutor.Execute(params.Pre, to, deployapi.PreHookPodSuffix, "pre"); err != nil {
			return fmt.Errorf("pre hook failed: %s", err)
		}
	}

	if s.until == "pre" {
		return strat.NewConditionReachedErr("pre hook succeeded")
	}

	// Record all warnings
	defer stratutil.RecordConfigWarnings(s.eventClient, from, s.decoder, s.out)
	defer stratutil.RecordConfigWarnings(s.eventClient, to, s.decoder, s.out)

	// Scale down the from deployment.
	if from != nil {
		fmt.Fprintf(s.out, "--> Scaling %s down to zero\n", from.Name)
		_, err := s.scaleAndWait(from, 0, s.retryParams, waitParams)
		if err != nil {
			return fmt.Errorf("couldn't scale %s to 0: %v", from.Name, err)
		}
		// Wait for pods to terminate.
		s.waitForTerminatedPods(from, time.Duration(*params.TimeoutSeconds)*time.Second)
	}

	if s.until == "0%" {
		return strat.NewConditionReachedErr("Reached 0% (no running pods)")
	}

	if params != nil && params.Mid != nil {
		if err := s.hookExecutor.Execute(params.Mid, to, deployapi.MidHookPodSuffix, "mid"); err != nil {
			return fmt.Errorf("mid hook failed: %s", err)
		}
	}

	if s.until == "mid" {
		return strat.NewConditionReachedErr("mid hook succeeded")
	}

	accepted := false

	// Scale up the to deployment.
	if desiredReplicas > 0 {
		if from != nil {
			// Scale up to 1 and validate the replica,
			// aborting if the replica isn't acceptable.
			fmt.Fprintf(s.out, "--> Scaling %s to 1 before performing acceptance check\n", to.Name)
			updatedTo, err := s.scaleAndWait(to, 1, s.retryParams, waitParams)
			if err != nil {
				return fmt.Errorf("couldn't scale %s to 1: %v", to.Name, err)
			}
			if err := updateAcceptor.Accept(updatedTo); err != nil {
				return fmt.Errorf("update acceptor rejected %s: %v", to.Name, err)
			}
			accepted = true
			to = updatedTo

			if strat.PercentageBetween(s.until, 1, 99) {
				return strat.NewConditionReachedErr(fmt.Sprintf("Reached %s", s.until))
			}
		}

		// Complete the scale up.
		if to.Spec.Replicas != int32(desiredReplicas) {
			fmt.Fprintf(s.out, "--> Scaling %s to %d\n", to.Name, desiredReplicas)
			updatedTo, err := s.scaleAndWait(to, desiredReplicas, s.retryParams, waitParams)
			if err != nil {
				return fmt.Errorf("couldn't scale %s to %d: %v", to.Name, desiredReplicas, err)
			}

			to = updatedTo
		}

		if !accepted {
			if err := updateAcceptor.Accept(to); err != nil {
				return fmt.Errorf("update acceptor rejected %s: %v", to.Name, err)
			}
		}
	}

	if (from == nil && strat.PercentageBetween(s.until, 1, 100)) || (from != nil && s.until == "100%") {
		return strat.NewConditionReachedErr(fmt.Sprintf("Reached %s", s.until))
	}

	// Execute any post-hook.
	if params != nil && params.Post != nil {
		if err := s.hookExecutor.Execute(params.Post, to, deployapi.PostHookPodSuffix, "post"); err != nil {
			return fmt.Errorf("post hook failed: %s", err)
		}
	}

	return nil
}

func (s *RecreateDeploymentStrategy) scaleAndWait(deployment *kapi.ReplicationController, replicas int, retry *kubectl.RetryParams, retryParams *kubectl.RetryParams) (*kapi.ReplicationController, error) {
	if int32(replicas) == deployment.Spec.Replicas && int32(replicas) == deployment.Status.Replicas {
		return deployment, nil
	}
	var scaleErr error
	err := wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
		scaleErr = s.scaler.Scale(deployment.Namespace, deployment.Name, uint(replicas), &kubectl.ScalePrecondition{Size: -1, ResourceVersion: ""}, retry, retryParams)
		if scaleErr == nil {
			return true, nil
		}
		// This error is returned when the lifecycle admission plugin cache is not fully
		// synchronized. In that case the scaling should be retried.
		//
		// FIXME: The error returned from admission should not be forbidden but come-back-later error.
		if errors.IsForbidden(scaleErr) && strings.Contains(scaleErr.Error(), "not yet ready to handle request") {
			return false, nil
		}
		return false, scaleErr
	})
	if err == wait.ErrWaitTimeout {
		return nil, fmt.Errorf("%v: %v", err, scaleErr)
	}
	if err != nil {
		return nil, err
	}

	return s.rcClient.ReplicationControllers(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{})
}

// waitForTerminatedPods waits until all pods for the provided replication controller are terminated.
func (s *RecreateDeploymentStrategy) waitForTerminatedPods(from *kapi.ReplicationController, timeout time.Duration) {
	selector := labels.Set(from.Spec.Selector).AsSelector()
	options := metav1.ListOptions{LabelSelector: selector.String()}
	podList, err := s.podClient.Pods(from.Namespace).List(options)
	if err != nil {
		fmt.Fprintf(s.out, "--> Cannot list pods: %v\nNew pods may be scaled up before old pods terminate\n", err)
		return
	}
	// If there are no pods left, we are done.
	if len(podList.Items) == 0 {
		return
	}
	// Watch from the resource version of the list and wait for all pods to be deleted
	// before proceeding with the Recreate strategy.
	options.ResourceVersion = podList.ResourceVersion
	w, err := s.podClient.Pods(from.Namespace).Watch(options)
	if err != nil {
		fmt.Fprintf(s.out, "--> Watch could not be established: %v\nNew pods may be scaled up before old pods terminate\n", err)
		return
	}
	defer w.Stop()
	// Observe as many deletions as the remaining pods and then return.
	deletionsNeeded := len(podList.Items)
	condition := func(event watch.Event) (bool, error) {
		if event.Type == watch.Deleted {
			deletionsNeeded--
		}
		return deletionsNeeded == 0, nil
	}
	// TODO: Timeout should be timeout - (time.Now - deployerPodStartTime)
	if _, err = watch.Until(timeout, w, condition); err != nil && err != wait.ErrWaitTimeout {
		fmt.Fprintf(s.out, "--> Watch failed: %v\nNew pods may be scaled up before old pods terminate\n", err)
	}
	return
}
