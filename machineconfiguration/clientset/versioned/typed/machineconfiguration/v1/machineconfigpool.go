// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	v1 "github.com/openshift/api/machineconfiguration/v1"
	machineconfigurationv1 "github.com/openshift/client-go/machineconfiguration/applyconfigurations/machineconfiguration/v1"
	scheme "github.com/openshift/client-go/machineconfiguration/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MachineConfigPoolsGetter has a method to return a MachineConfigPoolInterface.
// A group's client should implement this interface.
type MachineConfigPoolsGetter interface {
	MachineConfigPools() MachineConfigPoolInterface
}

// MachineConfigPoolInterface has methods to work with MachineConfigPool resources.
type MachineConfigPoolInterface interface {
	Create(ctx context.Context, machineConfigPool *v1.MachineConfigPool, opts metav1.CreateOptions) (*v1.MachineConfigPool, error)
	Update(ctx context.Context, machineConfigPool *v1.MachineConfigPool, opts metav1.UpdateOptions) (*v1.MachineConfigPool, error)
	UpdateStatus(ctx context.Context, machineConfigPool *v1.MachineConfigPool, opts metav1.UpdateOptions) (*v1.MachineConfigPool, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.MachineConfigPool, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.MachineConfigPoolList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MachineConfigPool, err error)
	Apply(ctx context.Context, machineConfigPool *machineconfigurationv1.MachineConfigPoolApplyConfiguration, opts metav1.ApplyOptions) (result *v1.MachineConfigPool, err error)
	ApplyStatus(ctx context.Context, machineConfigPool *machineconfigurationv1.MachineConfigPoolApplyConfiguration, opts metav1.ApplyOptions) (result *v1.MachineConfigPool, err error)
	MachineConfigPoolExpansion
}

// machineConfigPools implements MachineConfigPoolInterface
type machineConfigPools struct {
	client rest.Interface
}

// newMachineConfigPools returns a MachineConfigPools
func newMachineConfigPools(c *MachineconfigurationV1Client) *machineConfigPools {
	return &machineConfigPools{
		client: c.RESTClient(),
	}
}

// Get takes name of the machineConfigPool, and returns the corresponding machineConfigPool object, and an error if there is any.
func (c *machineConfigPools) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.MachineConfigPool, err error) {
	result = &v1.MachineConfigPool{}
	err = c.client.Get().
		Resource("machineconfigpools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MachineConfigPools that match those selectors.
func (c *machineConfigPools) List(ctx context.Context, opts metav1.ListOptions) (result *v1.MachineConfigPoolList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.MachineConfigPoolList{}
	err = c.client.Get().
		Resource("machineconfigpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested machineConfigPools.
func (c *machineConfigPools) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("machineconfigpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a machineConfigPool and creates it.  Returns the server's representation of the machineConfigPool, and an error, if there is any.
func (c *machineConfigPools) Create(ctx context.Context, machineConfigPool *v1.MachineConfigPool, opts metav1.CreateOptions) (result *v1.MachineConfigPool, err error) {
	result = &v1.MachineConfigPool{}
	err = c.client.Post().
		Resource("machineconfigpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(machineConfigPool).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a machineConfigPool and updates it. Returns the server's representation of the machineConfigPool, and an error, if there is any.
func (c *machineConfigPools) Update(ctx context.Context, machineConfigPool *v1.MachineConfigPool, opts metav1.UpdateOptions) (result *v1.MachineConfigPool, err error) {
	result = &v1.MachineConfigPool{}
	err = c.client.Put().
		Resource("machineconfigpools").
		Name(machineConfigPool.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(machineConfigPool).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *machineConfigPools) UpdateStatus(ctx context.Context, machineConfigPool *v1.MachineConfigPool, opts metav1.UpdateOptions) (result *v1.MachineConfigPool, err error) {
	result = &v1.MachineConfigPool{}
	err = c.client.Put().
		Resource("machineconfigpools").
		Name(machineConfigPool.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(machineConfigPool).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the machineConfigPool and deletes it. Returns an error if one occurs.
func (c *machineConfigPools) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("machineconfigpools").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *machineConfigPools) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("machineconfigpools").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched machineConfigPool.
func (c *machineConfigPools) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MachineConfigPool, err error) {
	result = &v1.MachineConfigPool{}
	err = c.client.Patch(pt).
		Resource("machineconfigpools").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied machineConfigPool.
func (c *machineConfigPools) Apply(ctx context.Context, machineConfigPool *machineconfigurationv1.MachineConfigPoolApplyConfiguration, opts metav1.ApplyOptions) (result *v1.MachineConfigPool, err error) {
	if machineConfigPool == nil {
		return nil, fmt.Errorf("machineConfigPool provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(machineConfigPool)
	if err != nil {
		return nil, err
	}
	name := machineConfigPool.Name
	if name == nil {
		return nil, fmt.Errorf("machineConfigPool.Name must be provided to Apply")
	}
	result = &v1.MachineConfigPool{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("machineconfigpools").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *machineConfigPools) ApplyStatus(ctx context.Context, machineConfigPool *machineconfigurationv1.MachineConfigPoolApplyConfiguration, opts metav1.ApplyOptions) (result *v1.MachineConfigPool, err error) {
	if machineConfigPool == nil {
		return nil, fmt.Errorf("machineConfigPool provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(machineConfigPool)
	if err != nil {
		return nil, err
	}

	name := machineConfigPool.Name
	if name == nil {
		return nil, fmt.Errorf("machineConfigPool.Name must be provided to Apply")
	}

	result = &v1.MachineConfigPool{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("machineconfigpools").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
