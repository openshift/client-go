package v1

import (
	v1 "github.com/openshift/origin/pkg/authorization/apis/authorization/v1"
	rest "k8s.io/client-go/rest"
)

// ResourceAccessReviewsGetter has a method to return a ResourceAccessReviewInterface.
// A group's client should implement this interface.
type ResourceAccessReviewsGetter interface {
	ResourceAccessReviews() ResourceAccessReviewInterface
}

// ResourceAccessReviewInterface has methods to work with ResourceAccessReview resources.
type ResourceAccessReviewInterface interface {
	Create(*v1.ResourceAccessReview) (*v1.ResourceAccessReview, error)
	ResourceAccessReviewExpansion
}

// resourceAccessReviews implements ResourceAccessReviewInterface
type resourceAccessReviews struct {
	client rest.Interface
}

// newResourceAccessReviews returns a ResourceAccessReviews
func newResourceAccessReviews(c *AuthorizationV1Client) *resourceAccessReviews {
	return &resourceAccessReviews{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a resourceAccessReview and creates it.  Returns the server's representation of the resourceAccessReview, and an error, if there is any.
func (c *resourceAccessReviews) Create(resourceAccessReview *v1.ResourceAccessReview) (result *v1.ResourceAccessReview, err error) {
	result = &v1.ResourceAccessReview{}
	err = c.client.Post().
		Resource("resourceaccessreviews").
		Body(resourceAccessReview).
		Do().
		Into(result)
	return
}
