package recreate

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	kcoreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	deploytest "github.com/openshift/origin/pkg/deploy/apis/apps/test"
	deployv1 "github.com/openshift/origin/pkg/deploy/apis/apps/v1"
	cmdtest "github.com/openshift/origin/pkg/deploy/cmd/test"
	"github.com/openshift/origin/pkg/deploy/strategy"
	deployutil "github.com/openshift/origin/pkg/deploy/util"

	_ "github.com/openshift/origin/pkg/api/install"
)

type fakeControllerClient struct {
	deployment *kapi.ReplicationController
}

func (c *fakeControllerClient) ReplicationControllers(ns string) kcoreclient.ReplicationControllerInterface {
	return fake.NewSimpleClientset(c.deployment).Core().ReplicationControllers(ns)
}

type fakePodClient struct {
	deployerName string
}

func (c *fakePodClient) Pods(ns string) kcoreclient.PodInterface {
	deployerPod := &kapi.Pod{}
	deployerPod.Name = c.deployerName
	deployerPod.Namespace = ns
	deployerPod.Status = kapi.PodStatus{}
	return fake.NewSimpleClientset(deployerPod).Core().Pods(ns)
}

type hookExecutorImpl struct {
	executeFunc func(hook *deployapi.LifecycleHook, deployment *kapi.ReplicationController, suffix, label string) error
}

func (h *hookExecutorImpl) Execute(hook *deployapi.LifecycleHook, rc *kapi.ReplicationController, suffix, label string) error {
	return h.executeFunc(hook, rc, suffix, label)
}

func TestRecreate_initialDeployment(t *testing.T) {
	var deployment *kapi.ReplicationController
	scaler := &cmdtest.FakeScaler{}
	strategy := &RecreateDeploymentStrategy{
		out:               &bytes.Buffer{},
		errOut:            &bytes.Buffer{},
		decoder:           kapi.Codecs.UniversalDecoder(),
		retryPeriod:       1 * time.Millisecond,
		getUpdateAcceptor: getUpdateAcceptor,
		scaler:            scaler,
		eventClient:       fake.NewSimpleClientset().Core(),
	}

	config := deploytest.OkDeploymentConfig(1)
	config.Spec.Strategy = recreateParams(30, "", "", "")
	deployment, _ = deployutil.MakeDeployment(config, kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))

	strategy.rcClient = &fakeControllerClient{deployment: deployment}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.Deploy(nil, deployment, 3)
	if err != nil {
		t.Fatalf("unexpected deploy error: %#v", err)
	}

	if e, a := 1, len(scaler.Events); e != a {
		t.Fatalf("expected %d scale calls, got %d", e, a)
	}
	if e, a := uint(3), scaler.Events[0].Size; e != a {
		t.Errorf("expected scale up to %d, got %d", e, a)
	}
}

func TestRecreate_deploymentPreHookSuccess(t *testing.T) {
	config := deploytest.OkDeploymentConfig(1)
	config.Spec.Strategy = recreateParams(30, deployapi.LifecycleHookFailurePolicyAbort, "", "")
	deployment, _ := deployutil.MakeDeployment(config, kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	scaler := &cmdtest.FakeScaler{}

	hookExecuted := false
	strategy := &RecreateDeploymentStrategy{
		out:               &bytes.Buffer{},
		errOut:            &bytes.Buffer{},
		decoder:           kapi.Codecs.UniversalDecoder(),
		retryPeriod:       1 * time.Millisecond,
		getUpdateAcceptor: getUpdateAcceptor,
		eventClient:       fake.NewSimpleClientset().Core(),
		rcClient:          &fakeControllerClient{deployment: deployment},
		hookExecutor: &hookExecutorImpl{
			executeFunc: func(hook *deployapi.LifecycleHook, deployment *kapi.ReplicationController, suffix, label string) error {
				hookExecuted = true
				return nil
			},
		},
		scaler: scaler,
	}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.Deploy(nil, deployment, 2)
	if err != nil {
		t.Fatalf("unexpected deploy error: %#v", err)
	}
	if !hookExecuted {
		t.Fatalf("expected hook execution")
	}
}

func TestRecreate_deploymentPreHookFail(t *testing.T) {
	config := deploytest.OkDeploymentConfig(1)
	config.Spec.Strategy = recreateParams(30, deployapi.LifecycleHookFailurePolicyAbort, "", "")
	deployment, _ := deployutil.MakeDeployment(config, kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	scaler := &cmdtest.FakeScaler{}

	strategy := &RecreateDeploymentStrategy{
		out:               &bytes.Buffer{},
		errOut:            &bytes.Buffer{},
		decoder:           kapi.Codecs.UniversalDecoder(),
		retryPeriod:       1 * time.Millisecond,
		getUpdateAcceptor: getUpdateAcceptor,
		eventClient:       fake.NewSimpleClientset().Core(),
		rcClient:          &fakeControllerClient{deployment: deployment},
		hookExecutor: &hookExecutorImpl{
			executeFunc: func(hook *deployapi.LifecycleHook, deployment *kapi.ReplicationController, suffix, label string) error {
				return fmt.Errorf("hook execution failure")
			},
		},
		scaler: scaler,
	}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.Deploy(nil, deployment, 2)
	if err == nil {
		t.Fatalf("expected a deploy error")
	}
	if len(scaler.Events) > 0 {
		t.Fatalf("unexpected scaling events: %v", scaler.Events)
	}
}

func TestRecreate_deploymentMidHookSuccess(t *testing.T) {
	config := deploytest.OkDeploymentConfig(1)
	config.Spec.Strategy = recreateParams(30, "", deployapi.LifecycleHookFailurePolicyAbort, "")
	deployment, _ := deployutil.MakeDeployment(config, kapi.Codecs.LegacyCodec(deployv1.SchemeGroupVersion))
	scaler := &cmdtest.FakeScaler{}

	strategy := &RecreateDeploymentStrategy{
		out:               &bytes.Buffer{},
		errOut:            &bytes.Buffer{},
		decoder:           kapi.Codecs.UniversalDecoder(),
		retryPeriod:       1 * time.Millisecond,
		rcClient:          &fakeControllerClient{deployment: deployment},
		eventClient:       fake.NewSimpleClientset().Core(),
		getUpdateAcceptor: getUpdateAcceptor,
		hookExecutor: &hookExecutorImpl{
			executeFunc: func(hook *deployapi.LifecycleHook, deployment *kapi.ReplicationController, suffix, label string) error {
				return fmt.Errorf("hook execution failure")
			},
		},
		scaler: scaler,
	}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.Deploy(nil, deployment, 2)
	if err == nil {
		t.Fatalf("expected a deploy error")
	}
	if len(scaler.Events) > 0 {
		t.Fatalf("unexpected scaling events: %v", scaler.Events)
	}
}
func TestRecreate_deploymentPostHookSuccess(t *testing.T) {
	config := deploytest.OkDeploymentConfig(1)
	config.Spec.Strategy = recreateParams(30, "", "", deployapi.LifecycleHookFailurePolicyAbort)
	deployment, _ := deployutil.MakeDeployment(config, kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	scaler := &cmdtest.FakeScaler{}

	hookExecuted := false
	strategy := &RecreateDeploymentStrategy{
		out:               &bytes.Buffer{},
		errOut:            &bytes.Buffer{},
		decoder:           kapi.Codecs.UniversalDecoder(),
		retryPeriod:       1 * time.Millisecond,
		rcClient:          &fakeControllerClient{deployment: deployment},
		eventClient:       fake.NewSimpleClientset().Core(),
		getUpdateAcceptor: getUpdateAcceptor,
		hookExecutor: &hookExecutorImpl{
			executeFunc: func(hook *deployapi.LifecycleHook, deployment *kapi.ReplicationController, suffix, label string) error {
				hookExecuted = true
				return nil
			},
		},
		scaler: scaler,
	}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.Deploy(nil, deployment, 2)
	if err != nil {
		t.Fatalf("unexpected deploy error: %#v", err)
	}
	if !hookExecuted {
		t.Fatalf("expected hook execution")
	}
}

func TestRecreate_deploymentPostHookFail(t *testing.T) {
	config := deploytest.OkDeploymentConfig(1)
	config.Spec.Strategy = recreateParams(30, "", "", deployapi.LifecycleHookFailurePolicyAbort)
	deployment, _ := deployutil.MakeDeployment(config, kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	scaler := &cmdtest.FakeScaler{}

	hookExecuted := false
	strategy := &RecreateDeploymentStrategy{
		out:               &bytes.Buffer{},
		errOut:            &bytes.Buffer{},
		decoder:           kapi.Codecs.UniversalDecoder(),
		retryPeriod:       1 * time.Millisecond,
		rcClient:          &fakeControllerClient{deployment: deployment},
		eventClient:       fake.NewSimpleClientset().Core(),
		getUpdateAcceptor: getUpdateAcceptor,
		hookExecutor: &hookExecutorImpl{
			executeFunc: func(hook *deployapi.LifecycleHook, deployment *kapi.ReplicationController, suffix, label string) error {
				hookExecuted = true
				return fmt.Errorf("post hook failure")
			},
		},
		scaler: scaler,
	}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.Deploy(nil, deployment, 2)
	if err == nil {
		t.Fatalf("unexpected non deploy error: %#v", err)
	}
	if !hookExecuted {
		t.Fatalf("expected hook execution")
	}
}

func TestRecreate_acceptorSuccess(t *testing.T) {
	var deployment *kapi.ReplicationController
	scaler := &cmdtest.FakeScaler{}

	strategy := &RecreateDeploymentStrategy{
		out:         &bytes.Buffer{},
		errOut:      &bytes.Buffer{},
		eventClient: fake.NewSimpleClientset().Core(),
		decoder:     kapi.Codecs.UniversalDecoder(),
		retryPeriod: 1 * time.Millisecond,
		scaler:      scaler,
	}

	acceptorCalled := false
	acceptor := &testAcceptor{
		acceptFn: func(deployment *kapi.ReplicationController) error {
			acceptorCalled = true
			return nil
		},
	}

	oldDeployment, _ := deployutil.MakeDeployment(deploytest.OkDeploymentConfig(1), kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	deployment, _ = deployutil.MakeDeployment(deploytest.OkDeploymentConfig(2), kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	strategy.rcClient = &fakeControllerClient{deployment: deployment}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.DeployWithAcceptor(oldDeployment, deployment, 2, acceptor)
	if err != nil {
		t.Fatalf("unexpected deploy error: %#v", err)
	}

	if !acceptorCalled {
		t.Fatalf("expected acceptor to be called")
	}

	if e, a := 2, len(scaler.Events); e != a {
		t.Fatalf("expected %d scale calls, got %d", e, a)
	}
	if e, a := uint(1), scaler.Events[0].Size; e != a {
		t.Errorf("expected scale down to %d, got %d", e, a)
	}
	if e, a := uint(2), scaler.Events[1].Size; e != a {
		t.Errorf("expected scale up to %d, got %d", e, a)
	}
}

func TestRecreate_acceptorSuccessWithColdCaches(t *testing.T) {
	var deployment *kapi.ReplicationController
	scaler := &cmdtest.FakeLaggedScaler{}

	strategy := &RecreateDeploymentStrategy{
		out:         &bytes.Buffer{},
		errOut:      &bytes.Buffer{},
		eventClient: fake.NewSimpleClientset().Core(),
		decoder:     kapi.Codecs.UniversalDecoder(),
		retryPeriod: 1 * time.Millisecond,
		scaler:      scaler,
	}

	acceptorCalled := false
	acceptor := &testAcceptor{
		acceptFn: func(deployment *kapi.ReplicationController) error {
			acceptorCalled = true
			return nil
		},
	}

	oldDeployment, _ := deployutil.MakeDeployment(deploytest.OkDeploymentConfig(1), kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	deployment, _ = deployutil.MakeDeployment(deploytest.OkDeploymentConfig(2), kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	strategy.rcClient = &fakeControllerClient{deployment: deployment}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}

	err := strategy.DeployWithAcceptor(oldDeployment, deployment, 2, acceptor)
	if err != nil {
		t.Fatalf("unexpected deploy error: %#v", err)
	}

	if !acceptorCalled {
		t.Fatalf("expected acceptor to be called")
	}

	if e, a := 2, len(scaler.Events); e != a {
		t.Fatalf("expected %d scale calls, got %d", e, a)
	}
	if e, a := uint(1), scaler.Events[0].Size; e != a {
		t.Errorf("expected scale down to %d, got %d", e, a)
	}
	if e, a := uint(2), scaler.Events[1].Size; e != a {
		t.Errorf("expected scale up to %d, got %d", e, a)
	}
	if scaler.RetryCount != 2 {
		t.Errorf("expected retry when the caches are not initialized")
	}
}

func TestRecreate_acceptorFail(t *testing.T) {
	var deployment *kapi.ReplicationController
	scaler := &cmdtest.FakeScaler{}

	strategy := &RecreateDeploymentStrategy{
		out:         &bytes.Buffer{},
		errOut:      &bytes.Buffer{},
		decoder:     kapi.Codecs.UniversalDecoder(),
		retryPeriod: 1 * time.Millisecond,
		scaler:      scaler,
		eventClient: fake.NewSimpleClientset().Core(),
	}

	acceptor := &testAcceptor{
		acceptFn: func(deployment *kapi.ReplicationController) error {
			return fmt.Errorf("rejected")
		},
	}

	oldDeployment, _ := deployutil.MakeDeployment(deploytest.OkDeploymentConfig(1), kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	deployment, _ = deployutil.MakeDeployment(deploytest.OkDeploymentConfig(2), kapi.Codecs.LegacyCodec(kapi.Registry.GroupOrDie(kapi.GroupName).GroupVersions[0]))
	strategy.rcClient = &fakeControllerClient{deployment: deployment}
	strategy.podClient = &fakePodClient{deployerName: deployutil.DeployerPodNameForDeployment(deployment.Name)}
	err := strategy.DeployWithAcceptor(oldDeployment, deployment, 2, acceptor)
	if err == nil {
		t.Fatalf("expected a deployment failure")
	}
	t.Logf("got expected error: %v", err)

	if e, a := 1, len(scaler.Events); e != a {
		t.Fatalf("expected %d scale calls, got %d", e, a)
	}
	if e, a := uint(1), scaler.Events[0].Size; e != a {
		t.Errorf("expected scale up to %d, got %d", e, a)
	}
}

func recreateParams(timeout int64, preFailurePolicy, midFailurePolicy, postFailurePolicy deployapi.LifecycleHookFailurePolicy) deployapi.DeploymentStrategy {
	var pre, mid, post *deployapi.LifecycleHook
	if len(preFailurePolicy) > 0 {
		pre = &deployapi.LifecycleHook{
			FailurePolicy: preFailurePolicy,
			ExecNewPod:    &deployapi.ExecNewPodHook{},
		}
	}
	if len(midFailurePolicy) > 0 {
		mid = &deployapi.LifecycleHook{
			FailurePolicy: midFailurePolicy,
			ExecNewPod:    &deployapi.ExecNewPodHook{},
		}
	}
	if len(postFailurePolicy) > 0 {
		post = &deployapi.LifecycleHook{
			FailurePolicy: postFailurePolicy,
			ExecNewPod:    &deployapi.ExecNewPodHook{},
		}
	}
	return deployapi.DeploymentStrategy{
		Type: deployapi.DeploymentStrategyTypeRecreate,
		RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
			TimeoutSeconds: &timeout,

			Pre:  pre,
			Mid:  mid,
			Post: post,
		},
	}
}

type testControllerClient struct {
	getReplicationControllerFunc    func(namespace, name string) (*kapi.ReplicationController, error)
	updateReplicationControllerFunc func(namespace string, ctrl *kapi.ReplicationController) (*kapi.ReplicationController, error)
}

func (t *testControllerClient) getReplicationController(namespace, name string) (*kapi.ReplicationController, error) {
	return t.getReplicationControllerFunc(namespace, name)
}

func (t *testControllerClient) updateReplicationController(namespace string, ctrl *kapi.ReplicationController) (*kapi.ReplicationController, error) {
	return t.updateReplicationControllerFunc(namespace, ctrl)
}

func getUpdateAcceptor(timeout time.Duration, minReadySeconds int32) strategy.UpdateAcceptor {
	return &testAcceptor{
		acceptFn: func(deployment *kapi.ReplicationController) error {
			return nil
		},
	}
}

type testAcceptor struct {
	acceptFn func(*kapi.ReplicationController) error
}

func (t *testAcceptor) Accept(deployment *kapi.ReplicationController) error {
	return t.acceptFn(deployment)
}
