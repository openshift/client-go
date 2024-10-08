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

// FakeConsoleLinks implements ConsoleLinkInterface
type FakeConsoleLinks struct {
	Fake *FakeConsoleV1
}

var consolelinksResource = v1.SchemeGroupVersion.WithResource("consolelinks")

var consolelinksKind = v1.SchemeGroupVersion.WithKind("ConsoleLink")

// Get takes name of the consoleLink, and returns the corresponding consoleLink object, and an error if there is any.
func (c *FakeConsoleLinks) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ConsoleLink, err error) {
	emptyResult := &v1.ConsoleLink{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(consolelinksResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleLink), err
}

// List takes label and field selectors, and returns the list of ConsoleLinks that match those selectors.
func (c *FakeConsoleLinks) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ConsoleLinkList, err error) {
	emptyResult := &v1.ConsoleLinkList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(consolelinksResource, consolelinksKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ConsoleLinkList{ListMeta: obj.(*v1.ConsoleLinkList).ListMeta}
	for _, item := range obj.(*v1.ConsoleLinkList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested consoleLinks.
func (c *FakeConsoleLinks) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(consolelinksResource, opts))
}

// Create takes the representation of a consoleLink and creates it.  Returns the server's representation of the consoleLink, and an error, if there is any.
func (c *FakeConsoleLinks) Create(ctx context.Context, consoleLink *v1.ConsoleLink, opts metav1.CreateOptions) (result *v1.ConsoleLink, err error) {
	emptyResult := &v1.ConsoleLink{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(consolelinksResource, consoleLink, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleLink), err
}

// Update takes the representation of a consoleLink and updates it. Returns the server's representation of the consoleLink, and an error, if there is any.
func (c *FakeConsoleLinks) Update(ctx context.Context, consoleLink *v1.ConsoleLink, opts metav1.UpdateOptions) (result *v1.ConsoleLink, err error) {
	emptyResult := &v1.ConsoleLink{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(consolelinksResource, consoleLink, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleLink), err
}

// Delete takes name of the consoleLink and deletes it. Returns an error if one occurs.
func (c *FakeConsoleLinks) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(consolelinksResource, name, opts), &v1.ConsoleLink{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConsoleLinks) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(consolelinksResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ConsoleLinkList{})
	return err
}

// Patch applies the patch and returns the patched consoleLink.
func (c *FakeConsoleLinks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ConsoleLink, err error) {
	emptyResult := &v1.ConsoleLink{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consolelinksResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleLink), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied consoleLink.
func (c *FakeConsoleLinks) Apply(ctx context.Context, consoleLink *consolev1.ConsoleLinkApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ConsoleLink, err error) {
	if consoleLink == nil {
		return nil, fmt.Errorf("consoleLink provided to Apply must not be nil")
	}
	data, err := json.Marshal(consoleLink)
	if err != nil {
		return nil, err
	}
	name := consoleLink.Name
	if name == nil {
		return nil, fmt.Errorf("consoleLink.Name must be provided to Apply")
	}
	emptyResult := &v1.ConsoleLink{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consolelinksResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleLink), err
}
