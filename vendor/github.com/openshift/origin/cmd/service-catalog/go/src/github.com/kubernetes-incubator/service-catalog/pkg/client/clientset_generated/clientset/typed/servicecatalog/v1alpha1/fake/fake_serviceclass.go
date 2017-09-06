/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	v1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceClasses implements ServiceClassInterface
type FakeServiceClasses struct {
	Fake *FakeServicecatalogV1alpha1
}

var serviceclassesResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Resource: "serviceclasses"}

var serviceclassesKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Kind: "ServiceClass"}

func (c *FakeServiceClasses) Create(serviceClass *v1alpha1.ServiceClass) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(serviceclassesResource, serviceClass), &v1alpha1.ServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}

func (c *FakeServiceClasses) Update(serviceClass *v1alpha1.ServiceClass) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(serviceclassesResource, serviceClass), &v1alpha1.ServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}

func (c *FakeServiceClasses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(serviceclassesResource, name), &v1alpha1.ServiceClass{})
	return err
}

func (c *FakeServiceClasses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(serviceclassesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ServiceClassList{})
	return err
}

func (c *FakeServiceClasses) Get(name string, options v1.GetOptions) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(serviceclassesResource, name), &v1alpha1.ServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}

func (c *FakeServiceClasses) List(opts v1.ListOptions) (result *v1alpha1.ServiceClassList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(serviceclassesResource, serviceclassesKind, opts), &v1alpha1.ServiceClassList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ServiceClassList{}
	for _, item := range obj.(*v1alpha1.ServiceClassList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceClasses.
func (c *FakeServiceClasses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(serviceclassesResource, opts))
}

// Patch applies the patch and returns the patched serviceClass.
func (c *FakeServiceClasses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(serviceclassesResource, name, data, subresources...), &v1alpha1.ServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}
