package etcd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/registry/cachesize"

	securityapi "github.com/openshift/origin/pkg/security/apis/security"
	"github.com/openshift/origin/pkg/security/registry/securitycontextconstraints"
)

// REST implements a RESTStorage for security context constraints against etcd
type REST struct {
	*registry.Store
}

var _ rest.StandardStorage = &REST{}

// NewREST returns a RESTStorage object that will work against security context constraints objects.
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &registry.Store{
		Copier:      api.Scheme,
		NewFunc:     func() runtime.Object { return &securityapi.SecurityContextConstraints{} },
		NewListFunc: func() runtime.Object { return &securityapi.SecurityContextConstraintsList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*securityapi.SecurityContextConstraints).Name, nil
		},
		PredicateFunc:            securitycontextconstraints.Matcher,
		DefaultQualifiedResource: securityapi.Resource("securitycontextconstraints"),
		WatchCacheSize:           cachesize.GetWatchCacheSizeByResource("securitycontextconstraints"),

		CreateStrategy:      securitycontextconstraints.Strategy,
		UpdateStrategy:      securitycontextconstraints.Strategy,
		DeleteStrategy:      securitycontextconstraints.Strategy,
		ReturnDeletedObject: true,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: securitycontextconstraints.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}
	return &REST{store}
}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"scc"}
}
