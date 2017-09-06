package etcd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	kapi "k8s.io/kubernetes/pkg/api"

	userapi "github.com/openshift/origin/pkg/user/apis/user"
	"github.com/openshift/origin/pkg/user/registry/identity"
	"github.com/openshift/origin/pkg/util/restoptions"
)

// REST implements a RESTStorage for identites against etcd
type REST struct {
	*registry.Store
}

var _ rest.StandardStorage = &REST{}

// NewREST returns a RESTStorage object that will work against identites
func NewREST(optsGetter restoptions.Getter) (*REST, error) {
	store := &registry.Store{
		Copier:                   kapi.Scheme,
		NewFunc:                  func() runtime.Object { return &userapi.Identity{} },
		NewListFunc:              func() runtime.Object { return &userapi.IdentityList{} },
		PredicateFunc:            identity.Matcher,
		DefaultQualifiedResource: userapi.Resource("identities"),

		CreateStrategy: identity.Strategy,
		UpdateStrategy: identity.Strategy,
		DeleteStrategy: identity.Strategy,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: identity.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &REST{store}, nil
}
