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

// KubeletConfigsGetter has a method to return a KubeletConfigInterface.
// A group's client should implement this interface.
type KubeletConfigsGetter interface {
	KubeletConfigs() KubeletConfigInterface
}

// KubeletConfigInterface has methods to work with KubeletConfig resources.
type KubeletConfigInterface interface {
	Create(ctx context.Context, kubeletConfig *v1.KubeletConfig, opts metav1.CreateOptions) (*v1.KubeletConfig, error)
	Update(ctx context.Context, kubeletConfig *v1.KubeletConfig, opts metav1.UpdateOptions) (*v1.KubeletConfig, error)
	UpdateStatus(ctx context.Context, kubeletConfig *v1.KubeletConfig, opts metav1.UpdateOptions) (*v1.KubeletConfig, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.KubeletConfig, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.KubeletConfigList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeletConfig, err error)
	Apply(ctx context.Context, kubeletConfig *machineconfigurationv1.KubeletConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeletConfig, err error)
	ApplyStatus(ctx context.Context, kubeletConfig *machineconfigurationv1.KubeletConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeletConfig, err error)
	KubeletConfigExpansion
}

// kubeletConfigs implements KubeletConfigInterface
type kubeletConfigs struct {
	client rest.Interface
}

// newKubeletConfigs returns a KubeletConfigs
func newKubeletConfigs(c *MachineconfigurationV1Client) *kubeletConfigs {
	return &kubeletConfigs{
		client: c.RESTClient(),
	}
}

// Get takes name of the kubeletConfig, and returns the corresponding kubeletConfig object, and an error if there is any.
func (c *kubeletConfigs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.KubeletConfig, err error) {
	result = &v1.KubeletConfig{}
	err = c.client.Get().
		Resource("kubeletconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KubeletConfigs that match those selectors.
func (c *kubeletConfigs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.KubeletConfigList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.KubeletConfigList{}
	err = c.client.Get().
		Resource("kubeletconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kubeletConfigs.
func (c *kubeletConfigs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("kubeletconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kubeletConfig and creates it.  Returns the server's representation of the kubeletConfig, and an error, if there is any.
func (c *kubeletConfigs) Create(ctx context.Context, kubeletConfig *v1.KubeletConfig, opts metav1.CreateOptions) (result *v1.KubeletConfig, err error) {
	result = &v1.KubeletConfig{}
	err = c.client.Post().
		Resource("kubeletconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeletConfig).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kubeletConfig and updates it. Returns the server's representation of the kubeletConfig, and an error, if there is any.
func (c *kubeletConfigs) Update(ctx context.Context, kubeletConfig *v1.KubeletConfig, opts metav1.UpdateOptions) (result *v1.KubeletConfig, err error) {
	result = &v1.KubeletConfig{}
	err = c.client.Put().
		Resource("kubeletconfigs").
		Name(kubeletConfig.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeletConfig).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kubeletConfigs) UpdateStatus(ctx context.Context, kubeletConfig *v1.KubeletConfig, opts metav1.UpdateOptions) (result *v1.KubeletConfig, err error) {
	result = &v1.KubeletConfig{}
	err = c.client.Put().
		Resource("kubeletconfigs").
		Name(kubeletConfig.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeletConfig).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kubeletConfig and deletes it. Returns an error if one occurs.
func (c *kubeletConfigs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("kubeletconfigs").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kubeletConfigs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("kubeletconfigs").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kubeletConfig.
func (c *kubeletConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeletConfig, err error) {
	result = &v1.KubeletConfig{}
	err = c.client.Patch(pt).
		Resource("kubeletconfigs").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied kubeletConfig.
func (c *kubeletConfigs) Apply(ctx context.Context, kubeletConfig *machineconfigurationv1.KubeletConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeletConfig, err error) {
	if kubeletConfig == nil {
		return nil, fmt.Errorf("kubeletConfig provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(kubeletConfig)
	if err != nil {
		return nil, err
	}
	name := kubeletConfig.Name
	if name == nil {
		return nil, fmt.Errorf("kubeletConfig.Name must be provided to Apply")
	}
	result = &v1.KubeletConfig{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("kubeletconfigs").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *kubeletConfigs) ApplyStatus(ctx context.Context, kubeletConfig *machineconfigurationv1.KubeletConfigApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeletConfig, err error) {
	if kubeletConfig == nil {
		return nil, fmt.Errorf("kubeletConfig provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(kubeletConfig)
	if err != nil {
		return nil, err
	}

	name := kubeletConfig.Name
	if name == nil {
		return nil, fmt.Errorf("kubeletConfig.Name must be provided to Apply")
	}

	result = &v1.KubeletConfig{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("kubeletconfigs").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
