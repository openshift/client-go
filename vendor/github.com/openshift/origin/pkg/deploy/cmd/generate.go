package cmd

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubectl"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps/v1"
)

var basic = kubectl.BasicReplicationController{}

type BasicDeploymentConfigController struct{}

func (BasicDeploymentConfigController) ParamNames() []kubectl.GeneratorParam {
	return basic.ParamNames()
}

func (BasicDeploymentConfigController) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	obj, err := basic.Generate(genericParams)
	if err != nil {
		return nil, err
	}
	switch t := obj.(type) {
	case *kapi.ReplicationController:
		obj = &deployapi.DeploymentConfig{
			ObjectMeta: t.ObjectMeta,
			Spec: deployapi.DeploymentConfigSpec{
				Selector: t.Spec.Selector,
				Replicas: *t.Spec.Replicas, // the generator never leaves this nil
				Template: t.Spec.Template,
			},
		}
	default:
		return nil, fmt.Errorf("unrecognized object type: %v", reflect.TypeOf(t))
	}
	return obj, nil
}
