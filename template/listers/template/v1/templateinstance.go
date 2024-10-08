// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// TemplateInstanceLister helps list TemplateInstances.
// All objects returned here must be treated as read-only.
type TemplateInstanceLister interface {
	// List lists all TemplateInstances in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.TemplateInstance, err error)
	// TemplateInstances returns an object that can list and get TemplateInstances.
	TemplateInstances(namespace string) TemplateInstanceNamespaceLister
	TemplateInstanceListerExpansion
}

// templateInstanceLister implements the TemplateInstanceLister interface.
type templateInstanceLister struct {
	listers.ResourceIndexer[*v1.TemplateInstance]
}

// NewTemplateInstanceLister returns a new TemplateInstanceLister.
func NewTemplateInstanceLister(indexer cache.Indexer) TemplateInstanceLister {
	return &templateInstanceLister{listers.New[*v1.TemplateInstance](indexer, v1.Resource("templateinstance"))}
}

// TemplateInstances returns an object that can list and get TemplateInstances.
func (s *templateInstanceLister) TemplateInstances(namespace string) TemplateInstanceNamespaceLister {
	return templateInstanceNamespaceLister{listers.NewNamespaced[*v1.TemplateInstance](s.ResourceIndexer, namespace)}
}

// TemplateInstanceNamespaceLister helps list and get TemplateInstances.
// All objects returned here must be treated as read-only.
type TemplateInstanceNamespaceLister interface {
	// List lists all TemplateInstances in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.TemplateInstance, err error)
	// Get retrieves the TemplateInstance from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.TemplateInstance, error)
	TemplateInstanceNamespaceListerExpansion
}

// templateInstanceNamespaceLister implements the TemplateInstanceNamespaceLister
// interface.
type templateInstanceNamespaceLister struct {
	listers.ResourceIndexer[*v1.TemplateInstance]
}
