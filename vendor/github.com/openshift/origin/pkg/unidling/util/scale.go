package util

import (
	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	kapi "k8s.io/kubernetes/pkg/api"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	kextapi "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kextensionsclient "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/extensions/v1beta1"
	kcoreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	deployapiv1 "github.com/openshift/origin/pkg/deploy/apis/apps/v1"
	deployclient "github.com/openshift/origin/pkg/deploy/generated/internalclientset/typed/apps/internalversion"
	unidlingapi "github.com/openshift/origin/pkg/unidling/api"
)

// TODO: remove the below functions once we get a way to mark/unmark an object as idled
// via the scale endpoint

type AnnotationFunc func(currentReplicas int32, annotations map[string]string)

func NewScaleAnnotater(scales kextensionsclient.ScalesGetter, dcs deployclient.DeploymentConfigsGetter, rcs kcoreclient.ReplicationControllersGetter, changeAnnots AnnotationFunc) *ScaleAnnotater {
	return &ScaleAnnotater{
		scales:            scales,
		dcs:               dcs,
		rcs:               rcs,
		ChangeAnnotations: changeAnnots,
	}
}

type ScaleAnnotater struct {
	scales            kextensionsclient.ScalesGetter
	dcs               deployclient.DeploymentConfigsGetter
	rcs               kcoreclient.ReplicationControllersGetter
	ChangeAnnotations AnnotationFunc
}

// ScaleUpdater implements a method "Update" that knows how to update a given object
type ScaleUpdater interface {
	Update(*ScaleAnnotater, runtime.Object, *kextapi.Scale) error
}

// ScaleUpdater implements unidlingutil.ScaleUpdater
type scaleUpdater struct {
	encoder   runtime.Encoder
	namespace string
	dcGetter  deployclient.DeploymentConfigsGetter
	rcGetter  kcoreclient.ReplicationControllersGetter
}

func NewScaleUpdater(encoder runtime.Encoder, namespace string, dcGetter deployclient.DeploymentConfigsGetter, rcGetter kcoreclient.ReplicationControllersGetter) ScaleUpdater {
	return scaleUpdater{
		encoder:   encoder,
		namespace: namespace,
		dcGetter:  dcGetter,
		rcGetter:  rcGetter,
	}
}

func (s scaleUpdater) Update(annotator *ScaleAnnotater, obj runtime.Object, scale *kextapi.Scale) error {
	var (
		err                             error
		patchBytes, originalObj, newObj []byte
	)

	originalObj, err = runtime.Encode(s.encoder, obj)
	if err != nil {
		return err
	}

	switch typedObj := obj.(type) {
	case *deployapi.DeploymentConfig:
		if typedObj.Annotations == nil {
			typedObj.Annotations = make(map[string]string)
		}

		annotator.ChangeAnnotations(typedObj.Spec.Replicas, typedObj.Annotations)
		typedObj.Spec.Replicas = scale.Spec.Replicas

		newObj, err = runtime.Encode(s.encoder, typedObj)
		if err != nil {
			return err
		}

		patchBytes, err = strategicpatch.CreateTwoWayMergePatch(originalObj, newObj, &deployapiv1.DeploymentConfig{})
		if err != nil {
			return err
		}

		_, err = s.dcGetter.DeploymentConfigs(s.namespace).Patch(typedObj.Name, types.StrategicMergePatchType, patchBytes)
	case *kapi.ReplicationController:
		if typedObj.Annotations == nil {
			typedObj.Annotations = make(map[string]string)
		}

		annotator.ChangeAnnotations(typedObj.Spec.Replicas, typedObj.Annotations)
		typedObj.Spec.Replicas = scale.Spec.Replicas

		newObj, err = runtime.Encode(s.encoder, typedObj)
		if err != nil {
			return err
		}

		patchBytes, err = strategicpatch.CreateTwoWayMergePatch(originalObj, newObj, &kapiv1.ReplicationController{})
		if err != nil {
			return err
		}

		_, err = s.rcGetter.ReplicationControllers(s.namespace).Patch(typedObj.Name, types.StrategicMergePatchType, patchBytes)
	}
	return err
}

// getObjectWithScale either fetches a known type of object and constructs a Scale from that, or uses the scale
// subresource to fetch a Scale by itself.
func (c *ScaleAnnotater) GetObjectWithScale(namespace string, ref unidlingapi.CrossGroupObjectReference) (runtime.Object, *kextapi.Scale, error) {
	var obj runtime.Object
	var err error
	var scale *kextapi.Scale

	switch {
	case ref.Kind == "DeploymentConfig" && (ref.Group == deployapi.GroupName || ref.Group == deployapi.LegacyGroupName):
		var dc *deployapi.DeploymentConfig
		dc, err = c.dcs.DeploymentConfigs(namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		scale = &kextapi.Scale{
			Spec: kextapi.ScaleSpec{Replicas: dc.Spec.Replicas},
		}
		obj = dc
	case ref.Kind == "ReplicationController" && ref.Group == kapi.GroupName:
		var rc *kapi.ReplicationController
		rc, err = c.rcs.ReplicationControllers(namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		scale = &kextapi.Scale{
			Spec: kextapi.ScaleSpec{Replicas: rc.Spec.Replicas},
		}
		obj = rc
	default:
		scale, err = c.scales.Scales(namespace).Get(ref.Kind, ref.Name)
		if err != nil {
			return nil, nil, err
		}
	}

	return obj, scale, err
}

// updateObjectScale updates the scale of an object and removes unidling annotations for objects of a know type.
// For objects of an unknown type, it scales the object using the scale subresource
// (and does not change annotations).
func (c *ScaleAnnotater) UpdateObjectScale(updater ScaleUpdater, namespace string, ref unidlingapi.CrossGroupObjectReference, obj runtime.Object, scale *kextapi.Scale) error {
	var err error

	if obj == nil {
		_, err = c.scales.Scales(namespace).Update(ref.Kind, scale)
		return err
	}

	switch obj.(type) {
	case *deployapi.DeploymentConfig, *kapi.ReplicationController:
		return updater.Update(c, obj, scale)
	default:
		glog.V(2).Infof("Unidling unknown type %t: using scale interface and not removing annotations")
		_, err = c.scales.Scales(namespace).Update(ref.Kind, scale)
	}

	return err
}
