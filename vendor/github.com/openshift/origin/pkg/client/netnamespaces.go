package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kapi "k8s.io/kubernetes/pkg/api"

	networkapi "github.com/openshift/origin/pkg/network/apis/network"
)

// NetNamespaceInterface has methods to work with NetNamespace resources
type NetNamespacesInterface interface {
	NetNamespaces() NetNamespaceInterface
}

// NetNamespaceInterface exposes methods on NetNamespace resources.
type NetNamespaceInterface interface {
	List(opts metav1.ListOptions) (*networkapi.NetNamespaceList, error)
	Get(name string, options metav1.GetOptions) (*networkapi.NetNamespace, error)
	Create(sub *networkapi.NetNamespace) (*networkapi.NetNamespace, error)
	Update(sub *networkapi.NetNamespace) (*networkapi.NetNamespace, error)
	Delete(name string) error
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

// netNamespace implements NetNamespaceInterface interface
type netNamespace struct {
	r *Client
}

// newNetNamespace returns a NetNamespace
func newNetNamespace(c *Client) *netNamespace {
	return &netNamespace{
		r: c,
	}
}

// List returns a list of NetNamespaces that match the label and field selectors.
func (c *netNamespace) List(opts metav1.ListOptions) (result *networkapi.NetNamespaceList, err error) {
	result = &networkapi.NetNamespaceList{}
	err = c.r.Get().
		Resource("netNamespaces").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular NetNamespace or an error
func (c *netNamespace) Get(netname string, options metav1.GetOptions) (result *networkapi.NetNamespace, err error) {
	result = &networkapi.NetNamespace{}
	err = c.r.Get().Resource("netNamespaces").Name(netname).Do().Into(result)
	return
}

// Create creates a new NetNamespace. Returns the server's representation of the NetNamespace and error if one occurs.
func (c *netNamespace) Create(netNamespace *networkapi.NetNamespace) (result *networkapi.NetNamespace, err error) {
	result = &networkapi.NetNamespace{}
	err = c.r.Post().Resource("netNamespaces").Body(netNamespace).Do().Into(result)
	return
}

// Update updates the NetNamespace. Returns the server's representation of the NetNamespace and error if one occurs.
func (c *netNamespace) Update(netNamespace *networkapi.NetNamespace) (result *networkapi.NetNamespace, err error) {
	result = &networkapi.NetNamespace{}
	err = c.r.Put().Resource("netNamespaces").Name(netNamespace.Name).Body(netNamespace).Do().Into(result)
	return
}

// Delete takes the name of the NetNamespace, and returns an error if one occurs during deletion of the NetNamespace
func (c *netNamespace) Delete(name string) error {
	return c.r.Delete().Resource("netNamespaces").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested NetNamespaces
func (c *netNamespace) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Resource("netNamespaces").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}
