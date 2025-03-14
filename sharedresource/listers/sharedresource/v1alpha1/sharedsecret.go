// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	sharedresourcev1alpha1 "github.com/openshift/api/sharedresource/v1alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// SharedSecretLister helps list SharedSecrets.
// All objects returned here must be treated as read-only.
type SharedSecretLister interface {
	// List lists all SharedSecrets in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*sharedresourcev1alpha1.SharedSecret, err error)
	// Get retrieves the SharedSecret from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*sharedresourcev1alpha1.SharedSecret, error)
	SharedSecretListerExpansion
}

// sharedSecretLister implements the SharedSecretLister interface.
type sharedSecretLister struct {
	listers.ResourceIndexer[*sharedresourcev1alpha1.SharedSecret]
}

// NewSharedSecretLister returns a new SharedSecretLister.
func NewSharedSecretLister(indexer cache.Indexer) SharedSecretLister {
	return &sharedSecretLister{listers.New[*sharedresourcev1alpha1.SharedSecret](indexer, sharedresourcev1alpha1.Resource("sharedsecret"))}
}
