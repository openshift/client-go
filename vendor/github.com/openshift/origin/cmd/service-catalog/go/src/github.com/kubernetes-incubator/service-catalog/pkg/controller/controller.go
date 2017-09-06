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

package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1alpha1"
	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1alpha1"
	listers "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1alpha1"
)

const (
	// maxRetries is the number of times a resource add/update will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a resource is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
	//
	pollingStartInterval      = 1 * time.Second
	pollingMaxBackoffDuration = 1 * time.Hour
)

// NewController returns a new Open Service Broker catalog controller.
func NewController(
	kubeClient kubernetes.Interface,
	serviceCatalogClient servicecatalogclientset.ServicecatalogV1alpha1Interface,
	brokerInformer informers.BrokerInformer,
	serviceClassInformer informers.ServiceClassInformer,
	instanceInformer informers.InstanceInformer,
	bindingInformer informers.BindingInformer,
	brokerClientCreateFunc osb.CreateFunc,
	brokerRelistInterval time.Duration,
	osbAPIPreferredVersion string,
	recorder record.EventRecorder,
) (Controller, error) {
	controller := &controller{
		kubeClient:             kubeClient,
		serviceCatalogClient:   serviceCatalogClient,
		brokerClientCreateFunc: brokerClientCreateFunc,
		brokerRelistInterval:   brokerRelistInterval,
		OSBAPIPreferredVersion: osbAPIPreferredVersion,
		recorder:               recorder,
		brokerQueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "broker"),
		serviceClassQueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-class"),
		instanceQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "instance"),
		bindingQueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "binding"),
		pollingQueue:           workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(pollingStartInterval, pollingMaxBackoffDuration), "poller"),
	}

	brokerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.brokerAdd,
		UpdateFunc: controller.brokerUpdate,
		DeleteFunc: controller.brokerDelete,
	})
	controller.brokerLister = brokerInformer.Lister()

	serviceClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.serviceClassAdd,
		UpdateFunc: controller.serviceClassUpdate,
		DeleteFunc: controller.serviceClassDelete,
	})
	controller.serviceClassLister = serviceClassInformer.Lister()

	instanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.instanceAdd,
		UpdateFunc: controller.instanceUpdate,
		DeleteFunc: controller.instanceDelete,
	})
	controller.instanceLister = instanceInformer.Lister()

	bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.bindingAdd,
		UpdateFunc: controller.bindingUpdate,
		DeleteFunc: controller.bindingDelete,
	})
	controller.bindingLister = bindingInformer.Lister()

	return controller, nil
}

// Controller describes a controller that backs the service catalog API for
// Open Service Broker compliant Brokers.
type Controller interface {
	// Run runs the controller until the given stop channel can be read from.
	// workers specifies the number of goroutines, per resource, processing work
	// from the resource workqueues
	Run(workers int, stopCh <-chan struct{})
}

// controller is a concrete Controller.
type controller struct {
	kubeClient             kubernetes.Interface
	serviceCatalogClient   servicecatalogclientset.ServicecatalogV1alpha1Interface
	brokerClientCreateFunc osb.CreateFunc
	brokerLister           listers.BrokerLister
	serviceClassLister     listers.ServiceClassLister
	instanceLister         listers.InstanceLister
	bindingLister          listers.BindingLister
	brokerRelistInterval   time.Duration
	OSBAPIPreferredVersion string
	recorder               record.EventRecorder
	brokerQueue            workqueue.RateLimitingInterface
	serviceClassQueue      workqueue.RateLimitingInterface
	instanceQueue          workqueue.RateLimitingInterface
	bindingQueue           workqueue.RateLimitingInterface
	// pollingQueue is separate from instanceQueue because we want
	// it to have different backoff / timeout characteristics from
	//  a reconciling of an instance.
	// TODO(vaikas): get rid of two queues per instance.
	pollingQueue workqueue.RateLimitingInterface
}

// Run runs the controller until the given stop channel can be read from.
func (c *controller) Run(workers int, stopCh <-chan struct{}) {
	defer runtimeutil.HandleCrash()

	glog.Info("Starting service-catalog controller")

	for i := 0; i < workers; i++ {
		go wait.Until(worker(c.brokerQueue, "Broker", maxRetries, c.reconcileBrokerKey), time.Second, stopCh)
		go wait.Until(worker(c.serviceClassQueue, "ServiceClass", maxRetries, c.reconcileServiceClassKey), time.Second, stopCh)
		go wait.Until(worker(c.instanceQueue, "Instance", maxRetries, c.reconcileInstanceKey), time.Second, stopCh)
		go wait.Until(worker(c.bindingQueue, "Binding", maxRetries, c.reconcileBindingKey), time.Second, stopCh)
		go wait.Until(worker(c.pollingQueue, "Poller", maxRetries, c.requeueInstanceForPoll), time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Shutting down service-catalog controller")

	c.brokerQueue.ShutDown()
	c.serviceClassQueue.ShutDown()
	c.instanceQueue.ShutDown()
	c.bindingQueue.ShutDown()
	c.pollingQueue.ShutDown()
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// If reconciler returns an error, requeue the item up to maxRetries before giving up.
// It enforces that the reconciler is never invoked concurrently with the same key.
func worker(queue workqueue.RateLimitingInterface, resourceType string, maxRetries int, reconciler func(key string) error) func() {
	return func() {
		exit := false
		for !exit {
			exit = func() bool {
				key, quit := queue.Get()
				if quit {
					return true
				}
				defer queue.Done(key)

				err := reconciler(key.(string))
				if err == nil {
					queue.Forget(key)
					return false
				}

				if queue.NumRequeues(key) < maxRetries {
					glog.V(4).Infof("Error syncing %s %v: %v", resourceType, key, err)
					queue.AddRateLimited(key)
					return false
				}

				glog.V(4).Infof("Dropping %s %q out of the queue: %v", resourceType, key, err)
				queue.Forget(key)
				return false
			}()
		}
	}
}

// getServiceClassPlanAndBroker is a sequence of operations that's done in couple of
// places so this method fetches the Service Class, Service Plan and creates
// a brokerClient to use for that method given an Instance.
func (c *controller) getServiceClassPlanAndBroker(instance *v1alpha1.Instance) (*v1alpha1.ServiceClass, *v1alpha1.ServicePlan, string, osb.Client, error) {
	serviceClass, err := c.serviceClassLister.Get(instance.Spec.ServiceClassName)
	if err != nil {
		s := fmt.Sprintf("Instance \"%s/%s\" references a non-existent ServiceClass %q", instance.Namespace, instance.Name, instance.Spec.ServiceClassName)
		glog.Info(s)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			errorNonexistentServiceClassReason,
			"The instance references a ServiceClass that does not exist. "+s,
		)
		c.recorder.Event(instance, api.EventTypeWarning, errorNonexistentServiceClassReason, s)
		return nil, nil, "", nil, err
	}

	servicePlan := findServicePlan(instance.Spec.PlanName, serviceClass.Plans)
	if servicePlan == nil {
		s := fmt.Sprintf("Instance \"%s/%s\" references a non-existent ServicePlan %q on ServiceClass %q", instance.Namespace, instance.Name, instance.Spec.PlanName, serviceClass.Name)
		glog.Warning(s)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentServicePlan",
			"The instance references a ServicePlan that does not exist. "+s,
		)
		c.recorder.Event(instance, api.EventTypeWarning, errorNonexistentServicePlanReason, s)
		return nil, nil, "", nil, fmt.Errorf(s)
	}

	broker, err := c.brokerLister.Get(serviceClass.BrokerName)
	if err != nil {
		s := fmt.Sprintf("Instance \"%s/%s\" references a non-existent broker %q", instance.Namespace, instance.Name, serviceClass.BrokerName)
		glog.Warning(s)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			errorNonexistentBrokerReason,
			"The instance references a Broker that does not exist. "+s,
		)
		c.recorder.Event(instance, api.EventTypeWarning, errorNonexistentBrokerReason, s)
		return nil, nil, "", nil, err
	}

	authConfig, err := getAuthCredentialsFromBroker(c.kubeClient, broker)
	if err != nil {
		s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
		glog.Info(s)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			errorAuthCredentialsReason,
			"Error getting auth credentials. "+s,
		)
		c.recorder.Event(instance, api.EventTypeWarning, errorAuthCredentialsReason, s)
		return nil, nil, "", nil, err
	}

	clientConfig := osb.DefaultClientConfiguration()
	clientConfig.Name = broker.Name
	clientConfig.URL = broker.Spec.URL
	clientConfig.AuthConfig = authConfig
	clientConfig.EnableAlphaFeatures = true
	clientConfig.Insecure = true

	glog.V(4).Infof("Creating client for Broker %v, URL: %v", broker.Name, broker.Spec.URL)
	brokerClient, err := c.brokerClientCreateFunc(clientConfig)
	if err != nil {
		return nil, nil, "", nil, err
	}

	return serviceClass, servicePlan, broker.Name, brokerClient, nil
}

// getServiceClassPlanAndBrokerForBinding is a sequence of operations that's
// done to validate service plan, service class exist, and handles creating
// a brokerclient to use for a given Instance.
func (c *controller) getServiceClassPlanAndBrokerForBinding(instance *v1alpha1.Instance, binding *v1alpha1.Binding) (*v1alpha1.ServiceClass, *v1alpha1.ServicePlan, string, osb.Client, error) {
	serviceClass, err := c.serviceClassLister.Get(instance.Spec.ServiceClassName)
	if err != nil {
		s := fmt.Sprintf("Binding \"%s/%s\" references a non-existent ServiceClass %q", binding.Namespace, binding.Name, instance.Spec.ServiceClassName)
		glog.Warning(s)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			errorNonexistentServiceClassReason,
			"The binding references a ServiceClass that does not exist. "+s,
		)
		c.recorder.Event(binding, api.EventTypeWarning, "ReferencesNonexistentServiceClass", s)
		return nil, nil, "", nil, err
	}

	servicePlan := findServicePlan(instance.Spec.PlanName, serviceClass.Plans)
	if servicePlan == nil {
		s := fmt.Sprintf("Instance \"%s/%s\" references a non-existent ServicePlan %q on ServiceClass %q", instance.Namespace, instance.Name, instance.Spec.PlanName, serviceClass.Name)
		glog.Warning(s)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			errorNonexistentServicePlanReason,
			"The Binding references an Instance which references ServicePlan that does not exist. "+s,
		)
		c.recorder.Event(binding, api.EventTypeWarning, errorNonexistentServicePlanReason, s)
		return nil, nil, "", nil, fmt.Errorf(s)
	}

	broker, err := c.brokerLister.Get(serviceClass.BrokerName)
	if err != nil {
		s := fmt.Sprintf("Binding \"%s/%s\" references a non-existent Broker %q", binding.Namespace, binding.Name, serviceClass.BrokerName)
		glog.Warning(s)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			errorNonexistentBrokerReason,
			"The binding references a Broker that does not exist. "+s,
		)
		c.recorder.Event(binding, api.EventTypeWarning, errorNonexistentBrokerReason, s)
		return nil, nil, "", nil, err
	}

	authConfig, err := getAuthCredentialsFromBroker(c.kubeClient, broker)
	if err != nil {
		s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
		glog.Warning(s)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			errorAuthCredentialsReason,
			"Error getting auth credentials. "+s,
		)
		c.recorder.Event(binding, api.EventTypeWarning, errorAuthCredentialsReason, s)
		return nil, nil, "", nil, err
	}

	clientConfig := osb.DefaultClientConfiguration()
	clientConfig.Name = broker.Name
	clientConfig.URL = broker.Spec.URL
	clientConfig.AuthConfig = authConfig
	clientConfig.EnableAlphaFeatures = true
	clientConfig.Insecure = true

	glog.V(4).Infof("Creating client for Broker %v, URL: %v", broker.Name, broker.Spec.URL)
	brokerClient, err := c.brokerClientCreateFunc(clientConfig)
	if err != nil {
		return nil, nil, "", nil, err
	}

	return serviceClass, servicePlan, broker.Name, brokerClient, nil
}

// Broker utility methods - move?

// getAuthCredentialsFromBroker returns the auth credentials, if any, or
// returns an error. If the AuthInfo field is nil, empty values are
// returned.
func getAuthCredentialsFromBroker(client kubernetes.Interface, broker *v1alpha1.Broker) (*osb.AuthConfig, error) {
	if broker.Spec.AuthInfo == nil {
		return nil, nil
	}

	authInfo := broker.Spec.AuthInfo
	if authInfo.Basic != nil {
		secretRef := authInfo.Basic.SecretRef
		secret, err := client.Core().Secrets(secretRef.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		basicAuthConfig, err := getBasicAuthConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BasicAuthConfig: basicAuthConfig,
		}, nil
	} else if authInfo.Bearer != nil {
		secretRef := authInfo.Bearer.SecretRef
		secret, err := client.Core().Secrets(secretRef.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		bearerConfig, err := getBearerConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BearerConfig: bearerConfig,
		}, nil
	} else if authInfo.BasicAuthSecret != nil {
		secretRef := authInfo.BasicAuthSecret
		secret, err := client.Core().Secrets(secretRef.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		basicAuthConfig, err := getBasicAuthConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BasicAuthConfig: basicAuthConfig,
		}, nil
	}
	return nil, fmt.Errorf("empty auth info or unsupported auth mode: %s", authInfo)
}

func getBasicAuthConfig(secret *apiv1.Secret) (*osb.BasicAuthConfig, error) {
	usernameBytes, ok := secret.Data["username"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain username")
	}

	passwordBytes, ok := secret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain password")
	}

	return &osb.BasicAuthConfig{
		Username: string(usernameBytes),
		Password: string(passwordBytes),
	}, nil
}

func getBearerConfig(secret *apiv1.Secret) (*osb.BearerConfig, error) {
	tokenBytes, ok := secret.Data["token"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain token")
	}

	return &osb.BearerConfig{
		Token: string(tokenBytes),
	}, nil
}

// convertCatalog converts a service broker catalog into an array of ServiceClasses
func convertCatalog(in *osb.CatalogResponse) ([]*v1alpha1.ServiceClass, error) {
	ret := make([]*v1alpha1.ServiceClass, len(in.Services))
	for i, svc := range in.Services {
		plans, err := convertServicePlans(svc.Plans)
		if err != nil {
			return nil, err
		}
		ret[i] = &v1alpha1.ServiceClass{
			Bindable:      svc.Bindable,
			Plans:         plans,
			PlanUpdatable: (svc.PlanUpdatable != nil && *svc.PlanUpdatable),
			ExternalID:    svc.ID,
			AlphaTags:     svc.Tags,
			Description:   svc.Description,
			AlphaRequires: svc.Requires,
		}

		if svc.Metadata != nil {
			metadata, err := json.Marshal(svc.Metadata)
			if err != nil {
				err = fmt.Errorf("Failed to marshal metadata\n%+v\n %v", svc.Metadata, err)
				glog.Error(err)
				return nil, err
			}
			ret[i].ExternalMetadata = &runtime.RawExtension{Raw: metadata}
		}

		ret[i].SetName(svc.Name)
	}
	return ret, nil
}

func convertServicePlans(plans []osb.Plan) ([]v1alpha1.ServicePlan, error) {
	ret := make([]v1alpha1.ServicePlan, len(plans))
	for i := range plans {
		ret[i] = v1alpha1.ServicePlan{
			Name:        plans[i].Name,
			ExternalID:  plans[i].ID,
			Free:        (plans[i].Free != nil && *plans[i].Free),
			Description: plans[i].Description,
		}

		if plans[i].Bindable != nil {
			b := *plans[i].Bindable
			ret[i].Bindable = &b
		}

		if plans[i].Metadata != nil {
			metadata, err := json.Marshal(plans[i].Metadata)
			if err != nil {
				err = fmt.Errorf("Failed to marshal metadata\n%+v\n %v", plans[i].Metadata, err)
				glog.Error(err)
				return nil, err
			}
			ret[i].ExternalMetadata = &runtime.RawExtension{Raw: metadata}
		}

		if schemas := plans[i].AlphaParameterSchemas; schemas != nil {
			if instanceSchemas := schemas.ServiceInstances; instanceSchemas != nil {
				if instanceCreateSchema := instanceSchemas.Create; instanceCreateSchema != nil && instanceCreateSchema.Parameters != nil {
					schema, err := json.Marshal(instanceCreateSchema.Parameters)
					if err != nil {
						err = fmt.Errorf("Failed to marshal instance create schema \n%+v\n %v", instanceCreateSchema.Parameters, err)
						glog.Error(err)
						return nil, err
					}
					ret[i].AlphaInstanceCreateParameterSchema = &runtime.RawExtension{Raw: schema}
				}
				if instanceUpdateSchema := instanceSchemas.Update; instanceUpdateSchema != nil && instanceUpdateSchema.Parameters != nil {
					schema, err := json.Marshal(instanceUpdateSchema.Parameters)
					if err != nil {
						err = fmt.Errorf("Failed to marshal instance update schema \n%+v\n %v", instanceUpdateSchema.Parameters, err)
						glog.Error(err)
						return nil, err
					}
					ret[i].AlphaInstanceUpdateParameterSchema = &runtime.RawExtension{Raw: schema}
				}
			}
			if bindingSchemas := schemas.ServiceBindings; bindingSchemas != nil {
				if bindingCreateSchema := bindingSchemas.Create; bindingCreateSchema != nil && bindingCreateSchema.Parameters != nil {
					schema, err := json.Marshal(bindingCreateSchema.Parameters)
					if err != nil {
						err = fmt.Errorf("Failed to marshal binding create schema \n%+v\n %v", bindingCreateSchema.Parameters, err)
						glog.Error(err)
						return nil, err
					}
					ret[i].AlphaBindingCreateParameterSchema = &runtime.RawExtension{Raw: schema}
				}
			}
		}

	}
	return ret, nil
}

// isInstanceReady returns whether the given instance has a ready condition
// with status true.
func isInstanceReady(instance *v1alpha1.Instance) bool {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == v1alpha1.InstanceConditionReady {
			return cond.Status == v1alpha1.ConditionTrue
		}
	}

	return false
}

// TODO (nilebox): The controllerRef methods below are merged into apimachinery and will be released in 1.8:
// https://github.com/kubernetes/kubernetes/pull/48319
// Remove them after 1.8 is released and Service Catalog is migrated to it

// IsControlledBy checks if the given object has a controller ownerReference set to the given owner
func IsControlledBy(obj metav1.Object, owner metav1.Object) bool {
	ref := GetControllerOf(obj)
	if ref == nil {
		return false
	}
	return ref.UID == owner.GetUID()
}

// GetControllerOf returns the controllerRef if controllee has a controller,
// otherwise returns nil.
func GetControllerOf(controllee metav1.Object) *metav1.OwnerReference {
	for _, ref := range controllee.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller == true {
			return &ref
		}
	}
	return nil
}

// NewControllerRef creates an OwnerReference pointing to the given owner.
func NewControllerRef(owner metav1.Object, gvk schema.GroupVersionKind) *metav1.OwnerReference {
	blockOwnerDeletion := true
	isController := true
	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}
