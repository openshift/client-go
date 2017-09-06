package client

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kextensionsclient "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/extensions/v1beta1"

	"github.com/openshift/origin/pkg/api/latest"
)

type delegatingScaleInterface struct {
	dcs    DeploymentConfigInterface
	scales kextensionsclient.ScaleInterface
}

type delegatingScaleNamespacer struct {
	dcNS    DeploymentConfigsNamespacer
	scaleNS kextensionsclient.ScalesGetter
}

func (c *delegatingScaleNamespacer) Scales(namespace string) kextensionsclient.ScaleInterface {
	return &delegatingScaleInterface{
		dcs:    c.dcNS.DeploymentConfigs(namespace),
		scales: c.scaleNS.Scales(namespace),
	}
}

func NewDelegatingScaleNamespacer(dcNamespacer DeploymentConfigsNamespacer, sNamespacer kextensionsclient.ScalesGetter) kextensionsclient.ScalesGetter {
	return &delegatingScaleNamespacer{
		dcNS:    dcNamespacer,
		scaleNS: sNamespacer,
	}
}

// Get takes the reference to scale subresource and returns the subresource or error, if one occurs.
func (c *delegatingScaleInterface) Get(kind string, name string) (result *extensions.Scale, err error) {
	switch {
	case kind == "DeploymentConfig":
		return c.dcs.GetScale(name)
	// TODO: This is borked because the interface for Get is broken. Kind is insufficient.
	case latest.IsKindInAnyOriginGroup(kind):
		return nil, errors.NewBadRequest(fmt.Sprintf("Kind %s has no Scale subresource", kind))
	default:
		return c.scales.Get(kind, name)
	}
}

// Update takes a scale subresource object, updates the stored version to match it, and
// returns the subresource or error, if one occurs.
func (c *delegatingScaleInterface) Update(kind string, scale *extensions.Scale) (result *extensions.Scale, err error) {
	switch {
	case kind == "DeploymentConfig":
		return c.dcs.UpdateScale(scale)
	// TODO: This is borked because the interface for Update is broken. Kind is insufficient.
	case latest.IsKindInAnyOriginGroup(kind):
		return nil, errors.NewBadRequest(fmt.Sprintf("Kind %s has no Scale subresource", kind))
	default:
		return c.scales.Update(kind, scale)
	}
}
