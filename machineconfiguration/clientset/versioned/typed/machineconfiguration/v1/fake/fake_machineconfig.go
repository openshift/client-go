// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1 "github.com/openshift/api/machineconfiguration/v1"
	machineconfigurationv1 "github.com/openshift/client-go/machineconfiguration/applyconfigurations/machineconfiguration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMachineConfigs implements MachineConfigInterface
type FakeMachineConfigs struct {
	Fake *FakeMachineconfigurationV1
}

var machineconfigsResource = v1.SchemeGroupVersion.WithResource("machineconfigs")

var machineconfigsKind = v1.SchemeGroupVersion.WithKind("MachineConfig")

// Get takes name of the machineConfig, and returns the corresponding machineConfig object, and an error if there is any.
func (c *FakeMachineConfigs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.MachineConfig, err error) {
	emptyResult := &v1.MachineConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(machineconfigsResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineConfig), err
}

// List takes label and field selectors, and returns the list of MachineConfigs that match those selectors.
func (c *FakeMachineConfigs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.MachineConfigList, err error) {
	emptyResult := &v1.MachineConfigList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(machineconfigsResource, machineconfigsKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.MachineConfigList{ListMeta: obj.(*v1.MachineConfigList).ListMeta}
	for _, item := range obj.(*v1.MachineConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machineConfigs.
func (c *FakeMachineConfigs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(machineconfigsResource, opts))
}

// Create takes the representation of a machineConfig and creates it.  Returns the server's representation of the machineConfig, and an error, if there is any.
func (c *FakeMachineConfigs) Create(ctx context.Context, machineConfig *v1.MachineConfig, opts metav1.CreateOptions) (result *v1.MachineConfig, err error) {
	emptyResult := &v1.MachineConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(machineconfigsResource, machineConfig, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineConfig), err
}

// Update takes the representation of a machineConfig and updates it. Returns the server's representation of the machineConfig, and an error, if there is any.
func (c *FakeMachineConfigs) Update(ctx context.Context, machineConfig *v1.MachineConfig, opts metav1.UpdateOptions) (result *v1.MachineConfig, err error) {
	emptyResult := &v1.MachineConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(machineconfigsResource, machineConfig, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineConfig), err
}

// Delete takes name of the machineConfig and deletes it. Returns an error if one occurs.
func (c *FakeMachineConfigs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(machineconfigsResource, name, opts), &v1.MachineConfig{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachineConfigs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(machineconfigsResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.MachineConfigList{})
	return err
}

// Patch applies the patch and returns the patched machineConfig.
func (c *FakeMachineConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MachineConfig, err error) {
	emptyResult := &v1.MachineConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(machineconfigsResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineConfig), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied machineConfig.
func (c *FakeMachineConfigs) Apply(ctx context.Context, machineConfig *machineconfigurationv1.MachineConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.MachineConfig, err error) {
	if machineConfig == nil {
		return nil, fmt.Errorf("machineConfig provided to Apply must not be nil")
	}
	data, err := json.Marshal(machineConfig)
	if err != nil {
		return nil, err
	}
	name := machineConfig.Name
	if name == nil {
		return nil, fmt.Errorf("machineConfig.Name must be provided to Apply")
	}
	emptyResult := &v1.MachineConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(machineconfigsResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineConfig), err
}
