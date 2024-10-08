// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1 "github.com/openshift/api/console/v1"
	consolev1 "github.com/openshift/client-go/console/applyconfigurations/console/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeConsoleQuickStarts implements ConsoleQuickStartInterface
type FakeConsoleQuickStarts struct {
	Fake *FakeConsoleV1
}

var consolequickstartsResource = v1.SchemeGroupVersion.WithResource("consolequickstarts")

var consolequickstartsKind = v1.SchemeGroupVersion.WithKind("ConsoleQuickStart")

// Get takes name of the consoleQuickStart, and returns the corresponding consoleQuickStart object, and an error if there is any.
func (c *FakeConsoleQuickStarts) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ConsoleQuickStart, err error) {
	emptyResult := &v1.ConsoleQuickStart{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(consolequickstartsResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleQuickStart), err
}

// List takes label and field selectors, and returns the list of ConsoleQuickStarts that match those selectors.
func (c *FakeConsoleQuickStarts) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ConsoleQuickStartList, err error) {
	emptyResult := &v1.ConsoleQuickStartList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(consolequickstartsResource, consolequickstartsKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ConsoleQuickStartList{ListMeta: obj.(*v1.ConsoleQuickStartList).ListMeta}
	for _, item := range obj.(*v1.ConsoleQuickStartList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested consoleQuickStarts.
func (c *FakeConsoleQuickStarts) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(consolequickstartsResource, opts))
}

// Create takes the representation of a consoleQuickStart and creates it.  Returns the server's representation of the consoleQuickStart, and an error, if there is any.
func (c *FakeConsoleQuickStarts) Create(ctx context.Context, consoleQuickStart *v1.ConsoleQuickStart, opts metav1.CreateOptions) (result *v1.ConsoleQuickStart, err error) {
	emptyResult := &v1.ConsoleQuickStart{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(consolequickstartsResource, consoleQuickStart, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleQuickStart), err
}

// Update takes the representation of a consoleQuickStart and updates it. Returns the server's representation of the consoleQuickStart, and an error, if there is any.
func (c *FakeConsoleQuickStarts) Update(ctx context.Context, consoleQuickStart *v1.ConsoleQuickStart, opts metav1.UpdateOptions) (result *v1.ConsoleQuickStart, err error) {
	emptyResult := &v1.ConsoleQuickStart{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(consolequickstartsResource, consoleQuickStart, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleQuickStart), err
}

// Delete takes name of the consoleQuickStart and deletes it. Returns an error if one occurs.
func (c *FakeConsoleQuickStarts) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(consolequickstartsResource, name, opts), &v1.ConsoleQuickStart{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConsoleQuickStarts) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(consolequickstartsResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ConsoleQuickStartList{})
	return err
}

// Patch applies the patch and returns the patched consoleQuickStart.
func (c *FakeConsoleQuickStarts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ConsoleQuickStart, err error) {
	emptyResult := &v1.ConsoleQuickStart{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consolequickstartsResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleQuickStart), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied consoleQuickStart.
func (c *FakeConsoleQuickStarts) Apply(ctx context.Context, consoleQuickStart *consolev1.ConsoleQuickStartApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ConsoleQuickStart, err error) {
	if consoleQuickStart == nil {
		return nil, fmt.Errorf("consoleQuickStart provided to Apply must not be nil")
	}
	data, err := json.Marshal(consoleQuickStart)
	if err != nil {
		return nil, err
	}
	name := consoleQuickStart.Name
	if name == nil {
		return nil, fmt.Errorf("consoleQuickStart.Name must be provided to Apply")
	}
	emptyResult := &v1.ConsoleQuickStart{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consolequickstartsResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleQuickStart), err
}
