package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kapi "k8s.io/kubernetes/pkg/api"

	userapi "github.com/openshift/origin/pkg/user/apis/user"
)

// UsersInterface has methods to work with User resources
type UsersInterface interface {
	Users() UserInterface
}

// UserInterface exposes methods on user resources.
type UserInterface interface {
	List(opts metav1.ListOptions) (*userapi.UserList, error)
	Get(name string, options metav1.GetOptions) (*userapi.User, error)
	Create(user *userapi.User) (*userapi.User, error)
	Update(user *userapi.User) (*userapi.User, error)
	Delete(name string) error
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

// users implements UserInterface interface
type users struct {
	r *Client
}

// newUsers returns a users
func newUsers(c *Client) *users {
	return &users{
		r: c,
	}
}

// List returns a list of users that match the label and field selectors.
func (c *users) List(opts metav1.ListOptions) (result *userapi.UserList, err error) {
	result = &userapi.UserList{}
	err = c.r.Get().
		Resource("users").
		VersionedParams(&opts, kapi.ParameterCodec).
		Do().
		Into(result)
	return
}

// Get returns information about a particular user or an error
func (c *users) Get(name string, options metav1.GetOptions) (result *userapi.User, err error) {
	result = &userapi.User{}
	err = c.r.Get().Resource("users").Name(name).VersionedParams(&options, kapi.ParameterCodec).Do().Into(result)
	return
}

// Create creates a new user. Returns the server's representation of the user and error if one occurs.
func (c *users) Create(user *userapi.User) (result *userapi.User, err error) {
	result = &userapi.User{}
	err = c.r.Post().Resource("users").Body(user).Do().Into(result)
	return
}

// Update updates the user on server. Returns the server's representation of the user and error if one occurs.
func (c *users) Update(user *userapi.User) (result *userapi.User, err error) {
	result = &userapi.User{}
	err = c.r.Put().Resource("users").Name(user.Name).Body(user).Do().Into(result)
	return
}

// Delete deletes the user on server. Returns an error if one occurs.
func (c *users) Delete(name string) (err error) {
	return c.r.Delete().Resource("users").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested users.
func (c *users) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Resource("users").
		VersionedParams(&opts, kapi.ParameterCodec).
		Watch()
}
