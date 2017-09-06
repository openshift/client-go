package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
)

// ClusterRoleBindingsInterface has methods to work with ClusterRoleBindings resources in a namespace
type ClusterRoleBindingsInterface interface {
	ClusterRoleBindings() ClusterRoleBindingInterface
}

// ClusterRoleBindingInterface exposes methods on ClusterRoleBindings resources
type ClusterRoleBindingInterface interface {
	List(opts metav1.ListOptions) (*authorizationapi.ClusterRoleBindingList, error)
	Get(name string, options metav1.GetOptions) (*authorizationapi.ClusterRoleBinding, error)
	Update(roleBinding *authorizationapi.ClusterRoleBinding) (*authorizationapi.ClusterRoleBinding, error)
	Create(roleBinding *authorizationapi.ClusterRoleBinding) (*authorizationapi.ClusterRoleBinding, error)
	Delete(name string) error
}

type clusterRoleBindings struct {
	r *Client
}

// newClusterRoleBindings returns a clusterRoleBindings
func newClusterRoleBindings(c *Client) *clusterRoleBindings {
	return &clusterRoleBindings{
		r: c,
	}
}

// List returns a list of clusterRoleBindings that match the label and field selectors.
func (c *clusterRoleBindings) List(opts metav1.ListOptions) (result *authorizationapi.ClusterRoleBindingList, err error) {
	result = &authorizationapi.ClusterRoleBindingList{}
	err = c.r.Get().Resource("clusterRoleBindings").VersionedParams(&opts, kapi.ParameterCodec).Do().Into(result)
	return
}

// Get returns information about a particular roleBinding and error if one occurs.
func (c *clusterRoleBindings) Get(name string, options metav1.GetOptions) (result *authorizationapi.ClusterRoleBinding, err error) {
	result = &authorizationapi.ClusterRoleBinding{}
	err = c.r.Get().Resource("clusterRoleBindings").Name(name).VersionedParams(&options, kapi.ParameterCodec).Do().Into(result)
	return
}

// Create creates new roleBinding. Returns the server's representation of the roleBinding and error if one occurs.
func (c *clusterRoleBindings) Create(roleBinding *authorizationapi.ClusterRoleBinding) (result *authorizationapi.ClusterRoleBinding, err error) {
	result = &authorizationapi.ClusterRoleBinding{}
	err = c.r.Post().Resource("clusterRoleBindings").Body(roleBinding).Do().Into(result)
	return
}

// Update updates the roleBinding on server. Returns the server's representation of the roleBinding and error if one occurs.
func (c *clusterRoleBindings) Update(roleBinding *authorizationapi.ClusterRoleBinding) (result *authorizationapi.ClusterRoleBinding, err error) {
	result = &authorizationapi.ClusterRoleBinding{}
	err = c.r.Put().Resource("clusterRoleBindings").Name(roleBinding.Name).Body(roleBinding).Do().Into(result)
	return
}

// Delete deletes a roleBinding, returns error if one occurs.
func (c *clusterRoleBindings) Delete(name string) (err error) {
	err = c.r.Delete().Resource("clusterRoleBindings").Name(name).Do().Error()
	return
}
