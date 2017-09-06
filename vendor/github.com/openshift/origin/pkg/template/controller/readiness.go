package controller

import (
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"

	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	buildutil "github.com/openshift/origin/pkg/build/util"
	"github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
)

// checkBuildReadiness determins if a Build is ready, failed or neither.
func checkBuildReadiness(oc client.Interface, obj runtime.Object) (bool, bool, error) {
	b := obj.(*buildapi.Build)

	ready := buildutil.IsTerminalPhase(b.Status.Phase) &&
		b.Status.Phase == buildapi.BuildPhaseComplete

	failed := buildutil.IsTerminalPhase(b.Status.Phase) &&
		b.Status.Phase != buildapi.BuildPhaseComplete

	return ready, failed, nil
}

// checkBuildConfigReadiness determins if a BuildConfig is ready, failed or
// neither.  TODO: this should be reported on the BuildConfig object itself.
func checkBuildConfigReadiness(oc client.Interface, obj runtime.Object) (bool, bool, error) {
	bc := obj.(*buildapi.BuildConfig)

	builds, err := oc.Builds(bc.Namespace).List(metav1.ListOptions{LabelSelector: buildutil.BuildConfigSelector(bc.Name).String()})
	if err != nil {
		return false, false, err
	}

	for _, build := range builds.Items {
		if build.Annotations[buildapi.BuildNumberAnnotation] == strconv.FormatInt(bc.Status.LastVersion, 10) {
			return checkBuildReadiness(oc, &build)
		}
	}

	return false, false, nil
}

// checkDeploymentReadiness determins if a Deployment is ready, failed or
// neither.
func checkDeploymentReadiness(oc client.Interface, obj runtime.Object) (bool, bool, error) {
	d := obj.(*extensions.Deployment)

	var progressing, available *extensions.DeploymentCondition
	for i, condition := range d.Status.Conditions {
		switch condition.Type {
		case extensions.DeploymentProgressing:
			progressing = &d.Status.Conditions[i]

		case extensions.DeploymentAvailable:
			available = &d.Status.Conditions[i]
		}
	}

	ready := d.Status.ObservedGeneration == d.Generation &&
		progressing != nil &&
		progressing.Status == kapi.ConditionTrue &&
		progressing.Reason == deploymentutil.NewRSAvailableReason &&
		available != nil &&
		available.Status == kapi.ConditionTrue

	failed := d.Status.ObservedGeneration == d.Generation &&
		progressing != nil &&
		progressing.Status == kapi.ConditionFalse

	return ready, failed, nil
}

// checkDeploymentConfigReadiness determins if a DeploymentConfig is ready,
// failed or neither.
func checkDeploymentConfigReadiness(oc client.Interface, obj runtime.Object) (bool, bool, error) {
	dc := obj.(*deployapi.DeploymentConfig)

	var progressing, available *deployapi.DeploymentCondition
	for i, condition := range dc.Status.Conditions {
		switch condition.Type {
		case deployapi.DeploymentProgressing:
			progressing = &dc.Status.Conditions[i]

		case deployapi.DeploymentAvailable:
			available = &dc.Status.Conditions[i]
		}
	}

	ready := dc.Status.ObservedGeneration == dc.Generation &&
		progressing != nil &&
		progressing.Status == kapi.ConditionTrue &&
		progressing.Reason == deployapi.NewRcAvailableReason &&
		available != nil &&
		available.Status == kapi.ConditionTrue

	failed := dc.Status.ObservedGeneration == dc.Generation &&
		progressing != nil &&
		progressing.Status == kapi.ConditionFalse

	return ready, failed, nil
}

// checkJobReadiness determins if a Job is ready, failed or neither.
func checkJobReadiness(oc client.Interface, obj runtime.Object) (bool, bool, error) {
	job := obj.(*batch.Job)

	ready := job.Status.CompletionTime != nil
	failed := job.Status.Failed > 0

	return ready, failed, nil
}

// checkStatefulSetReadiness determins if a StatefulSet is ready, failed or
// neither.
func checkStatefulSetReadiness(oc client.Interface, obj runtime.Object) (bool, bool, error) {
	ss := obj.(*apps.StatefulSet)

	ready := ss.Status.ObservedGeneration != nil &&
		*ss.Status.ObservedGeneration == ss.Generation &&
		ss.Status.ReadyReplicas == ss.Spec.Replicas
	failed := false

	return ready, failed, nil
}

// readinessCheckers maps GroupKinds to the appropriate function.  Note that in
// some cases more than one GK maps to the same function.
var readinessCheckers = map[schema.GroupKind]func(client.Interface, runtime.Object) (bool, bool, error){
	buildapi.LegacyKind("Build"):             checkBuildReadiness,
	buildapi.Kind("Build"):                   checkBuildReadiness,
	buildapi.LegacyKind("BuildConfig"):       checkBuildConfigReadiness,
	buildapi.Kind("BuildConfig"):             checkBuildConfigReadiness,
	apps.Kind("Deployment"):                  checkDeploymentReadiness,
	extensions.Kind("Deployment"):            checkDeploymentReadiness,
	deployapi.LegacyKind("DeploymentConfig"): checkDeploymentConfigReadiness,
	deployapi.Kind("DeploymentConfig"):       checkDeploymentConfigReadiness,
	batch.Kind("Job"):                        checkJobReadiness,
	apps.Kind("StatefulSet"):                 checkStatefulSetReadiness,
}

// canCheckReadiness indicates whether a readiness check exists for a GK.
func canCheckReadiness(ref kapi.ObjectReference) bool {
	_, found := readinessCheckers[ref.GroupVersionKind().GroupKind()]
	return found
}

// checkReadiness runs the readiness check on a given object.  TODO: remove
// "oc client.Interface" and error once BuildConfigs can report on the status of
// their latest build.
func checkReadiness(oc client.Interface, ref kapi.ObjectReference, obj runtime.Object) (bool, bool, error) {
	return readinessCheckers[ref.GroupVersionKind().GroupKind()](oc, obj)
}
