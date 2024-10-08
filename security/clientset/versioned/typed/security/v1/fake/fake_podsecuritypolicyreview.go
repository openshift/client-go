// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/openshift/api/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testing "k8s.io/client-go/testing"
)

// FakePodSecurityPolicyReviews implements PodSecurityPolicyReviewInterface
type FakePodSecurityPolicyReviews struct {
	Fake *FakeSecurityV1
	ns   string
}

var podsecuritypolicyreviewsResource = v1.SchemeGroupVersion.WithResource("podsecuritypolicyreviews")

var podsecuritypolicyreviewsKind = v1.SchemeGroupVersion.WithKind("PodSecurityPolicyReview")

// Create takes the representation of a podSecurityPolicyReview and creates it.  Returns the server's representation of the podSecurityPolicyReview, and an error, if there is any.
func (c *FakePodSecurityPolicyReviews) Create(ctx context.Context, podSecurityPolicyReview *v1.PodSecurityPolicyReview, opts metav1.CreateOptions) (result *v1.PodSecurityPolicyReview, err error) {
	emptyResult := &v1.PodSecurityPolicyReview{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(podsecuritypolicyreviewsResource, c.ns, podSecurityPolicyReview, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.PodSecurityPolicyReview), err
}
