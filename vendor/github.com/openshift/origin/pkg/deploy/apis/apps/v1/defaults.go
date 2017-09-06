package v1

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
)

// Applies defaults only for API group "apps.openshift.io" and not for the legacy API.
// This function is called from storage layer where differentiation
// between legacy and group API can be made and is not related to other functions here
// which are called fom auto-generated code.
func AppsV1DeploymentConfigLayeredDefaults(dc *deployapi.DeploymentConfig) {
	if dc.Spec.RevisionHistoryLimit == nil {
		v := deployapi.DefaultRevisionHistoryLimit
		dc.Spec.RevisionHistoryLimit = &v
	}
}

// Keep this in sync with pkg/api/serialization_test.go#defaultHookContainerName
func defaultHookContainerName(hook *LifecycleHook, containerName string) {
	if hook == nil {
		return
	}
	for i := range hook.TagImages {
		if len(hook.TagImages[i].ContainerName) == 0 {
			hook.TagImages[i].ContainerName = containerName
		}
	}
	if hook.ExecNewPod != nil {
		if len(hook.ExecNewPod.ContainerName) == 0 {
			hook.ExecNewPod.ContainerName = containerName
		}
	}
}

func SetDefaults_DeploymentConfigSpec(obj *DeploymentConfigSpec) {
	if obj.Triggers == nil {
		obj.Triggers = []DeploymentTriggerPolicy{
			{Type: DeploymentTriggerOnConfigChange},
		}
	}
	if len(obj.Selector) == 0 && obj.Template != nil {
		obj.Selector = obj.Template.Labels
	}

	// if you only specify a single container, default the TagImages hook to the container name
	if obj.Template != nil && len(obj.Template.Spec.Containers) == 1 {
		containerName := obj.Template.Spec.Containers[0].Name
		if p := obj.Strategy.RecreateParams; p != nil {
			defaultHookContainerName(p.Pre, containerName)
			defaultHookContainerName(p.Mid, containerName)
			defaultHookContainerName(p.Post, containerName)
		}
		if p := obj.Strategy.RollingParams; p != nil {
			defaultHookContainerName(p.Pre, containerName)
			defaultHookContainerName(p.Post, containerName)
		}
	}
}

func SetDefaults_DeploymentStrategy(obj *DeploymentStrategy) {
	if len(obj.Type) == 0 {
		obj.Type = DeploymentStrategyTypeRolling
	}

	if obj.Type == DeploymentStrategyTypeRolling && obj.RollingParams == nil {
		obj.RollingParams = &RollingDeploymentStrategyParams{
			IntervalSeconds:     mkintp(deployapi.DefaultRollingIntervalSeconds),
			UpdatePeriodSeconds: mkintp(deployapi.DefaultRollingUpdatePeriodSeconds),
			TimeoutSeconds:      mkintp(deployapi.DefaultRollingTimeoutSeconds),
		}
	}
	if obj.Type == DeploymentStrategyTypeRecreate && obj.RecreateParams == nil {
		obj.RecreateParams = &RecreateDeploymentStrategyParams{}
	}

	if obj.ActiveDeadlineSeconds == nil {
		obj.ActiveDeadlineSeconds = mkintp(deployapi.MaxDeploymentDurationSeconds)
	}
}

func SetDefaults_RecreateDeploymentStrategyParams(obj *RecreateDeploymentStrategyParams) {
	if obj.TimeoutSeconds == nil {
		obj.TimeoutSeconds = mkintp(deployapi.DefaultRecreateTimeoutSeconds)
	}
}

func SetDefaults_RollingDeploymentStrategyParams(obj *RollingDeploymentStrategyParams) {
	if obj.IntervalSeconds == nil {
		obj.IntervalSeconds = mkintp(deployapi.DefaultRollingIntervalSeconds)
	}

	if obj.UpdatePeriodSeconds == nil {
		obj.UpdatePeriodSeconds = mkintp(deployapi.DefaultRollingUpdatePeriodSeconds)
	}

	if obj.TimeoutSeconds == nil {
		obj.TimeoutSeconds = mkintp(deployapi.DefaultRollingTimeoutSeconds)
	}

	if obj.MaxUnavailable == nil && obj.MaxSurge == nil {
		maxUnavailable := intstr.FromString("25%")
		obj.MaxUnavailable = &maxUnavailable

		maxSurge := intstr.FromString("25%")
		obj.MaxSurge = &maxSurge
	}

	if obj.MaxUnavailable == nil && obj.MaxSurge != nil &&
		(*obj.MaxSurge == intstr.FromInt(0) || *obj.MaxSurge == intstr.FromString("0%")) {
		maxUnavailable := intstr.FromString("25%")
		obj.MaxUnavailable = &maxUnavailable
	}

	if obj.MaxSurge == nil && obj.MaxUnavailable != nil &&
		(*obj.MaxUnavailable == intstr.FromInt(0) || *obj.MaxUnavailable == intstr.FromString("0%")) {
		maxSurge := intstr.FromString("25%")
		obj.MaxSurge = &maxSurge
	}
}

func SetDefaults_DeploymentConfig(obj *DeploymentConfig) {
	for _, t := range obj.Spec.Triggers {
		if t.ImageChangeParams != nil {
			if len(t.ImageChangeParams.From.Kind) == 0 {
				t.ImageChangeParams.From.Kind = "ImageStreamTag"
			}
			if len(t.ImageChangeParams.From.Namespace) == 0 {
				t.ImageChangeParams.From.Namespace = obj.Namespace
			}
		}
	}
}

func mkintp(i int64) *int64 {
	return &i
}
