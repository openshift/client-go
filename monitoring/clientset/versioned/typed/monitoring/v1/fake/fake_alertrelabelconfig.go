// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1 "github.com/openshift/api/monitoring/v1"
	monitoringv1 "github.com/openshift/client-go/monitoring/applyconfigurations/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAlertRelabelConfigs implements AlertRelabelConfigInterface
type FakeAlertRelabelConfigs struct {
	Fake *FakeMonitoringV1
	ns   string
}

var alertrelabelconfigsResource = v1.SchemeGroupVersion.WithResource("alertrelabelconfigs")

var alertrelabelconfigsKind = v1.SchemeGroupVersion.WithKind("AlertRelabelConfig")

// Get takes name of the alertRelabelConfig, and returns the corresponding alertRelabelConfig object, and an error if there is any.
func (c *FakeAlertRelabelConfigs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.AlertRelabelConfig, err error) {
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(alertrelabelconfigsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}

// List takes label and field selectors, and returns the list of AlertRelabelConfigs that match those selectors.
func (c *FakeAlertRelabelConfigs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.AlertRelabelConfigList, err error) {
	emptyResult := &v1.AlertRelabelConfigList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(alertrelabelconfigsResource, alertrelabelconfigsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.AlertRelabelConfigList{ListMeta: obj.(*v1.AlertRelabelConfigList).ListMeta}
	for _, item := range obj.(*v1.AlertRelabelConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested alertRelabelConfigs.
func (c *FakeAlertRelabelConfigs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(alertrelabelconfigsResource, c.ns, opts))

}

// Create takes the representation of a alertRelabelConfig and creates it.  Returns the server's representation of the alertRelabelConfig, and an error, if there is any.
func (c *FakeAlertRelabelConfigs) Create(ctx context.Context, alertRelabelConfig *v1.AlertRelabelConfig, opts metav1.CreateOptions) (result *v1.AlertRelabelConfig, err error) {
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(alertrelabelconfigsResource, c.ns, alertRelabelConfig, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}

// Update takes the representation of a alertRelabelConfig and updates it. Returns the server's representation of the alertRelabelConfig, and an error, if there is any.
func (c *FakeAlertRelabelConfigs) Update(ctx context.Context, alertRelabelConfig *v1.AlertRelabelConfig, opts metav1.UpdateOptions) (result *v1.AlertRelabelConfig, err error) {
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(alertrelabelconfigsResource, c.ns, alertRelabelConfig, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAlertRelabelConfigs) UpdateStatus(ctx context.Context, alertRelabelConfig *v1.AlertRelabelConfig, opts metav1.UpdateOptions) (result *v1.AlertRelabelConfig, err error) {
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(alertrelabelconfigsResource, "status", c.ns, alertRelabelConfig, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}

// Delete takes name of the alertRelabelConfig and deletes it. Returns an error if one occurs.
func (c *FakeAlertRelabelConfigs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(alertrelabelconfigsResource, c.ns, name, opts), &v1.AlertRelabelConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAlertRelabelConfigs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(alertrelabelconfigsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.AlertRelabelConfigList{})
	return err
}

// Patch applies the patch and returns the patched alertRelabelConfig.
func (c *FakeAlertRelabelConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AlertRelabelConfig, err error) {
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(alertrelabelconfigsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied alertRelabelConfig.
func (c *FakeAlertRelabelConfigs) Apply(ctx context.Context, alertRelabelConfig *monitoringv1.AlertRelabelConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.AlertRelabelConfig, err error) {
	if alertRelabelConfig == nil {
		return nil, fmt.Errorf("alertRelabelConfig provided to Apply must not be nil")
	}
	data, err := json.Marshal(alertRelabelConfig)
	if err != nil {
		return nil, err
	}
	name := alertRelabelConfig.Name
	if name == nil {
		return nil, fmt.Errorf("alertRelabelConfig.Name must be provided to Apply")
	}
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(alertrelabelconfigsResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeAlertRelabelConfigs) ApplyStatus(ctx context.Context, alertRelabelConfig *monitoringv1.AlertRelabelConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.AlertRelabelConfig, err error) {
	if alertRelabelConfig == nil {
		return nil, fmt.Errorf("alertRelabelConfig provided to Apply must not be nil")
	}
	data, err := json.Marshal(alertRelabelConfig)
	if err != nil {
		return nil, err
	}
	name := alertRelabelConfig.Name
	if name == nil {
		return nil, fmt.Errorf("alertRelabelConfig.Name must be provided to Apply")
	}
	emptyResult := &v1.AlertRelabelConfig{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(alertrelabelconfigsResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.AlertRelabelConfig), err
}
