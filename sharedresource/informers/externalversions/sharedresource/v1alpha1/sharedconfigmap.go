// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	context "context"
	time "time"

	apisharedresourcev1alpha1 "github.com/openshift/api/sharedresource/v1alpha1"
	versioned "github.com/openshift/client-go/sharedresource/clientset/versioned"
	internalinterfaces "github.com/openshift/client-go/sharedresource/informers/externalversions/internalinterfaces"
	sharedresourcev1alpha1 "github.com/openshift/client-go/sharedresource/listers/sharedresource/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SharedConfigMapInformer provides access to a shared informer and lister for
// SharedConfigMaps.
type SharedConfigMapInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() sharedresourcev1alpha1.SharedConfigMapLister
}

type sharedConfigMapInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewSharedConfigMapInformer constructs a new informer for SharedConfigMap type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSharedConfigMapInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSharedConfigMapInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredSharedConfigMapInformer constructs a new informer for SharedConfigMap type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSharedConfigMapInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SharedresourceV1alpha1().SharedConfigMaps().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SharedresourceV1alpha1().SharedConfigMaps().Watch(context.TODO(), options)
			},
		},
		&apisharedresourcev1alpha1.SharedConfigMap{},
		resyncPeriod,
		indexers,
	)
}

func (f *sharedConfigMapInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSharedConfigMapInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *sharedConfigMapInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apisharedresourcev1alpha1.SharedConfigMap{}, f.defaultInformer)
}

func (f *sharedConfigMapInformer) Lister() sharedresourcev1alpha1.SharedConfigMapLister {
	return sharedresourcev1alpha1.NewSharedConfigMapLister(f.Informer().GetIndexer())
}
