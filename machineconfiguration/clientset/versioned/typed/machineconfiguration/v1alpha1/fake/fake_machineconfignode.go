// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "github.com/openshift/api/machineconfiguration/v1alpha1"
	machineconfigurationv1alpha1 "github.com/openshift/client-go/machineconfiguration/applyconfigurations/machineconfiguration/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMachineConfigNodes implements MachineConfigNodeInterface
type FakeMachineConfigNodes struct {
	Fake *FakeMachineconfigurationV1alpha1
}

var machineconfignodesResource = v1alpha1.SchemeGroupVersion.WithResource("machineconfignodes")

var machineconfignodesKind = v1alpha1.SchemeGroupVersion.WithKind("MachineConfigNode")

// Get takes name of the machineConfigNode, and returns the corresponding machineConfigNode object, and an error if there is any.
func (c *FakeMachineConfigNodes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MachineConfigNode, err error) {
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(machineconfignodesResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}

// List takes label and field selectors, and returns the list of MachineConfigNodes that match those selectors.
func (c *FakeMachineConfigNodes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MachineConfigNodeList, err error) {
	emptyResult := &v1alpha1.MachineConfigNodeList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(machineconfignodesResource, machineconfignodesKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MachineConfigNodeList{ListMeta: obj.(*v1alpha1.MachineConfigNodeList).ListMeta}
	for _, item := range obj.(*v1alpha1.MachineConfigNodeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machineConfigNodes.
func (c *FakeMachineConfigNodes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(machineconfignodesResource, opts))
}

// Create takes the representation of a machineConfigNode and creates it.  Returns the server's representation of the machineConfigNode, and an error, if there is any.
func (c *FakeMachineConfigNodes) Create(ctx context.Context, machineConfigNode *v1alpha1.MachineConfigNode, opts v1.CreateOptions) (result *v1alpha1.MachineConfigNode, err error) {
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(machineconfignodesResource, machineConfigNode, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}

// Update takes the representation of a machineConfigNode and updates it. Returns the server's representation of the machineConfigNode, and an error, if there is any.
func (c *FakeMachineConfigNodes) Update(ctx context.Context, machineConfigNode *v1alpha1.MachineConfigNode, opts v1.UpdateOptions) (result *v1alpha1.MachineConfigNode, err error) {
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(machineconfignodesResource, machineConfigNode, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMachineConfigNodes) UpdateStatus(ctx context.Context, machineConfigNode *v1alpha1.MachineConfigNode, opts v1.UpdateOptions) (result *v1alpha1.MachineConfigNode, err error) {
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceActionWithOptions(machineconfignodesResource, "status", machineConfigNode, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}

// Delete takes name of the machineConfigNode and deletes it. Returns an error if one occurs.
func (c *FakeMachineConfigNodes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(machineconfignodesResource, name, opts), &v1alpha1.MachineConfigNode{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachineConfigNodes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(machineconfignodesResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MachineConfigNodeList{})
	return err
}

// Patch applies the patch and returns the patched machineConfigNode.
func (c *FakeMachineConfigNodes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MachineConfigNode, err error) {
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(machineconfignodesResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied machineConfigNode.
func (c *FakeMachineConfigNodes) Apply(ctx context.Context, machineConfigNode *machineconfigurationv1alpha1.MachineConfigNodeApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.MachineConfigNode, err error) {
	if machineConfigNode == nil {
		return nil, fmt.Errorf("machineConfigNode provided to Apply must not be nil")
	}
	data, err := json.Marshal(machineConfigNode)
	if err != nil {
		return nil, err
	}
	name := machineConfigNode.Name
	if name == nil {
		return nil, fmt.Errorf("machineConfigNode.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(machineconfignodesResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeMachineConfigNodes) ApplyStatus(ctx context.Context, machineConfigNode *machineconfigurationv1alpha1.MachineConfigNodeApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.MachineConfigNode, err error) {
	if machineConfigNode == nil {
		return nil, fmt.Errorf("machineConfigNode provided to Apply must not be nil")
	}
	data, err := json.Marshal(machineConfigNode)
	if err != nil {
		return nil, err
	}
	name := machineConfigNode.Name
	if name == nil {
		return nil, fmt.Errorf("machineConfigNode.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.MachineConfigNode{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(machineconfignodesResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.MachineConfigNode), err
}
