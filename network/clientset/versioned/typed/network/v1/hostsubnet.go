// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	v1 "github.com/openshift/api/network/v1"
	networkv1 "github.com/openshift/client-go/network/applyconfigurations/network/v1"
	scheme "github.com/openshift/client-go/network/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HostSubnetsGetter has a method to return a HostSubnetInterface.
// A group's client should implement this interface.
type HostSubnetsGetter interface {
	HostSubnets() HostSubnetInterface
}

// HostSubnetInterface has methods to work with HostSubnet resources.
type HostSubnetInterface interface {
	Create(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.CreateOptions) (*v1.HostSubnet, error)
	Update(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.UpdateOptions) (*v1.HostSubnet, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.HostSubnet, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.HostSubnetList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.HostSubnet, err error)
	Apply(ctx context.Context, hostSubnet *networkv1.HostSubnetApplyConfiguration, opts metav1.ApplyOptions) (result *v1.HostSubnet, err error)
	HostSubnetExpansion
}

// hostSubnets implements HostSubnetInterface
type hostSubnets struct {
	client rest.Interface
}

// newHostSubnets returns a HostSubnets
func newHostSubnets(c *NetworkV1Client) *hostSubnets {
	return &hostSubnets{
		client: c.RESTClient(),
	}
}

// Get takes name of the hostSubnet, and returns the corresponding hostSubnet object, and an error if there is any.
func (c *hostSubnets) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.HostSubnet, err error) {
	result = &v1.HostSubnet{}
	err = c.client.Get().
		Resource("hostsubnets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of HostSubnets that match those selectors.
func (c *hostSubnets) List(ctx context.Context, opts metav1.ListOptions) (result *v1.HostSubnetList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.HostSubnetList{}
	err = c.client.Get().
		Resource("hostsubnets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested hostSubnets.
func (c *hostSubnets) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("hostsubnets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a hostSubnet and creates it.  Returns the server's representation of the hostSubnet, and an error, if there is any.
func (c *hostSubnets) Create(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.CreateOptions) (result *v1.HostSubnet, err error) {
	result = &v1.HostSubnet{}
	err = c.client.Post().
		Resource("hostsubnets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hostSubnet).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a hostSubnet and updates it. Returns the server's representation of the hostSubnet, and an error, if there is any.
func (c *hostSubnets) Update(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.UpdateOptions) (result *v1.HostSubnet, err error) {
	result = &v1.HostSubnet{}
	err = c.client.Put().
		Resource("hostsubnets").
		Name(hostSubnet.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hostSubnet).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the hostSubnet and deletes it. Returns an error if one occurs.
func (c *hostSubnets) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("hostsubnets").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hostSubnets) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("hostsubnets").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched hostSubnet.
func (c *hostSubnets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.HostSubnet, err error) {
	result = &v1.HostSubnet{}
	err = c.client.Patch(pt).
		Resource("hostsubnets").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied hostSubnet.
func (c *hostSubnets) Apply(ctx context.Context, hostSubnet *networkv1.HostSubnetApplyConfiguration, opts metav1.ApplyOptions) (result *v1.HostSubnet, err error) {
	if hostSubnet == nil {
		return nil, fmt.Errorf("hostSubnet provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(hostSubnet)
	if err != nil {
		return nil, err
	}
	name := hostSubnet.Name
	if name == nil {
		return nil, fmt.Errorf("hostSubnet.Name must be provided to Apply")
	}
	result = &v1.HostSubnet{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("hostsubnets").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
