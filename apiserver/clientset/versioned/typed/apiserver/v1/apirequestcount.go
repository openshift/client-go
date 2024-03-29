// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	v1 "github.com/openshift/api/apiserver/v1"
	apiserverv1 "github.com/openshift/client-go/apiserver/applyconfigurations/apiserver/v1"
	scheme "github.com/openshift/client-go/apiserver/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// APIRequestCountsGetter has a method to return a APIRequestCountInterface.
// A group's client should implement this interface.
type APIRequestCountsGetter interface {
	APIRequestCounts() APIRequestCountInterface
}

// APIRequestCountInterface has methods to work with APIRequestCount resources.
type APIRequestCountInterface interface {
	Create(ctx context.Context, aPIRequestCount *v1.APIRequestCount, opts metav1.CreateOptions) (*v1.APIRequestCount, error)
	Update(ctx context.Context, aPIRequestCount *v1.APIRequestCount, opts metav1.UpdateOptions) (*v1.APIRequestCount, error)
	UpdateStatus(ctx context.Context, aPIRequestCount *v1.APIRequestCount, opts metav1.UpdateOptions) (*v1.APIRequestCount, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.APIRequestCount, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.APIRequestCountList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.APIRequestCount, err error)
	Apply(ctx context.Context, aPIRequestCount *apiserverv1.APIRequestCountApplyConfiguration, opts metav1.ApplyOptions) (result *v1.APIRequestCount, err error)
	ApplyStatus(ctx context.Context, aPIRequestCount *apiserverv1.APIRequestCountApplyConfiguration, opts metav1.ApplyOptions) (result *v1.APIRequestCount, err error)
	APIRequestCountExpansion
}

// aPIRequestCounts implements APIRequestCountInterface
type aPIRequestCounts struct {
	client rest.Interface
}

// newAPIRequestCounts returns a APIRequestCounts
func newAPIRequestCounts(c *ApiserverV1Client) *aPIRequestCounts {
	return &aPIRequestCounts{
		client: c.RESTClient(),
	}
}

// Get takes name of the aPIRequestCount, and returns the corresponding aPIRequestCount object, and an error if there is any.
func (c *aPIRequestCounts) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.APIRequestCount, err error) {
	result = &v1.APIRequestCount{}
	err = c.client.Get().
		Resource("apirequestcounts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of APIRequestCounts that match those selectors.
func (c *aPIRequestCounts) List(ctx context.Context, opts metav1.ListOptions) (result *v1.APIRequestCountList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.APIRequestCountList{}
	err = c.client.Get().
		Resource("apirequestcounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested aPIRequestCounts.
func (c *aPIRequestCounts) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("apirequestcounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a aPIRequestCount and creates it.  Returns the server's representation of the aPIRequestCount, and an error, if there is any.
func (c *aPIRequestCounts) Create(ctx context.Context, aPIRequestCount *v1.APIRequestCount, opts metav1.CreateOptions) (result *v1.APIRequestCount, err error) {
	result = &v1.APIRequestCount{}
	err = c.client.Post().
		Resource("apirequestcounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(aPIRequestCount).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a aPIRequestCount and updates it. Returns the server's representation of the aPIRequestCount, and an error, if there is any.
func (c *aPIRequestCounts) Update(ctx context.Context, aPIRequestCount *v1.APIRequestCount, opts metav1.UpdateOptions) (result *v1.APIRequestCount, err error) {
	result = &v1.APIRequestCount{}
	err = c.client.Put().
		Resource("apirequestcounts").
		Name(aPIRequestCount.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(aPIRequestCount).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *aPIRequestCounts) UpdateStatus(ctx context.Context, aPIRequestCount *v1.APIRequestCount, opts metav1.UpdateOptions) (result *v1.APIRequestCount, err error) {
	result = &v1.APIRequestCount{}
	err = c.client.Put().
		Resource("apirequestcounts").
		Name(aPIRequestCount.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(aPIRequestCount).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the aPIRequestCount and deletes it. Returns an error if one occurs.
func (c *aPIRequestCounts) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("apirequestcounts").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *aPIRequestCounts) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("apirequestcounts").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched aPIRequestCount.
func (c *aPIRequestCounts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.APIRequestCount, err error) {
	result = &v1.APIRequestCount{}
	err = c.client.Patch(pt).
		Resource("apirequestcounts").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied aPIRequestCount.
func (c *aPIRequestCounts) Apply(ctx context.Context, aPIRequestCount *apiserverv1.APIRequestCountApplyConfiguration, opts metav1.ApplyOptions) (result *v1.APIRequestCount, err error) {
	if aPIRequestCount == nil {
		return nil, fmt.Errorf("aPIRequestCount provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(aPIRequestCount)
	if err != nil {
		return nil, err
	}
	name := aPIRequestCount.Name
	if name == nil {
		return nil, fmt.Errorf("aPIRequestCount.Name must be provided to Apply")
	}
	result = &v1.APIRequestCount{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("apirequestcounts").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *aPIRequestCounts) ApplyStatus(ctx context.Context, aPIRequestCount *apiserverv1.APIRequestCountApplyConfiguration, opts metav1.ApplyOptions) (result *v1.APIRequestCount, err error) {
	if aPIRequestCount == nil {
		return nil, fmt.Errorf("aPIRequestCount provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(aPIRequestCount)
	if err != nil {
		return nil, err
	}

	name := aPIRequestCount.Name
	if name == nil {
		return nil, fmt.Errorf("aPIRequestCount.Name must be provided to Apply")
	}

	result = &v1.APIRequestCount{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("apirequestcounts").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
