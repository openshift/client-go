package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	kapi "k8s.io/kubernetes/pkg/api"
	extensionsv1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"

	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
)

// DeploymentConfigsNamespacer has methods to work with DeploymentConfig resources in a namespace
type DeploymentConfigsNamespacer interface {
	DeploymentConfigs(namespace string) DeploymentConfigInterface
}

// DeploymentConfigInterface contains methods for working with DeploymentConfigs
type DeploymentConfigInterface interface {
	List(opts metav1.ListOptions) (*deployapi.DeploymentConfigList, error)
	Get(name string, options metav1.GetOptions) (*deployapi.DeploymentConfig, error)
	Create(config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
	Update(config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *deployapi.DeploymentConfig, err error)
	Delete(name string) error
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Rollback(config *deployapi.DeploymentConfigRollback) (*deployapi.DeploymentConfig, error)
	RollbackDeprecated(config *deployapi.DeploymentConfigRollback) (*deployapi.DeploymentConfig, error)
	GetScale(name string) (*extensionsv1beta1.Scale, error)
	UpdateScale(scale *extensionsv1beta1.Scale) (*extensionsv1beta1.Scale, error)
	UpdateStatus(config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
	Instantiate(request *deployapi.DeploymentRequest) (*deployapi.DeploymentConfig, error)
}

// deploymentConfigs implements DeploymentConfigsNamespacer interface
type deploymentConfigs struct {
	r  *Client
	ns string
}

// newDeploymentConfigs returns a deploymentConfigs
func newDeploymentConfigs(c *Client, namespace string) *deploymentConfigs {
	return &deploymentConfigs{
		r:  c,
		ns: namespace,
	}
}

// List takes a label and field selectors, and returns the list of deploymentConfigs that match that selectors
func (c *deploymentConfigs) List(opts metav1.ListOptions) (result *deployapi.DeploymentConfigList, err error) {
	result = &deployapi.DeploymentConfigList{}
	err = c.r.Get().
		Namespace(c.ns).
		Resource("deploymentConfigs").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular deploymentConfig
func (c *deploymentConfigs) Get(name string, options metav1.GetOptions) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Get().Namespace(c.ns).Resource("deploymentConfigs").Name(name).VersionedParams(&options, kapi.ParameterCodec).Do().Into(result)
	return
}

// Create creates a new deploymentConfig
func (c *deploymentConfigs) Create(deploymentConfig *deployapi.DeploymentConfig) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Post().Namespace(c.ns).Resource("deploymentConfigs").Body(deploymentConfig).Do().Into(result)
	return
}

// Update updates an existing deploymentConfig
func (c *deploymentConfigs) Update(deploymentConfig *deployapi.DeploymentConfig) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Put().Namespace(c.ns).Resource("deploymentConfigs").Name(deploymentConfig.Name).Body(deploymentConfig).Do().Into(result)
	return
}

// Patch takes the partial representation of a deployment config and updates it.
// Returns the server's representation of the deployment config, and an error, if there is any.
func (c *deploymentConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Patch(types.StrategicMergePatchType).Namespace(c.ns).Resource("deploymentConfigs").SubResource(subresources...).Name(name).Body(data).Do().Into(result)
	return
}

// Delete deletes an existing deploymentConfig.
func (c *deploymentConfigs) Delete(name string) error {
	return c.r.Delete().Namespace(c.ns).Resource("deploymentConfigs").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested deploymentConfigs.
func (c *deploymentConfigs) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("deploymentConfigs").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}

// Rollback rolls a deploymentConfig back to a previous configuration
func (c *deploymentConfigs) Rollback(config *deployapi.DeploymentConfigRollback) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource("deploymentConfigs").
		Name(config.Name).
		SubResource("rollback").
		Body(config).
		Do().
		Into(result)
	return
}

// RollbackDeprecated rolls a deploymentConfig back to a previous configuration
func (c *deploymentConfigs) RollbackDeprecated(config *deployapi.DeploymentConfigRollback) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Post().
		Namespace(c.ns).
		Resource("deploymentConfigRollbacks").
		Body(config).
		Do().
		Into(result)
	return
}

// GetScale returns information about a particular deploymentConfig via its scale subresource
func (c *deploymentConfigs) GetScale(name string) (result *extensionsv1beta1.Scale, err error) {
	result = &extensionsv1beta1.Scale{}
	err = c.r.Get().Namespace(c.ns).Resource("deploymentConfigs").Name(name).SubResource("scale").Do().Into(result)
	return
}

// UpdateScale scales an existing deploymentConfig via its scale subresource
func (c *deploymentConfigs) UpdateScale(scale *extensionsv1beta1.Scale) (result *extensionsv1beta1.Scale, err error) {
	result = &extensionsv1beta1.Scale{}

	// TODO fix by making the client understand how to encode using different codecs for different resources
	encodedBytes, err := runtime.Encode(kapi.Codecs.LegacyCodec(extensionsv1beta1.SchemeGroupVersion), scale)
	if err != nil {
		return result, err
	}

	err = c.r.Put().Namespace(c.ns).Resource("deploymentConfigs").Name(scale.Name).SubResource("scale").Body(encodedBytes).Do().Into(result)
	return
}

// UpdateStatus updates the status for an existing deploymentConfig.
func (c *deploymentConfigs) UpdateStatus(deploymentConfig *deployapi.DeploymentConfig) (result *deployapi.DeploymentConfig, err error) {
	result = &deployapi.DeploymentConfig{}
	err = c.r.Put().Namespace(c.ns).Resource("deploymentConfigs").Name(deploymentConfig.Name).SubResource("status").Body(deploymentConfig).Do().Into(result)
	return
}

// Instantiate instantiates a new build from build config returning new object or an error
func (c *deploymentConfigs) Instantiate(request *deployapi.DeploymentRequest) (*deployapi.DeploymentConfig, error) {
	result := &deployapi.DeploymentConfig{}
	resp := c.r.Post().Namespace(c.ns).Resource("deploymentConfigs").Name(request.Name).SubResource("instantiate").Body(request).Do()
	var statusCode int
	if resp.StatusCode(&statusCode); statusCode == 204 {
		return nil, nil
	}
	err := resp.Into(result)
	return result, err
}
