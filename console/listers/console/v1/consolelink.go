// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/api/console/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// ConsoleLinkLister helps list ConsoleLinks.
// All objects returned here must be treated as read-only.
type ConsoleLinkLister interface {
	// List lists all ConsoleLinks in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.ConsoleLink, err error)
	// Get retrieves the ConsoleLink from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.ConsoleLink, error)
	ConsoleLinkListerExpansion
}

// consoleLinkLister implements the ConsoleLinkLister interface.
type consoleLinkLister struct {
	listers.ResourceIndexer[*v1.ConsoleLink]
}

// NewConsoleLinkLister returns a new ConsoleLinkLister.
func NewConsoleLinkLister(indexer cache.Indexer) ConsoleLinkLister {
	return &consoleLinkLister{listers.New[*v1.ConsoleLink](indexer, v1.Resource("consolelink"))}
}
