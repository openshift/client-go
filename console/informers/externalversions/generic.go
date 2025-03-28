// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	fmt "fmt"

	v1 "github.com/openshift/api/console/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=console.openshift.io, Version=v1
	case v1.SchemeGroupVersion.WithResource("consoleclidownloads"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleCLIDownloads().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consoleexternalloglinks"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleExternalLogLinks().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consolelinks"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleLinks().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consolenotifications"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleNotifications().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consoleplugins"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsolePlugins().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consolequickstarts"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleQuickStarts().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consolesamples"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleSamples().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("consoleyamlsamples"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Console().V1().ConsoleYAMLSamples().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
