package validation

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kapi "k8s.io/kubernetes/pkg/api"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	"github.com/openshift/origin/pkg/deploy/apis/apps/test"
)

// Convenience methods

func manualTrigger() []deployapi.DeploymentTriggerPolicy {
	return []deployapi.DeploymentTriggerPolicy{
		{
			Type: deployapi.DeploymentTriggerManual,
		},
	}
}

func rollingConfig(interval, updatePeriod, timeout int) deployapi.DeploymentConfig {
	return deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
		Spec: deployapi.DeploymentConfigSpec{
			Triggers: manualTrigger(),
			Strategy: deployapi.DeploymentStrategy{
				Type: deployapi.DeploymentStrategyTypeRolling,
				RollingParams: &deployapi.RollingDeploymentStrategyParams{
					IntervalSeconds:     mkint64p(interval),
					UpdatePeriodSeconds: mkint64p(updatePeriod),
					TimeoutSeconds:      mkint64p(timeout),
					MaxSurge:            intstr.FromInt(1),
				},
				ActiveDeadlineSeconds: mkint64p(3600),
			},
			Template: test.OkPodTemplate(),
			Selector: test.OkSelector(),
		},
	}
}

func rollingConfigMax(maxSurge, maxUnavailable intstr.IntOrString) deployapi.DeploymentConfig {
	return deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
		Spec: deployapi.DeploymentConfigSpec{
			Triggers: manualTrigger(),
			Strategy: deployapi.DeploymentStrategy{
				Type: deployapi.DeploymentStrategyTypeRolling,
				RollingParams: &deployapi.RollingDeploymentStrategyParams{
					IntervalSeconds:     mkint64p(1),
					UpdatePeriodSeconds: mkint64p(1),
					TimeoutSeconds:      mkint64p(1),
					MaxSurge:            maxSurge,
					MaxUnavailable:      maxUnavailable,
				},
				ActiveDeadlineSeconds: mkint64p(3600),
			},
			Template: test.OkPodTemplate(),
			Selector: test.OkSelector(),
		},
	}
}

func TestValidateDeploymentConfigOK(t *testing.T) {
	errs := ValidateDeploymentConfig(&deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: manualTrigger(),
			Selector: test.OkSelector(),
			Strategy: test.OkStrategy(),
			Template: test.OkPodTemplate(),
		},
	})

	if len(errs) > 0 {
		t.Errorf("Unxpected non-empty error list: %s", errs)
	}
}

func TestValidateDeploymentConfigICTMissingImage(t *testing.T) {
	dc := &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: []deployapi.DeploymentTriggerPolicy{test.OkImageChangeTrigger()},
			Selector: test.OkSelector(),
			Strategy: test.OkStrategy(),
			Template: test.OkPodTemplateMissingImage("container1"),
		},
	}
	errs := ValidateDeploymentConfig(dc)

	if len(errs) > 0 {
		t.Errorf("Unexpected non-empty error list: %+v", errs)
	}

	for _, c := range dc.Spec.Template.Spec.Containers {
		if c.Image == "unset" {
			t.Errorf("%s image field still has validation fake out value of %s", c.Name, c.Image)
		}
	}
}

func TestValidateDeploymentConfigMissingFields(t *testing.T) {
	errorCases := map[string]struct {
		DeploymentConfig deployapi.DeploymentConfig
		ErrorType        field.ErrorType
		Field            string
	}{
		"empty container field": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: []deployapi.DeploymentTriggerPolicy{test.OkConfigChangeTrigger()},
					Selector: test.OkSelector(),
					Strategy: test.OkStrategy(),
					Template: test.OkPodTemplateMissingImage("container1"),
				},
			},
			field.ErrorTypeRequired,
			"spec.template.spec.containers[0].image",
		},
		"missing name": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: "bar"},
				Spec:       test.OkDeploymentConfigSpec(),
			},
			field.ErrorTypeRequired,
			"metadata.name",
		},
		"missing namespace": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: ""},
				Spec:       test.OkDeploymentConfigSpec(),
			},
			field.ErrorTypeRequired,
			"metadata.namespace",
		},
		"invalid name": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "-foo", Namespace: "bar"},
				Spec:       test.OkDeploymentConfigSpec(),
			},
			field.ErrorTypeInvalid,
			"metadata.name",
		},
		"invalid namespace": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "-bar"},
				Spec:       test.OkDeploymentConfigSpec(),
			},
			field.ErrorTypeInvalid,
			"metadata.namespace",
		},

		"missing trigger.type": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: []deployapi.DeploymentTriggerPolicy{
						{
							ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
								ContainerNames: []string{"foo"},
							},
						},
					},
					Selector: test.OkSelector(),
					Strategy: test.OkStrategy(),
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeRequired,
			"spec.triggers[0].type",
		},
		"missing Trigger imageChangeParams.from": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: []deployapi.DeploymentTriggerPolicy{
						{
							Type: deployapi.DeploymentTriggerOnImageChange,
							ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
								ContainerNames: []string{"foo"},
							},
						},
					},
					Selector: test.OkSelector(),
					Strategy: test.OkStrategy(),
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeRequired,
			"spec.triggers[0].imageChangeParams.from",
		},
		"invalid Trigger imageChangeParams.from.kind": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: []deployapi.DeploymentTriggerPolicy{
						{
							Type: deployapi.DeploymentTriggerOnImageChange,
							ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
								From: kapi.ObjectReference{
									Kind: "Invalid",
									Name: "name:tag",
								},
								ContainerNames: []string{"foo"},
							},
						},
					},
					Selector: test.OkSelector(),
					Strategy: test.OkStrategy(),
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.triggers[0].imageChangeParams.from.kind",
		},
		"missing Trigger imageChangeParams.containerNames": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: []deployapi.DeploymentTriggerPolicy{
						{
							Type: deployapi.DeploymentTriggerOnImageChange,
							ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
								From: kapi.ObjectReference{
									Kind: "ImageStreamTag",
									Name: "foo:v1",
								},
							},
						},
					},
					Selector: test.OkSelector(),
					Strategy: test.OkStrategy(),
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeRequired,
			"spec.triggers[0].imageChangeParams.containerNames",
		},
		"missing strategy.type": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: manualTrigger(),
					Selector: test.OkSelector(),
					Strategy: deployapi.DeploymentStrategy{
						CustomParams:          test.OkCustomParams(),
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.type",
		},
		"missing strategy.customParams": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: manualTrigger(),
					Selector: test.OkSelector(),
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeCustom,
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.customParams",
		},
		"invalid spec.strategy.customParams.environment": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Triggers: manualTrigger(),
					Selector: test.OkSelector(),
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeCustom,
						CustomParams: &deployapi.CustomDeploymentStrategyParams{
							Environment: []kapi.EnvVar{
								{Name: "A=B"},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.customParams.environment[0].name",
		},
		"missing spec.strategy.recreateParams.pre.failurePolicy": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Pre: &deployapi.LifecycleHook{
								ExecNewPod: &deployapi.ExecNewPodHook{
									Command:       []string{"cmd"},
									ContainerName: "container",
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.recreateParams.pre.failurePolicy",
		},
		"missing spec.strategy.recreateParams.pre.execNewPod": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Pre: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.pre",
		},
		"missing spec.strategy.recreateParams.pre.execNewPod.command": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Pre: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								ExecNewPod: &deployapi.ExecNewPodHook{
									ContainerName: "container",
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.recreateParams.pre.execNewPod.command",
		},
		"missing spec.strategy.recreateParams.pre.execNewPod.containerName": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Pre: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								ExecNewPod: &deployapi.ExecNewPodHook{
									Command: []string{"cmd"},
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.recreateParams.pre.execNewPod.containerName",
		},
		"invalid spec.strategy.recreateParams.pre.execNewPod.volumes": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Pre: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								ExecNewPod: &deployapi.ExecNewPodHook{
									ContainerName: "container",
									Command:       []string{"cmd"},
									Volumes:       []string{"good", ""},
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.pre.execNewPod.volumes[1]",
		},
		"missing spec.strategy.recreateParams.mid.execNewPod": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Mid: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.mid",
		},
		"missing spec.strategy.recreateParams.post.execNewPod": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Post: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.post",
		},
		"missing spec.strategy.after.tagImages": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Post: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								TagImages: []deployapi.TagImageHook{
									{
										ContainerName: "missing",
										To:            kapi.ObjectReference{Kind: "ImageStreamTag", Name: "stream:tag"},
									},
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.post.tagImages[0].containerName",
		},
		"missing spec.strategy.after.tagImages.to.kind": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Post: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								TagImages: []deployapi.TagImageHook{
									{
										ContainerName: "container1",
										To:            kapi.ObjectReference{Name: "stream:tag"},
									},
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.post.tagImages[0].to.kind",
		},
		"missing spec.strategy.after.tagImages.to.name": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Post: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								TagImages: []deployapi.TagImageHook{
									{
										ContainerName: "container1",
										To:            kapi.ObjectReference{Kind: "ImageStreamTag"},
									},
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.recreateParams.post.tagImages[0].to.name",
		},
		"can't have both tag and execNewPod": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRecreate,
						RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
							Post: &deployapi.LifecycleHook{
								FailurePolicy: deployapi.LifecycleHookFailurePolicyRetry,
								ExecNewPod:    &deployapi.ExecNewPodHook{},
								TagImages:     []deployapi.TagImageHook{{}},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeInvalid,
			"spec.strategy.recreateParams.post",
		},
		"invalid spec.strategy.rollingParams.intervalSeconds": {
			rollingConfig(-20, 1, 1),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.intervalSeconds",
		},
		"invalid spec.strategy.rollingParams.updatePeriodSeconds": {
			rollingConfig(1, -20, 1),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.updatePeriodSeconds",
		},
		"invalid spec.strategy.rollingParams.timeoutSeconds": {
			rollingConfig(1, 1, -20),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.timeoutSeconds",
		},
		"missing spec.strategy.rollingParams.pre.failurePolicy": {
			deployapi.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
				Spec: deployapi.DeploymentConfigSpec{
					Replicas: 1,
					Strategy: deployapi.DeploymentStrategy{
						Type: deployapi.DeploymentStrategyTypeRolling,
						RollingParams: &deployapi.RollingDeploymentStrategyParams{
							IntervalSeconds:     mkint64p(1),
							UpdatePeriodSeconds: mkint64p(1),
							TimeoutSeconds:      mkint64p(20),
							MaxSurge:            intstr.FromInt(1),
							Pre: &deployapi.LifecycleHook{
								ExecNewPod: &deployapi.ExecNewPodHook{
									Command:       []string{"cmd"},
									ContainerName: "container",
								},
							},
						},
						ActiveDeadlineSeconds: mkint64p(3600),
					},
					Template: test.OkPodTemplate(),
					Selector: test.OkSelector(),
				},
			},
			field.ErrorTypeRequired,
			"spec.strategy.rollingParams.pre.failurePolicy",
		},
		"both maxSurge and maxUnavailable 0 spec.strategy.rollingParams.maxUnavailable": {
			rollingConfigMax(intstr.FromInt(0), intstr.FromInt(0)),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxUnavailable",
		},
		"invalid lower bound spec.strategy.rollingParams.maxUnavailable": {
			rollingConfigMax(intstr.FromInt(0), intstr.FromInt(-100)),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxUnavailable",
		},
		"invalid lower bound spec.strategy.rollingParams.maxSurge": {
			rollingConfigMax(intstr.FromInt(-1), intstr.FromInt(0)),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxSurge",
		},
		"both maxSurge and maxUnavailable 0 percent spec.strategy.rollingParams.maxUnavailable": {
			rollingConfigMax(intstr.FromString("0%"), intstr.FromString("0%")),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxUnavailable",
		},
		"invalid lower bound percent spec.strategy.rollingParams.maxUnavailable": {
			rollingConfigMax(intstr.FromInt(0), intstr.FromString("-1%")),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxUnavailable",
		},
		"invalid upper bound percent spec.strategy.rollingParams.maxUnavailable": {
			rollingConfigMax(intstr.FromInt(0), intstr.FromString("101%")),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxUnavailable",
		},
		"invalid percent spec.strategy.rollingParams.maxUnavailable": {
			rollingConfigMax(intstr.FromInt(0), intstr.FromString("foo")),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxUnavailable",
		},
		"invalid percent spec.strategy.rollingParams.maxSurge": {
			rollingConfigMax(intstr.FromString("foo"), intstr.FromString("100%")),
			field.ErrorTypeInvalid,
			"spec.strategy.rollingParams.maxSurge",
		},
	}

	for testName, v := range errorCases {
		t.Logf("running scenario %q", testName)
		errs := ValidateDeploymentConfig(&v.DeploymentConfig)
		if len(v.ErrorType) == 0 {
			if len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("%s: unexpected error: %s", testName, e)
				}
			}
			continue
		}
		if len(errs) == 0 {
			t.Errorf("%s: expected test failure, got success", testName)
		}
		for i := range errs {
			if got, exp := errs[i].Type, v.ErrorType; got != exp {
				t.Errorf("%s: expected error \"%v\" to have type %q, but got %q", testName, errs[i], exp, got)
			}
			if got, exp := errs[i].Field, v.Field; got != exp {
				t.Errorf("%s: expected error \"%v\" to have field %q, but got %q", testName, errs[i], exp, got)
			}
		}
	}
}

func TestValidateDeploymentConfigUpdate(t *testing.T) {
	oldConfig := &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar", ResourceVersion: "1"},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: manualTrigger(),
			Selector: test.OkSelector(),
			Strategy: test.OkStrategy(),
			Template: test.OkPodTemplate(),
		},
		Status: deployapi.DeploymentConfigStatus{
			LatestVersion: 5,
		},
	}
	newConfig := &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar", ResourceVersion: "1"},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: manualTrigger(),
			Selector: test.OkSelector(),
			Strategy: test.OkStrategy(),
			Template: test.OkPodTemplate(),
		},
		Status: deployapi.DeploymentConfigStatus{
			LatestVersion: 3,
		},
	}

	scenarios := []struct {
		oldLatestVersion int64
		newLatestVersion int64
	}{
		{5, 3},
		{5, 7},
		{0, -1},
	}

	for _, values := range scenarios {
		oldConfig.Status.LatestVersion = values.oldLatestVersion
		newConfig.Status.LatestVersion = values.newLatestVersion
		errs := ValidateDeploymentConfigUpdate(newConfig, oldConfig)
		if len(errs) == 0 {
			t.Errorf("Expected update failure")
		}
		for i := range errs {
			if errs[i].Type != field.ErrorTypeInvalid {
				t.Errorf("expected update error to have type %s: %v", field.ErrorTypeInvalid, errs[i])
			}
			if errs[i].Field != "status.latestVersion" {
				t.Errorf("expected update error to have field %s: %v", "latestVersion", errs[i])
			}
		}
	}

	// testing for a successful update
	oldConfig.Status.LatestVersion = 5
	newConfig.Status.LatestVersion = 6
	errs := ValidateDeploymentConfigUpdate(newConfig, oldConfig)
	if len(errs) > 0 {
		t.Errorf("Unexpected update failure")
	}
}

func TestValidateDeploymentConfigRollbackOK(t *testing.T) {
	rollback := &deployapi.DeploymentConfigRollback{
		Name: "config",
		Spec: deployapi.DeploymentConfigRollbackSpec{
			Revision: 2,
		},
	}

	errs := ValidateDeploymentConfigRollback(rollback)
	if len(errs) > 0 {
		t.Errorf("Unxpected non-empty error list: %v", errs)
	}
}

func TestValidateDeploymentConfigRollbackDeprecatedOK(t *testing.T) {
	rollback := &deployapi.DeploymentConfigRollback{
		Spec: deployapi.DeploymentConfigRollbackSpec{
			From: kapi.ObjectReference{
				Name: "deployment",
			},
		},
	}

	errs := ValidateDeploymentConfigRollbackDeprecated(rollback)
	if len(errs) > 0 {
		t.Errorf("Unxpected non-empty error list: %v", errs)
	}

	if e, a := "ReplicationController", rollback.Spec.From.Kind; e != a {
		t.Errorf("expected kind %s, got %s", e, a)
	}
}

func TestValidateDeploymentConfigRollbackInvalidFields(t *testing.T) {
	errorCases := map[string]struct {
		D deployapi.DeploymentConfigRollback
		T field.ErrorType
		F string
	}{
		"missing name": {
			deployapi.DeploymentConfigRollback{
				Spec: deployapi.DeploymentConfigRollbackSpec{
					Revision: 2,
				},
			},
			field.ErrorTypeRequired,
			"name",
		},
		"invalid name": {
			deployapi.DeploymentConfigRollback{
				Name: "*_*myconfig",
				Spec: deployapi.DeploymentConfigRollbackSpec{
					Revision: 2,
				},
			},
			field.ErrorTypeInvalid,
			"name",
		},
		"invalid revision": {
			deployapi.DeploymentConfigRollback{
				Name: "config",
				Spec: deployapi.DeploymentConfigRollbackSpec{
					Revision: -1,
				},
			},
			field.ErrorTypeInvalid,
			"spec.revision",
		},
	}

	for k, v := range errorCases {
		errs := ValidateDeploymentConfigRollback(&v.D)
		if len(errs) == 0 {
			t.Errorf("Expected failure for scenario %q", k)
		}
		for i := range errs {
			if errs[i].Type != v.T {
				t.Errorf("%s: expected errors to have type %q: %v", k, v.T, errs[i])
			}
			if errs[i].Field != v.F {
				t.Errorf("%s: expected errors to have field %q: %v", k, v.F, errs[i])
			}
		}
	}
}

func TestValidateDeploymentConfigRollbackDeprecatedInvalidFields(t *testing.T) {
	errorCases := map[string]struct {
		D deployapi.DeploymentConfigRollback
		T field.ErrorType
		F string
	}{
		"missing spec.from.name": {
			deployapi.DeploymentConfigRollback{
				Spec: deployapi.DeploymentConfigRollbackSpec{
					From: kapi.ObjectReference{},
				},
			},
			field.ErrorTypeRequired,
			"spec.from.name",
		},
		"wrong spec.from.kind": {
			deployapi.DeploymentConfigRollback{
				Spec: deployapi.DeploymentConfigRollbackSpec{
					From: kapi.ObjectReference{
						Kind: "unknown",
						Name: "deployment",
					},
				},
			},
			field.ErrorTypeInvalid,
			"spec.from.kind",
		},
	}

	for k, v := range errorCases {
		errs := ValidateDeploymentConfigRollbackDeprecated(&v.D)
		if len(errs) == 0 {
			t.Errorf("Expected failure for scenario %s", k)
		}
		for i := range errs {
			if errs[i].Type != v.T {
				t.Errorf("%s: expected errors to have type %s: %v", k, v.T, errs[i])
			}
			if errs[i].Field != v.F {
				t.Errorf("%s: expected errors to have field %s: %v", k, v.F, errs[i])
			}
		}
	}
}

func TestValidateDeploymentConfigDefaultImageStreamKind(t *testing.T) {
	config := &deployapi.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "bar"},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: []deployapi.DeploymentTriggerPolicy{
				{
					Type: deployapi.DeploymentTriggerOnImageChange,
					ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
						From: kapi.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "name:v1",
						},
						ContainerNames: []string{"foo"},
					},
				},
			},
			Selector: test.OkSelector(),
			Template: test.OkPodTemplate(),
			Strategy: test.OkStrategy(),
		},
	}

	if errs := ValidateDeploymentConfig(config); len(errs) > 0 {
		t.Errorf("Unxpected non-empty error list: %v", errs)
	}
}

func mkint64p(i int) *int64 {
	v := int64(i)
	return &v
}

func mkintp(i int) *int {
	return &i
}

func TestValidateSelectorMatchesPodTemplateLabels(t *testing.T) {
	tests := map[string]struct {
		spec        deployapi.DeploymentConfigSpec
		expectedErr bool
		errorType   field.ErrorType
		field       string
	}{
		"valid template labels": {
			spec: deployapi.DeploymentConfigSpec{
				Selector: test.OkSelector(),
				Strategy: test.OkStrategy(),
				Template: test.OkPodTemplate(),
			},
		},
		"invalid template labels": {
			spec: deployapi.DeploymentConfigSpec{
				Selector: test.OkSelector(),
				Strategy: test.OkStrategy(),
				Template: test.OkPodTemplate(),
			},
			expectedErr: true,
			errorType:   field.ErrorTypeInvalid,
			field:       "spec.template.metadata.labels",
		},
	}

	for name, test := range tests {
		if test.expectedErr {
			test.spec.Template.Labels["a"] = "c"
		}
		errs := ValidateDeploymentConfigSpec(test.spec)
		if len(errs) == 0 && test.expectedErr {
			t.Errorf("%s: expected failure", name)
			continue
		}
		if !test.expectedErr {
			continue
		}
		if len(errs) != 1 {
			t.Errorf("%s: expected one error, got %d", name, len(errs))
			continue
		}
		err := errs[0]
		if err.Type != test.errorType {
			t.Errorf("%s: expected error to have type %q, got %q", name, test.errorType, err.Type)
		}
		if err.Field != test.field {
			t.Errorf("%s: expected error to have field %q, got %q", name, test.field, err.Field)
		}
	}
}
