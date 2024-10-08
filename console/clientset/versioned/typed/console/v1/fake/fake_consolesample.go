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

// FakeConsoleSamples implements ConsoleSampleInterface
type FakeConsoleSamples struct {
	Fake *FakeConsoleV1
}

var consolesamplesResource = v1.SchemeGroupVersion.WithResource("consolesamples")

var consolesamplesKind = v1.SchemeGroupVersion.WithKind("ConsoleSample")

// Get takes name of the consoleSample, and returns the corresponding consoleSample object, and an error if there is any.
func (c *FakeConsoleSamples) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ConsoleSample, err error) {
	emptyResult := &v1.ConsoleSample{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(consolesamplesResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleSample), err
}

// List takes label and field selectors, and returns the list of ConsoleSamples that match those selectors.
func (c *FakeConsoleSamples) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ConsoleSampleList, err error) {
	emptyResult := &v1.ConsoleSampleList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(consolesamplesResource, consolesamplesKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ConsoleSampleList{ListMeta: obj.(*v1.ConsoleSampleList).ListMeta}
	for _, item := range obj.(*v1.ConsoleSampleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested consoleSamples.
func (c *FakeConsoleSamples) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(consolesamplesResource, opts))
}

// Create takes the representation of a consoleSample and creates it.  Returns the server's representation of the consoleSample, and an error, if there is any.
func (c *FakeConsoleSamples) Create(ctx context.Context, consoleSample *v1.ConsoleSample, opts metav1.CreateOptions) (result *v1.ConsoleSample, err error) {
	emptyResult := &v1.ConsoleSample{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(consolesamplesResource, consoleSample, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleSample), err
}

// Update takes the representation of a consoleSample and updates it. Returns the server's representation of the consoleSample, and an error, if there is any.
func (c *FakeConsoleSamples) Update(ctx context.Context, consoleSample *v1.ConsoleSample, opts metav1.UpdateOptions) (result *v1.ConsoleSample, err error) {
	emptyResult := &v1.ConsoleSample{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(consolesamplesResource, consoleSample, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleSample), err
}

// Delete takes name of the consoleSample and deletes it. Returns an error if one occurs.
func (c *FakeConsoleSamples) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(consolesamplesResource, name, opts), &v1.ConsoleSample{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConsoleSamples) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(consolesamplesResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ConsoleSampleList{})
	return err
}

// Patch applies the patch and returns the patched consoleSample.
func (c *FakeConsoleSamples) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ConsoleSample, err error) {
	emptyResult := &v1.ConsoleSample{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consolesamplesResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleSample), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied consoleSample.
func (c *FakeConsoleSamples) Apply(ctx context.Context, consoleSample *consolev1.ConsoleSampleApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ConsoleSample, err error) {
	if consoleSample == nil {
		return nil, fmt.Errorf("consoleSample provided to Apply must not be nil")
	}
	data, err := json.Marshal(consoleSample)
	if err != nil {
		return nil, err
	}
	name := consoleSample.Name
	if name == nil {
		return nil, fmt.Errorf("consoleSample.Name must be provided to Apply")
	}
	emptyResult := &v1.ConsoleSample{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consolesamplesResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleSample), err
}
