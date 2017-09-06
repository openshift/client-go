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
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	checksum "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/checksum/versioned/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/tools/cache"
)

func (c *controller) brokerAdd(obj interface{}) {
	// DeletionHandlingMetaNamespaceKeyFunc returns a unique key for the resource and
	// handles the special case where the resource is of DeletedFinalStateUnknown type, which
	// acts a place holder for resources that have been deleted from storage but the watch event
	// confirming the deletion has not yet arrived.
	// Generally, the key is "namespace/name" for namespaced-scoped resources and
	// just "name" for cluster scoped resources.
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.brokerQueue.Add(key)
}

func (c *controller) brokerUpdate(oldObj, newObj interface{}) {
	c.brokerAdd(newObj)
}

func (c *controller) brokerDelete(obj interface{}) {
	broker, ok := obj.(*v1alpha1.Broker)
	if broker == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for Broker %v; no further processing will occur", broker.Name)
}

// the Message strings have a terminating period and space so they can
// be easily combined with a follow on specific message.
const (
	errorFetchingCatalogReason            string = "ErrorFetchingCatalog"
	errorFetchingCatalogMessage           string = "Error fetching catalog. "
	errorSyncingCatalogReason             string = "ErrorSyncingCatalog"
	errorSyncingCatalogMessage            string = "Error syncing catalog from Broker. "
	errorWithParameters                   string = "ErrorWithParameters"
	errorListingServiceClassesReason      string = "ErrorListingServiceClasses"
	errorListingServiceClassesMessage     string = "Error listing service classes."
	errorDeletingServiceClassReason       string = "ErrorDeletingServiceClass"
	errorDeletingServiceClassMessage      string = "Error deleting service class."
	errorNonexistentServiceClassReason    string = "ReferencesNonexistentServiceClass"
	errorNonexistentServiceClassMessage   string = "ReferencesNonexistentServiceClass"
	errorNonexistentServicePlanReason     string = "ReferencesNonexistentServicePlan"
	errorNonexistentBrokerReason          string = "ReferencesNonexistentBroker"
	errorNonexistentInstanceReason        string = "ReferencesNonexistentInstance"
	errorAuthCredentialsReason            string = "ErrorGettingAuthCredentials"
	errorFindingNamespaceInstanceReason   string = "ErrorFindingNamespaceForInstance"
	errorProvisionCallFailedReason        string = "ProvisionCallFailed"
	errorErrorCallingProvisionReason      string = "ErrorCallingProvision"
	errorDeprovisionCalledReason          string = "DeprovisionCallFailed"
	errorBindCallReason                   string = "BindCallFailed"
	errorInjectingBindResultReason        string = "ErrorInjectingBindResult"
	errorEjectingBindReason               string = "ErrorEjectingBinding"
	errorEjectingBindMessage              string = "Error ejecting binding."
	errorUnbindCallReason                 string = "UnbindCallFailed"
	errorWithOngoingAsyncOperation        string = "ErrorAsyncOperationInProgress"
	errorWithOngoingAsyncOperationMessage string = "Another operation for this service instance is in progress. "
	errorNonbindableServiceClassReason    string = "ErrorNonbindableServiceClass"
	errorInstanceNotReadyReason           string = "ErrorInstanceNotReady"
	errorPollingLastOperationReason       string = "ErrorPollingLastOperation"

	successInjectedBindResultReason  string = "InjectedBindResult"
	successInjectedBindResultMessage string = "Injected bind result"
	successDeprovisionReason         string = "DeprovisionedSuccessfully"
	successDeprovisionMessage        string = "The instance was deprovisioned successfully"
	successProvisionReason           string = "ProvisionedSuccessfully"
	successProvisionMessage          string = "The instance was provisioned successfully"
	successFetchedCatalogReason      string = "FetchedCatalog"
	successFetchedCatalogMessage     string = "Successfully fetched catalog entries from broker."
	successBrokerDeletedReason       string = "DeletedSuccessfully"
	successBrokerDeletedMessage      string = "The broker %v was deleted successfully."
	successUnboundReason             string = "UnboundSuccessfully"
	asyncProvisioningReason          string = "Provisioning"
	asyncProvisioningMessage         string = "The instance is being provisioned asynchronously"
	asyncDeprovisioningReason        string = "Deprovisioning"
	asyncDeprovisioningMessage       string = "The instance is being deprovisioned asynchronously"
)

// shouldReconcileBroker determines whether a broker should be reconciled; it
// returns true unless the broker has a ready condition with status true and
// the controller's broker relist interval has not elapsed since the broker's
// ready condition became true.
func shouldReconcileBroker(broker *v1alpha1.Broker, now time.Time, relistInterval time.Duration) bool {
	brokerChecksum := checksum.BrokerSpecChecksum(broker.Spec)
	if broker.Status.Checksum != nil && brokerChecksum != *broker.Status.Checksum {
		// If the spec has changed, we should reconcile the broker.
		return true
	}
	if broker.DeletionTimestamp != nil || len(broker.Status.Conditions) == 0 {
		// If the deletion timestamp is set or the broker has no status
		// conditions, we should reconcile it.
		return true
	}

	// find the ready condition in the broker's status
	for _, condition := range broker.Status.Conditions {
		if condition.Type == v1alpha1.BrokerConditionReady {
			// The broker has a ready condition

			if condition.Status == v1alpha1.ConditionTrue {
				// The broker's ready condition has status true, meaning that
				// at some point, we successfully listed the broker's catalog.
				// We should reconcile the broker (relist the broker's
				// catalog) if it has been longer than the configured relist
				// interval since the broker's ready condition became true.
				return now.After(condition.LastTransitionTime.Add(relistInterval))
			}

			// The broker's ready condition wasn't true; we should try to re-
			// list the broker.
			return true
		}
	}

	// The broker didn't have a ready condition; we should reconcile it.
	return true
}

func (c *controller) reconcileBrokerKey(key string) error {
	broker, err := c.brokerLister.Get(key)
	if errors.IsNotFound(err) {
		glog.Infof("Not doing work for Broker %v because it has been deleted", key)
		return nil
	}
	if err != nil {
		glog.Infof("Unable to retrieve Broker %v from store: %v", key, err)
		return err
	}

	return c.reconcileBroker(broker)
}

// reconcileBroker is the control-loop that reconciles a Broker. An
// error is returned to indicate that the binding has not been fully
// processed and should be resubmitted at a later time.
func (c *controller) reconcileBroker(broker *v1alpha1.Broker) error {
	glog.V(4).Infof("Processing Broker %v", broker.Name)

	// If the broker's ready condition is true and the relist interval has not
	// elapsed, do not reconcile it.
	if !shouldReconcileBroker(broker, time.Now(), c.brokerRelistInterval) {
		glog.V(10).Infof(
			"Not processing Broker %v because relist interval has not elapsed since the broker became ready",
			broker.Name,
		)
		return nil
	}

	if broker.DeletionTimestamp == nil { // Add or update
		authConfig, err := getAuthCredentialsFromBroker(c.kubeClient, broker)
		if err != nil {
			s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
			glog.Info(s)
			c.recorder.Event(broker, api.EventTypeWarning, errorAuthCredentialsReason, s)
			c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s)
			return err
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
			s := fmt.Sprintf("Error creating client for broker %q: %s", broker.Name, err)
			glog.Info(s)
			c.recorder.Event(broker, api.EventTypeWarning, errorAuthCredentialsReason, s)
			c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s)
			return err
		}

		glog.V(4).Infof("Adding/Updating Broker %v", broker.Name)
		brokerCatalog, err := brokerClient.GetCatalog()
		if err != nil {
			s := fmt.Sprintf("Error getting broker catalog for broker %q: %s", broker.Name, err)
			glog.Warning(s)
			c.recorder.Eventf(broker, api.EventTypeWarning, errorFetchingCatalogReason, s)
			c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionFalse, errorFetchingCatalogReason,
				errorFetchingCatalogMessage+s)
			return err
		}
		glog.V(5).Infof("Successfully fetched %v catalog entries for Broker %v", len(brokerCatalog.Services), broker.Name)

		glog.V(4).Infof("Converting catalog response for Broker %v into service-catalog API", broker.Name)
		catalog, err := convertCatalog(brokerCatalog)
		if err != nil {
			s := fmt.Sprintf("Error converting catalog payload for broker %q to service-catalog API: %s", broker.Name, err)
			glog.Warning(s)
			c.recorder.Eventf(broker, api.EventTypeWarning, errorSyncingCatalogReason, s)
			c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s)
			return err
		}
		glog.V(5).Infof("Successfully converted catalog payload from Broker %v to service-catalog API", broker.Name)

		if len(catalog) == 0 {
			s := fmt.Sprintf("Error getting catalog payload for broker %q; received zero services; at least one service is required", broker.Name)
			glog.Warning(s)
			c.recorder.Eventf(broker, api.EventTypeWarning, errorSyncingCatalogReason, s)
			c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s)
			return stderrors.New(s)
		}

		for _, serviceClass := range catalog {
			glog.V(4).Infof("Reconciling serviceClass %v (broker %v)", serviceClass.Name, broker.Name)
			if err := c.reconcileServiceClassFromBrokerCatalog(broker, serviceClass); err != nil {
				s := fmt.Sprintf(
					"Error reconciling serviceClass %q (broker %q): %s",
					serviceClass.Name,
					broker.Name,
					err,
				)
				glog.Warning(s)
				c.recorder.Eventf(broker, api.EventTypeWarning, errorSyncingCatalogReason, s)
				c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s)
				return err
			}

			glog.V(5).Infof("Reconciled serviceClass %v (broker %v)", serviceClass.Name, broker.Name)
		}

		c.updateBrokerCondition(broker, v1alpha1.BrokerConditionReady, v1alpha1.ConditionTrue, successFetchedCatalogReason, successFetchedCatalogMessage)
		c.recorder.Event(broker, api.EventTypeNormal, successFetchedCatalogReason, successFetchedCatalogMessage)
		return nil
	}

	// All updates not having a DeletingTimestamp will have been handled above
	// and returned early. If we reach this point, we're dealing with an update
	// that's actually a soft delete-- i.e. we have some finalization to do.
	if finalizers := sets.NewString(broker.Finalizers...); finalizers.Has(v1alpha1.FinalizerServiceCatalog) {
		glog.V(4).Infof("Finalizing Broker %v", broker.Name)

		// Get ALL ServiceClasses. Remove those that reference this Broker.
		svcClasses, err := c.serviceClassLister.List(labels.Everything())
		if err != nil {
			c.updateBrokerCondition(
				broker,
				v1alpha1.BrokerConditionReady,
				v1alpha1.ConditionUnknown,
				errorListingServiceClassesReason,
				errorListingServiceClassesMessage,
			)
			c.recorder.Eventf(broker, api.EventTypeWarning, errorListingServiceClassesReason, "%v %v", errorListingServiceClassesMessage, err)
			return err
		}

		// Delete ServiceClasses that are for THIS Broker.
		for _, svcClass := range svcClasses {
			if svcClass.BrokerName == broker.Name {
				err := c.serviceCatalogClient.ServiceClasses().Delete(svcClass.Name, &metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					s := fmt.Sprintf(
						"Error deleting ServiceClass %q (Broker %q): %s",
						svcClass.Name,
						broker.Name,
						err,
					)
					glog.Warning(s)
					c.updateBrokerCondition(
						broker,
						v1alpha1.BrokerConditionReady,
						v1alpha1.ConditionUnknown,
						errorDeletingServiceClassMessage,
						errorDeletingServiceClassReason+s,
					)
					c.recorder.Eventf(broker, api.EventTypeWarning, errorDeletingServiceClassReason, "%v %v", errorDeletingServiceClassMessage, s)
					return err
				}
			}
		}

		c.updateBrokerCondition(
			broker,
			v1alpha1.BrokerConditionReady,
			v1alpha1.ConditionFalse,
			successBrokerDeletedReason,
			"The broker was deleted successfully",
		)
		// Clear the finalizer
		finalizers.Delete(v1alpha1.FinalizerServiceCatalog)
		c.updateBrokerFinalizers(broker, finalizers.List())

		c.recorder.Eventf(broker, api.EventTypeNormal, successBrokerDeletedReason, successBrokerDeletedMessage, broker.Name)
		glog.V(5).Infof("Successfully deleted Broker %v", broker.Name)
		return nil
	}

	return nil
}

// reconcileServiceClassFromBrokerCatalog reconciles a ServiceClass after the
// Broker's catalog has been re-listed.
func (c *controller) reconcileServiceClassFromBrokerCatalog(broker *v1alpha1.Broker, serviceClass *v1alpha1.ServiceClass) error {
	serviceClass.BrokerName = broker.Name

	existingServiceClass, err := c.serviceClassLister.Get(serviceClass.Name)
	if errors.IsNotFound(err) {
		// An error returned from a lister Get call means that the object does
		// not exist.  Create a new ServiceClass.
		if _, err := c.serviceCatalogClient.ServiceClasses().Create(serviceClass); err != nil {
			glog.Errorf("Error creating serviceClass %v from Broker %v: %v", serviceClass.Name, broker.Name, err)
			return err
		}

		return nil
	} else if err != nil {
		glog.Errorf("Error getting serviceClass %v: %v", serviceClass.Name, err)
		return err
	}

	if existingServiceClass.BrokerName != broker.Name {
		errMsg := fmt.Sprintf("ServiceClass %q for Broker %q already exists for Broker %q", serviceClass.Name, broker.Name, existingServiceClass.BrokerName)
		glog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	if existingServiceClass.ExternalID != serviceClass.ExternalID {
		errMsg := fmt.Sprintf("ServiceClass %q already exists with OSB guid %q, received different guid %q", serviceClass.Name, existingServiceClass.ExternalID, serviceClass.ExternalID)
		glog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	glog.V(5).Infof("Found existing serviceClass %v; updating", serviceClass.Name)

	// There was an existing service class -- project the update onto it and
	// update it.
	clone, err := api.Scheme.DeepCopy(existingServiceClass)
	if err != nil {
		return err
	}

	toUpdate := clone.(*v1alpha1.ServiceClass)
	toUpdate.Bindable = serviceClass.Bindable
	toUpdate.Plans = serviceClass.Plans
	toUpdate.PlanUpdatable = serviceClass.PlanUpdatable
	toUpdate.AlphaTags = serviceClass.AlphaTags
	toUpdate.Description = serviceClass.Description
	toUpdate.AlphaRequires = serviceClass.AlphaRequires

	if _, err := c.serviceCatalogClient.ServiceClasses().Update(toUpdate); err != nil {
		glog.Errorf("Error updating serviceClass %v from Broker %v: %v", serviceClass.Name, broker.Name, err)
		return err
	}

	return nil
}

// updateBrokerCondition updates the ready condition for the given Broker
// with the given status, reason, and message.
func (c *controller) updateBrokerCondition(broker *v1alpha1.Broker, conditionType v1alpha1.BrokerConditionType, status v1alpha1.ConditionStatus, reason, message string) error {
	clone, err := api.Scheme.DeepCopy(broker)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.Broker)
	newCondition := v1alpha1.BrokerCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	t := time.Now()

	if len(broker.Status.Conditions) == 0 {
		glog.Infof("Setting lastTransitionTime for Broker %q condition %q to %v", broker.Name, conditionType, t)
		newCondition.LastTransitionTime = metav1.NewTime(t)
		toUpdate.Status.Conditions = []v1alpha1.BrokerCondition{newCondition}
	} else {
		for i, cond := range broker.Status.Conditions {
			if cond.Type == conditionType {
				if cond.Status != newCondition.Status {
					glog.Infof("Found status change for Broker %q condition %q: %q -> %q; setting lastTransitionTime to %v", broker.Name, conditionType, cond.Status, status, t)
					newCondition.LastTransitionTime = metav1.NewTime(t)
				} else {
					newCondition.LastTransitionTime = cond.LastTransitionTime
				}

				toUpdate.Status.Conditions[i] = newCondition
				break
			}
		}
	}

	glog.V(4).Infof("Updating ready condition for Broker %v to %v", broker.Name, status)
	_, err = c.serviceCatalogClient.Brokers().UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Error updating ready condition for Broker %v: %v", broker.Name, err)
	} else {
		glog.V(5).Infof("Updated ready condition for Broker %v to %v", broker.Name, status)
	}

	return err
}

// updateBrokerFinalizers updates the given finalizers for the given Broker.
func (c *controller) updateBrokerFinalizers(
	broker *v1alpha1.Broker,
	finalizers []string) error {

	// Get the latest version of the broker so that we can avoid conflicts
	// (since we have probably just updated the status of the broker and are
	// now removing the last finalizer).
	broker, err := c.serviceCatalogClient.Brokers().Get(broker.Name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Error getting Broker %v to finalize: %v", broker.Name, err)
	}

	clone, err := api.Scheme.DeepCopy(broker)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.Broker)

	toUpdate.Finalizers = finalizers

	logContext := fmt.Sprintf("finalizers for Broker %v to %v",
		broker.Name, finalizers)

	glog.V(4).Infof("Updating %v", logContext)
	_, err = c.serviceCatalogClient.Brokers().UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Error updating %v: %v", logContext, err)
	}
	return err
}
