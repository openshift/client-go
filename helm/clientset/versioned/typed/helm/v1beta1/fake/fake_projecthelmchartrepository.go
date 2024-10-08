// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1beta1 "github.com/openshift/api/helm/v1beta1"
	helmv1beta1 "github.com/openshift/client-go/helm/applyconfigurations/helm/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeProjectHelmChartRepositories implements ProjectHelmChartRepositoryInterface
type FakeProjectHelmChartRepositories struct {
	Fake *FakeHelmV1beta1
	ns   string
}

var projecthelmchartrepositoriesResource = v1beta1.SchemeGroupVersion.WithResource("projecthelmchartrepositories")

var projecthelmchartrepositoriesKind = v1beta1.SchemeGroupVersion.WithKind("ProjectHelmChartRepository")

// Get takes name of the projectHelmChartRepository, and returns the corresponding projectHelmChartRepository object, and an error if there is any.
func (c *FakeProjectHelmChartRepositories) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.ProjectHelmChartRepository, err error) {
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(projecthelmchartrepositoriesResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}

// List takes label and field selectors, and returns the list of ProjectHelmChartRepositories that match those selectors.
func (c *FakeProjectHelmChartRepositories) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ProjectHelmChartRepositoryList, err error) {
	emptyResult := &v1beta1.ProjectHelmChartRepositoryList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(projecthelmchartrepositoriesResource, projecthelmchartrepositoriesKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ProjectHelmChartRepositoryList{ListMeta: obj.(*v1beta1.ProjectHelmChartRepositoryList).ListMeta}
	for _, item := range obj.(*v1beta1.ProjectHelmChartRepositoryList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested projectHelmChartRepositories.
func (c *FakeProjectHelmChartRepositories) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(projecthelmchartrepositoriesResource, c.ns, opts))

}

// Create takes the representation of a projectHelmChartRepository and creates it.  Returns the server's representation of the projectHelmChartRepository, and an error, if there is any.
func (c *FakeProjectHelmChartRepositories) Create(ctx context.Context, projectHelmChartRepository *v1beta1.ProjectHelmChartRepository, opts v1.CreateOptions) (result *v1beta1.ProjectHelmChartRepository, err error) {
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(projecthelmchartrepositoriesResource, c.ns, projectHelmChartRepository, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}

// Update takes the representation of a projectHelmChartRepository and updates it. Returns the server's representation of the projectHelmChartRepository, and an error, if there is any.
func (c *FakeProjectHelmChartRepositories) Update(ctx context.Context, projectHelmChartRepository *v1beta1.ProjectHelmChartRepository, opts v1.UpdateOptions) (result *v1beta1.ProjectHelmChartRepository, err error) {
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(projecthelmchartrepositoriesResource, c.ns, projectHelmChartRepository, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeProjectHelmChartRepositories) UpdateStatus(ctx context.Context, projectHelmChartRepository *v1beta1.ProjectHelmChartRepository, opts v1.UpdateOptions) (result *v1beta1.ProjectHelmChartRepository, err error) {
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(projecthelmchartrepositoriesResource, "status", c.ns, projectHelmChartRepository, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}

// Delete takes name of the projectHelmChartRepository and deletes it. Returns an error if one occurs.
func (c *FakeProjectHelmChartRepositories) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(projecthelmchartrepositoriesResource, c.ns, name, opts), &v1beta1.ProjectHelmChartRepository{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeProjectHelmChartRepositories) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(projecthelmchartrepositoriesResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.ProjectHelmChartRepositoryList{})
	return err
}

// Patch applies the patch and returns the patched projectHelmChartRepository.
func (c *FakeProjectHelmChartRepositories) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ProjectHelmChartRepository, err error) {
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(projecthelmchartrepositoriesResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied projectHelmChartRepository.
func (c *FakeProjectHelmChartRepositories) Apply(ctx context.Context, projectHelmChartRepository *helmv1beta1.ProjectHelmChartRepositoryApplyConfiguration, opts v1.ApplyOptions) (result *v1beta1.ProjectHelmChartRepository, err error) {
	if projectHelmChartRepository == nil {
		return nil, fmt.Errorf("projectHelmChartRepository provided to Apply must not be nil")
	}
	data, err := json.Marshal(projectHelmChartRepository)
	if err != nil {
		return nil, err
	}
	name := projectHelmChartRepository.Name
	if name == nil {
		return nil, fmt.Errorf("projectHelmChartRepository.Name must be provided to Apply")
	}
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(projecthelmchartrepositoriesResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeProjectHelmChartRepositories) ApplyStatus(ctx context.Context, projectHelmChartRepository *helmv1beta1.ProjectHelmChartRepositoryApplyConfiguration, opts v1.ApplyOptions) (result *v1beta1.ProjectHelmChartRepository, err error) {
	if projectHelmChartRepository == nil {
		return nil, fmt.Errorf("projectHelmChartRepository provided to Apply must not be nil")
	}
	data, err := json.Marshal(projectHelmChartRepository)
	if err != nil {
		return nil, err
	}
	name := projectHelmChartRepository.Name
	if name == nil {
		return nil, fmt.Errorf("projectHelmChartRepository.Name must be provided to Apply")
	}
	emptyResult := &v1beta1.ProjectHelmChartRepository{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(projecthelmchartrepositoriesResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ProjectHelmChartRepository), err
}
