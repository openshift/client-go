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

// FakeAlertingRules implements AlertingRuleInterface
type FakeAlertingRules struct {
	Fake *FakeMonitoringV1
	ns   string
}

var alertingrulesResource = v1.SchemeGroupVersion.WithResource("alertingrules")

var alertingrulesKind = v1.SchemeGroupVersion.WithKind("AlertingRule")

// Get takes name of the alertingRule, and returns the corresponding alertingRule object, and an error if there is any.
func (c *FakeAlertingRules) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.AlertingRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(alertingrulesResource, c.ns, name), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}

// List takes label and field selectors, and returns the list of AlertingRules that match those selectors.
func (c *FakeAlertingRules) List(ctx context.Context, opts metav1.ListOptions) (result *v1.AlertingRuleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(alertingrulesResource, alertingrulesKind, c.ns, opts), &v1.AlertingRuleList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.AlertingRuleList{ListMeta: obj.(*v1.AlertingRuleList).ListMeta}
	for _, item := range obj.(*v1.AlertingRuleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested alertingRules.
func (c *FakeAlertingRules) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(alertingrulesResource, c.ns, opts))

}

// Create takes the representation of a alertingRule and creates it.  Returns the server's representation of the alertingRule, and an error, if there is any.
func (c *FakeAlertingRules) Create(ctx context.Context, alertingRule *v1.AlertingRule, opts metav1.CreateOptions) (result *v1.AlertingRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(alertingrulesResource, c.ns, alertingRule), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}

// Update takes the representation of a alertingRule and updates it. Returns the server's representation of the alertingRule, and an error, if there is any.
func (c *FakeAlertingRules) Update(ctx context.Context, alertingRule *v1.AlertingRule, opts metav1.UpdateOptions) (result *v1.AlertingRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(alertingrulesResource, c.ns, alertingRule), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAlertingRules) UpdateStatus(ctx context.Context, alertingRule *v1.AlertingRule, opts metav1.UpdateOptions) (*v1.AlertingRule, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(alertingrulesResource, "status", c.ns, alertingRule), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}

// Delete takes name of the alertingRule and deletes it. Returns an error if one occurs.
func (c *FakeAlertingRules) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(alertingrulesResource, c.ns, name, opts), &v1.AlertingRule{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAlertingRules) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(alertingrulesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.AlertingRuleList{})
	return err
}

// Patch applies the patch and returns the patched alertingRule.
func (c *FakeAlertingRules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AlertingRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(alertingrulesResource, c.ns, name, pt, data, subresources...), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied alertingRule.
func (c *FakeAlertingRules) Apply(ctx context.Context, alertingRule *monitoringv1.AlertingRuleApplyConfiguration, opts metav1.ApplyOptions) (result *v1.AlertingRule, err error) {
	if alertingRule == nil {
		return nil, fmt.Errorf("alertingRule provided to Apply must not be nil")
	}
	data, err := json.Marshal(alertingRule)
	if err != nil {
		return nil, err
	}
	name := alertingRule.Name
	if name == nil {
		return nil, fmt.Errorf("alertingRule.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(alertingrulesResource, c.ns, *name, types.ApplyPatchType, data), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeAlertingRules) ApplyStatus(ctx context.Context, alertingRule *monitoringv1.AlertingRuleApplyConfiguration, opts metav1.ApplyOptions) (result *v1.AlertingRule, err error) {
	if alertingRule == nil {
		return nil, fmt.Errorf("alertingRule provided to Apply must not be nil")
	}
	data, err := json.Marshal(alertingRule)
	if err != nil {
		return nil, err
	}
	name := alertingRule.Name
	if name == nil {
		return nil, fmt.Errorf("alertingRule.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(alertingrulesResource, c.ns, *name, types.ApplyPatchType, data, "status"), &v1.AlertingRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.AlertingRule), err
}
