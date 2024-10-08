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

// FakeContainerRuntimeConfigs implements ContainerRuntimeConfigInterface
type FakeContainerRuntimeConfigs struct {
	Fake *FakeMachineconfigurationV1
}

var containerruntimeconfigsResource = v1.SchemeGroupVersion.WithResource("containerruntimeconfigs")

var containerruntimeconfigsKind = v1.SchemeGroupVersion.WithKind("ContainerRuntimeConfig")

// Get takes name of the containerRuntimeConfig, and returns the corresponding containerRuntimeConfig object, and an error if there is any.
func (c *FakeContainerRuntimeConfigs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ContainerRuntimeConfig, err error) {
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(containerruntimeconfigsResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}

// List takes label and field selectors, and returns the list of ContainerRuntimeConfigs that match those selectors.
func (c *FakeContainerRuntimeConfigs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ContainerRuntimeConfigList, err error) {
	emptyResult := &v1.ContainerRuntimeConfigList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(containerruntimeconfigsResource, containerruntimeconfigsKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ContainerRuntimeConfigList{ListMeta: obj.(*v1.ContainerRuntimeConfigList).ListMeta}
	for _, item := range obj.(*v1.ContainerRuntimeConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested containerRuntimeConfigs.
func (c *FakeContainerRuntimeConfigs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(containerruntimeconfigsResource, opts))
}

// Create takes the representation of a containerRuntimeConfig and creates it.  Returns the server's representation of the containerRuntimeConfig, and an error, if there is any.
func (c *FakeContainerRuntimeConfigs) Create(ctx context.Context, containerRuntimeConfig *v1.ContainerRuntimeConfig, opts metav1.CreateOptions) (result *v1.ContainerRuntimeConfig, err error) {
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(containerruntimeconfigsResource, containerRuntimeConfig, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}

// Update takes the representation of a containerRuntimeConfig and updates it. Returns the server's representation of the containerRuntimeConfig, and an error, if there is any.
func (c *FakeContainerRuntimeConfigs) Update(ctx context.Context, containerRuntimeConfig *v1.ContainerRuntimeConfig, opts metav1.UpdateOptions) (result *v1.ContainerRuntimeConfig, err error) {
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(containerruntimeconfigsResource, containerRuntimeConfig, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeContainerRuntimeConfigs) UpdateStatus(ctx context.Context, containerRuntimeConfig *v1.ContainerRuntimeConfig, opts metav1.UpdateOptions) (result *v1.ContainerRuntimeConfig, err error) {
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceActionWithOptions(containerruntimeconfigsResource, "status", containerRuntimeConfig, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}

// Delete takes name of the containerRuntimeConfig and deletes it. Returns an error if one occurs.
func (c *FakeContainerRuntimeConfigs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(containerruntimeconfigsResource, name, opts), &v1.ContainerRuntimeConfig{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeContainerRuntimeConfigs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(containerruntimeconfigsResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ContainerRuntimeConfigList{})
	return err
}

// Patch applies the patch and returns the patched containerRuntimeConfig.
func (c *FakeContainerRuntimeConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ContainerRuntimeConfig, err error) {
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(containerruntimeconfigsResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied containerRuntimeConfig.
func (c *FakeContainerRuntimeConfigs) Apply(ctx context.Context, containerRuntimeConfig *machineconfigurationv1.ContainerRuntimeConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ContainerRuntimeConfig, err error) {
	if containerRuntimeConfig == nil {
		return nil, fmt.Errorf("containerRuntimeConfig provided to Apply must not be nil")
	}
	data, err := json.Marshal(containerRuntimeConfig)
	if err != nil {
		return nil, err
	}
	name := containerRuntimeConfig.Name
	if name == nil {
		return nil, fmt.Errorf("containerRuntimeConfig.Name must be provided to Apply")
	}
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(containerruntimeconfigsResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeContainerRuntimeConfigs) ApplyStatus(ctx context.Context, containerRuntimeConfig *machineconfigurationv1.ContainerRuntimeConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ContainerRuntimeConfig, err error) {
	if containerRuntimeConfig == nil {
		return nil, fmt.Errorf("containerRuntimeConfig provided to Apply must not be nil")
	}
	data, err := json.Marshal(containerRuntimeConfig)
	if err != nil {
		return nil, err
	}
	name := containerRuntimeConfig.Name
	if name == nil {
		return nil, fmt.Errorf("containerRuntimeConfig.Name must be provided to Apply")
	}
	emptyResult := &v1.ContainerRuntimeConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(containerruntimeconfigsResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ContainerRuntimeConfig), err
}
