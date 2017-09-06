package fake

import (
	image "github.com/openshift/origin/pkg/image/apis/image"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeImageStreamMappings implements ImageStreamMappingInterface
type FakeImageStreamMappings struct {
	Fake *FakeImage
	ns   string
}

var imagestreammappingsResource = schema.GroupVersionResource{Group: "image.openshift.io", Version: "", Resource: "imagestreammappings"}

var imagestreammappingsKind = schema.GroupVersionKind{Group: "image.openshift.io", Version: "", Kind: "ImageStreamMapping"}

// Create takes the representation of a imageStreamMapping and creates it.  Returns the server's representation of the imageStreamMapping, and an error, if there is any.
func (c *FakeImageStreamMappings) Create(imageStreamMapping *image.ImageStreamMapping) (result *image.ImageStreamMapping, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(imagestreammappingsResource, c.ns, imageStreamMapping), &image.ImageStreamMapping{})

	if obj == nil {
		return nil, err
	}
	return obj.(*image.ImageStreamMapping), err
}
