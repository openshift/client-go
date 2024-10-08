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

// FakeConsoleCLIDownloads implements ConsoleCLIDownloadInterface
type FakeConsoleCLIDownloads struct {
	Fake *FakeConsoleV1
}

var consoleclidownloadsResource = v1.SchemeGroupVersion.WithResource("consoleclidownloads")

var consoleclidownloadsKind = v1.SchemeGroupVersion.WithKind("ConsoleCLIDownload")

// Get takes name of the consoleCLIDownload, and returns the corresponding consoleCLIDownload object, and an error if there is any.
func (c *FakeConsoleCLIDownloads) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ConsoleCLIDownload, err error) {
	emptyResult := &v1.ConsoleCLIDownload{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(consoleclidownloadsResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleCLIDownload), err
}

// List takes label and field selectors, and returns the list of ConsoleCLIDownloads that match those selectors.
func (c *FakeConsoleCLIDownloads) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ConsoleCLIDownloadList, err error) {
	emptyResult := &v1.ConsoleCLIDownloadList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(consoleclidownloadsResource, consoleclidownloadsKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ConsoleCLIDownloadList{ListMeta: obj.(*v1.ConsoleCLIDownloadList).ListMeta}
	for _, item := range obj.(*v1.ConsoleCLIDownloadList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested consoleCLIDownloads.
func (c *FakeConsoleCLIDownloads) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(consoleclidownloadsResource, opts))
}

// Create takes the representation of a consoleCLIDownload and creates it.  Returns the server's representation of the consoleCLIDownload, and an error, if there is any.
func (c *FakeConsoleCLIDownloads) Create(ctx context.Context, consoleCLIDownload *v1.ConsoleCLIDownload, opts metav1.CreateOptions) (result *v1.ConsoleCLIDownload, err error) {
	emptyResult := &v1.ConsoleCLIDownload{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(consoleclidownloadsResource, consoleCLIDownload, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleCLIDownload), err
}

// Update takes the representation of a consoleCLIDownload and updates it. Returns the server's representation of the consoleCLIDownload, and an error, if there is any.
func (c *FakeConsoleCLIDownloads) Update(ctx context.Context, consoleCLIDownload *v1.ConsoleCLIDownload, opts metav1.UpdateOptions) (result *v1.ConsoleCLIDownload, err error) {
	emptyResult := &v1.ConsoleCLIDownload{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(consoleclidownloadsResource, consoleCLIDownload, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleCLIDownload), err
}

// Delete takes name of the consoleCLIDownload and deletes it. Returns an error if one occurs.
func (c *FakeConsoleCLIDownloads) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(consoleclidownloadsResource, name, opts), &v1.ConsoleCLIDownload{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConsoleCLIDownloads) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(consoleclidownloadsResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ConsoleCLIDownloadList{})
	return err
}

// Patch applies the patch and returns the patched consoleCLIDownload.
func (c *FakeConsoleCLIDownloads) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ConsoleCLIDownload, err error) {
	emptyResult := &v1.ConsoleCLIDownload{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consoleclidownloadsResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleCLIDownload), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied consoleCLIDownload.
func (c *FakeConsoleCLIDownloads) Apply(ctx context.Context, consoleCLIDownload *consolev1.ConsoleCLIDownloadApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ConsoleCLIDownload, err error) {
	if consoleCLIDownload == nil {
		return nil, fmt.Errorf("consoleCLIDownload provided to Apply must not be nil")
	}
	data, err := json.Marshal(consoleCLIDownload)
	if err != nil {
		return nil, err
	}
	name := consoleCLIDownload.Name
	if name == nil {
		return nil, fmt.Errorf("consoleCLIDownload.Name must be provided to Apply")
	}
	emptyResult := &v1.ConsoleCLIDownload{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(consoleclidownloadsResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ConsoleCLIDownload), err
}
