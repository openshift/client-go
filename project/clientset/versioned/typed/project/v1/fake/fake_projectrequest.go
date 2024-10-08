// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1 "github.com/openshift/api/project/v1"
	projectv1 "github.com/openshift/client-go/project/applyconfigurations/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	testing "k8s.io/client-go/testing"
)

// FakeProjectRequests implements ProjectRequestInterface
type FakeProjectRequests struct {
	Fake *FakeProjectV1
}

var projectrequestsResource = v1.SchemeGroupVersion.WithResource("projectrequests")

var projectrequestsKind = v1.SchemeGroupVersion.WithKind("ProjectRequest")

// Apply takes the given apply declarative configuration, applies it and returns the applied projectRequest.
func (c *FakeProjectRequests) Apply(ctx context.Context, projectRequest *projectv1.ProjectRequestApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ProjectRequest, err error) {
	if projectRequest == nil {
		return nil, fmt.Errorf("projectRequest provided to Apply must not be nil")
	}
	data, err := json.Marshal(projectRequest)
	if err != nil {
		return nil, err
	}
	name := projectRequest.Name
	if name == nil {
		return nil, fmt.Errorf("projectRequest.Name must be provided to Apply")
	}
	emptyResult := &v1.ProjectRequest{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(projectrequestsResource, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ProjectRequest), err
}

// Create takes the representation of a projectRequest and creates it.  Returns the server's representation of the project, and an error, if there is any.
func (c *FakeProjectRequests) Create(ctx context.Context, projectRequest *v1.ProjectRequest, opts metav1.CreateOptions) (result *v1.Project, err error) {
	emptyResult := &v1.Project{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(projectrequestsResource, projectRequest, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.Project), err
}
