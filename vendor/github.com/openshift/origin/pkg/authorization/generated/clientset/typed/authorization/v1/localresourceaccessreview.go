package v1

import (
	v1 "github.com/openshift/origin/pkg/authorization/apis/authorization/v1"
	rest "k8s.io/client-go/rest"
)

// LocalResourceAccessReviewsGetter has a method to return a LocalResourceAccessReviewInterface.
// A group's client should implement this interface.
type LocalResourceAccessReviewsGetter interface {
	LocalResourceAccessReviews(namespace string) LocalResourceAccessReviewInterface
}

// LocalResourceAccessReviewInterface has methods to work with LocalResourceAccessReview resources.
type LocalResourceAccessReviewInterface interface {
	Create(*v1.LocalResourceAccessReview) (*v1.LocalResourceAccessReview, error)
	LocalResourceAccessReviewExpansion
}

// localResourceAccessReviews implements LocalResourceAccessReviewInterface
type localResourceAccessReviews struct {
	client rest.Interface
	ns     string
}

// newLocalResourceAccessReviews returns a LocalResourceAccessReviews
func newLocalResourceAccessReviews(c *AuthorizationV1Client, namespace string) *localResourceAccessReviews {
	return &localResourceAccessReviews{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a localResourceAccessReview and creates it.  Returns the server's representation of the localResourceAccessReview, and an error, if there is any.
func (c *localResourceAccessReviews) Create(localResourceAccessReview *v1.LocalResourceAccessReview) (result *v1.LocalResourceAccessReview, err error) {
	result = &v1.LocalResourceAccessReview{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("localresourceaccessreviews").
		Body(localResourceAccessReview).
		Do().
		Into(result)
	return
}
