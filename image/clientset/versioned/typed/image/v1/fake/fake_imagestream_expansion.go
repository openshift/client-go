package fake

import (
	image_v1 "github.com/openshift/api/image/v1"
	testing "k8s.io/client-go/testing"
)

// Layers takes name of the imageStream, and returns the corresponding image stream layers object, and an error if there is any.
func (c *FakeImageStreams) Layers(name string) (result *image_v1.ImageStreamLayers, err error) {
	obj, err := c.Fake.Invokes(testing.NewGetSubresourceAction(imagestreamsResource, c.ns, "layers", name), &image_v1.ImageStreamLayers{})

	if obj == nil {
		return nil, err
	}
	return obj.(*image_v1.ImageStreamLayers), err
}
