package fake

import (
	project "github.com/openshift/origin/pkg/project/apis/project"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeProjects implements ProjectResourceInterface
type FakeProjects struct {
	Fake *FakeProject
}

var projectsResource = schema.GroupVersionResource{Group: "project.openshift.io", Version: "", Resource: "projects"}

var projectsKind = schema.GroupVersionKind{Group: "project.openshift.io", Version: "", Kind: "Project"}

// Get takes name of the projectResource, and returns the corresponding projectResource object, and an error if there is any.
func (c *FakeProjects) Get(name string, options v1.GetOptions) (result *project.Project, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(projectsResource, name), &project.Project{})
	if obj == nil {
		return nil, err
	}
	return obj.(*project.Project), err
}

// List takes label and field selectors, and returns the list of Projects that match those selectors.
func (c *FakeProjects) List(opts v1.ListOptions) (result *project.ProjectList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(projectsResource, projectsKind, opts), &project.ProjectList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &project.ProjectList{}
	for _, item := range obj.(*project.ProjectList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested projects.
func (c *FakeProjects) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(projectsResource, opts))
}

// Create takes the representation of a projectResource and creates it.  Returns the server's representation of the projectResource, and an error, if there is any.
func (c *FakeProjects) Create(projectResource *project.Project) (result *project.Project, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(projectsResource, projectResource), &project.Project{})
	if obj == nil {
		return nil, err
	}
	return obj.(*project.Project), err
}

// Update takes the representation of a projectResource and updates it. Returns the server's representation of the projectResource, and an error, if there is any.
func (c *FakeProjects) Update(projectResource *project.Project) (result *project.Project, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(projectsResource, projectResource), &project.Project{})
	if obj == nil {
		return nil, err
	}
	return obj.(*project.Project), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeProjects) UpdateStatus(projectResource *project.Project) (*project.Project, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(projectsResource, "status", projectResource), &project.Project{})
	if obj == nil {
		return nil, err
	}
	return obj.(*project.Project), err
}

// Delete takes name of the projectResource and deletes it. Returns an error if one occurs.
func (c *FakeProjects) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(projectsResource, name), &project.Project{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeProjects) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(projectsResource, listOptions)

	_, err := c.Fake.Invokes(action, &project.ProjectList{})
	return err
}

// Patch applies the patch and returns the patched projectResource.
func (c *FakeProjects) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *project.Project, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(projectsResource, name, data, subresources...), &project.Project{})
	if obj == nil {
		return nil, err
	}
	return obj.(*project.Project), err
}
