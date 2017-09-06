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
	"errors"
	"reflect"
	"testing"
	"time"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	fakeosb "github.com/pmorie/go-open-service-broker-client/v2/fake"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"

	"strings"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	clientgotesting "k8s.io/client-go/testing"
)

// TestShouldReconcileBroker ensures that with the expected conditions the
// reconciler is reported as needing to run.
//
// The test cases are proving:
// - broker without ready condition will reconcile
// - broker with deletion timestamp set will reconcile
// - broker without ready condition, with status will reconcile
// - broker without ready condition, without status will reconcile
// - broker with status/ready, past relist interval will reconcile
// - broker with status/ready, within relist interval will NOT reconcile
// - broker with status/ready/checksum, will reconcile
func TestShouldReconcileBroker(t *testing.T) {
	// Anonymous struct fields:
	// name: short description of the test
	// broker: broker object to test
	// now: what time the interval is calculated with respect to interval
	// internal: the time that has elapsed since now
	// reconcile: whether or not the reconciler should run, the return of
	// shouldReconcileBroker
	cases := []struct {
		name      string
		broker    *v1alpha1.Broker
		now       time.Time
		interval  time.Duration
		reconcile bool
	}{
		{
			name:      "no status",
			broker:    getTestBroker(),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name: "deletionTimestamp set",
			broker: func() *v1alpha1.Broker {
				b := getTestBrokerWithStatus(v1alpha1.ConditionTrue)
				b.DeletionTimestamp = &metav1.Time{}
				return b
			}(),
			now:       time.Now(),
			interval:  3 * time.Hour,
			reconcile: true,
		},
		{
			name: "no ready condition",
			broker: func() *v1alpha1.Broker {
				b := getTestBroker()
				b.Status = v1alpha1.BrokerStatus{
					Conditions: []v1alpha1.BrokerCondition{
						{
							Type:   v1alpha1.BrokerConditionType("NotARealCondition"),
							Status: v1alpha1.ConditionTrue,
						},
					},
				}
				return b
			}(),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name:      "not ready",
			broker:    getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name: "ready, interval elapsed",
			broker: func() *v1alpha1.Broker {
				broker := getTestBrokerWithStatus(v1alpha1.ConditionTrue)
				return broker
			}(),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name: "ready, interval not elapsed",
			broker: func() *v1alpha1.Broker {
				broker := getTestBrokerWithStatus(v1alpha1.ConditionTrue)
				return broker
			}(),
			now:       time.Now(),
			interval:  3 * time.Hour,
			reconcile: false,
		},
		{
			name: "ready, interval not elapsed, checksum changed",
			broker: func() *v1alpha1.Broker {
				broker := getTestBrokerWithStatus(v1alpha1.ConditionTrue)
				cs := "22081-9471-471"
				broker.Status.Checksum = &cs
				return broker
			}(),
			now:       time.Now(),
			interval:  3 * time.Hour,
			reconcile: true,
		},
	}

	for _, tc := range cases {
		var ltt *time.Time
		if len(tc.broker.Status.Conditions) != 0 {
			ltt = &tc.broker.Status.Conditions[0].LastTransitionTime.Time
		}

		t.Logf("%v: now: %v, interval: %v, last transition time: %v", tc.name, tc.now, tc.interval, ltt)
		actual := shouldReconcileBroker(tc.broker, tc.now, tc.interval)

		if e, a := tc.reconcile, actual; e != a {
			t.Errorf("%v: unexpected result: expected %v, got %v", tc.name, e, a)
		}
	}
}

// TestReconcileBrokerExistingServiceClass verifies a simple, successful run
// of reconcileBroker().  This test will cause reconcileBroker() to fetch the
// catalog from the Broker, create a Service Class for the single service that
// it lists and reconcile the service class ensuring the name and id of the
// relisted service matches the existing entry and updates the service catalog.
func TestReconcileBrokerExistingServiceClass(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t, getTestCatalogConfig())

	testServiceClass := getTestServiceClass()
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	if err := testController.reconcileBroker(getTestBroker()); err != nil {
		t.Fatalf("This should not fail : %v", err)
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertGetCatalog(t, brokerActions[0])

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 2)

	// first action should be an update action for a service class
	assertUpdate(t, actions[0], testServiceClass)

	// second action should be an update action for broker status subresource
	updatedBroker := assertUpdateStatus(t, actions[1], getTestBroker())
	assertBrokerReadyTrue(t, updatedBroker)

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	assertNumberOfActions(t, kubeActions, 0)
}

// TestReconcileBrokerExistingServiceClassDifferentExternalID simulates catalog
// refresh where broker lists an existing service but there is a mismatch on the
// service class ID which should result in an error
func TestReconcileBrokerExistingServiceClassDifferentExternalID(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t, getTestCatalogConfig())

	testServiceClass := getTestServiceClass()
	testServiceClass.ExternalID = "notTheSame"
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	if err := testController.reconcileBroker(getTestBroker()); err == nil {
		t.Fatal("The same service class should not be allowed with a different ID")
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertGetCatalog(t, brokerActions[0])

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 1)

	updatedBroker := assertUpdateStatus(t, actions[0], getTestBroker())
	assertBrokerReadyFalse(t, updatedBroker)

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	assertNumberOfActions(t, kubeActions, 0)

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	expectedEvent := api.EventTypeWarning + " " + errorSyncingCatalogReason + ` Error reconciling serviceClass "test-serviceclass" (broker "test-broker"): ServiceClass "test-serviceclass" already exists with OSB guid "notTheSame", received different guid "SCGUID"`
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event; expected\n%v, got\n%v", e, a)
	}
}

// TestReconcileBrokerExistingServiceClassDifferentBroker simulates catalog
// refresh where broker lists a service which matches an existing, already
// cataloged service but the service points to a different Broker.  Results in an error.
func TestReconcileBrokerExistingServiceClassDifferentBroker(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t, getTestCatalogConfig())

	testServiceClass := getTestServiceClass()
	testServiceClass.BrokerName = "notTheSame"
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	if err := testController.reconcileBroker(getTestBroker()); err == nil {
		t.Fatal("The same service class should not belong to two different brokers.")
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertGetCatalog(t, brokerActions[0])

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 1)

	updatedBroker := assertUpdateStatus(t, actions[0], getTestBroker())
	assertBrokerReadyFalse(t, updatedBroker)

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	assertNumberOfActions(t, kubeActions, 0)

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	expectedEvent := api.EventTypeWarning + " " + errorSyncingCatalogReason + ` Error reconciling serviceClass "test-serviceclass" (broker "test-broker"): ServiceClass "test-serviceclass" for Broker "test-broker" already exists for Broker "notTheSame"`
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event; expected\n%v, got\n%v", e, a)
	}
}

// TestReconcileBrokerDelete simulates a broker reconciliation where broker was marked for deletion.
// Results in service class and broker both being deleted.
func TestReconcileBrokerDelete(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t, getTestCatalogConfig())

	testServiceClass := getTestServiceClass()
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	broker := getTestBroker()
	broker.DeletionTimestamp = &metav1.Time{}
	broker.Finalizers = []string{v1alpha1.FinalizerServiceCatalog}
	fakeCatalogClient.AddReactor("get", "brokers", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, broker, nil
	})

	err := testController.reconcileBroker(broker)
	if err != nil {
		t.Fatalf("This should not fail : %v", err)
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 0)

	// Verify no core kube actions occurred
	kubeActions := fakeKubeClient.Actions()
	assertNumberOfActions(t, kubeActions, 0)

	actions := fakeCatalogClient.Actions()
	// The four actions should be:
	// 0. Deleting the associated ServiceClass
	// 1. Updating the ready condition
	// 2. Getting the broker
	// 3. Removing the finalizer
	assertNumberOfActions(t, actions, 4)

	assertDelete(t, actions[0], testServiceClass)

	updatedBroker := assertUpdateStatus(t, actions[1], broker)
	assertBrokerReadyFalse(t, updatedBroker)

	assertGet(t, actions[2], broker)

	updatedBroker = assertUpdateStatus(t, actions[3], broker)
	assertEmptyFinalizers(t, updatedBroker)

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	expectedEvent := api.EventTypeNormal + " " + successBrokerDeletedReason + " " + "The broker test-broker was deleted successfully."
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

// TestReconcileBrokerErrorFetchingCatalog simulates broker reconciliation where
// OSB client responds with an error for getting the catalog which in turn causes
// reconcileBroker() to return an error.
func TestReconcileBrokerErrorFetchingCatalog(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t, fakeosb.FakeClientConfiguration{
		CatalogReaction: &fakeosb.CatalogReaction{
			Error: errors.New("ooops"),
		},
	})

	broker := getTestBroker()

	if err := testController.reconcileBroker(broker); err == nil {
		t.Fatal("Should have failed to get the catalog.")
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertGetCatalog(t, brokerActions[0])

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 1)

	updatedBroker := assertUpdateStatus(t, actions[0], broker)
	assertBrokerReadyFalse(t, updatedBroker)

	assertNumberOfActions(t, fakeKubeClient.Actions(), 0)

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	expectedEvent := api.EventTypeWarning + " " + errorFetchingCatalogReason + " " + "Error getting broker catalog for broker \"test-broker\": ooops"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

// TestReconcileBrokerZeroServices simulates broker reconciliation where
// OSB client responds with zero services which causes reconcileBroker()
// to return an error
func TestReconcileBrokerZeroServices(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t, fakeosb.FakeClientConfiguration{
		CatalogReaction: &fakeosb.CatalogReaction{
			Response: &osb.CatalogResponse{},
		},
	})

	broker := getTestBroker()

	if err := testController.reconcileBroker(broker); err == nil {
		t.Fatal("Broker should not have had any Service Classes.")
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertGetCatalog(t, brokerActions[0])

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 1)

	updatedBroker := assertUpdateStatus(t, actions[0], broker)
	assertBrokerReadyFalse(t, updatedBroker)

	assertNumberOfActions(t, fakeKubeClient.Actions(), 0)

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	expectedEvent := api.EventTypeWarning + " " + errorSyncingCatalogReason + ` Error getting catalog payload for broker "test-broker"; received zero services; at least one service is required`
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event; \nexpected: %v\ngot:     %v", e, a)
	}
}

func TestReconcileBrokerWithAuth(t *testing.T) {
	basicAuthInfo := &v1alpha1.BrokerAuthInfo{
		Basic: &v1alpha1.BasicAuthConfig{
			SecretRef: &v1.ObjectReference{
				Namespace: "test-ns",
				Name:      "auth-secret",
			},
		},
	}
	bearerAuthInfo := &v1alpha1.BrokerAuthInfo{
		Bearer: &v1alpha1.BearerTokenAuthConfig{
			SecretRef: &v1.ObjectReference{
				Namespace: "test-ns",
				Name:      "auth-secret",
			},
		},
	}
	basicAuthSecret := &v1.Secret{
		Data: map[string][]byte{
			v1alpha1.BasicAuthUsernameKey: []byte("foo"),
			v1alpha1.BasicAuthPasswordKey: []byte("bar"),
		},
	}
	bearerAuthSecret := &v1.Secret{
		Data: map[string][]byte{
			v1alpha1.BearerTokenKey: []byte("token"),
		},
	}

	// The test cases here are testing the correctness of authentication with broker
	//
	// Anonymous struct fields:
	// name: short description of the test
	// authInfo: broker auth configuration
	// secret: auth secret to be returned upon request from Service Catalog
	// shouldSucceed: whether authentication should succeed
	cases := []struct {
		name          string
		authInfo      *v1alpha1.BrokerAuthInfo
		secret        *v1.Secret
		shouldSucceed bool
	}{
		{
			name:          "basic auth - normal",
			authInfo:      basicAuthInfo,
			secret:        basicAuthSecret,
			shouldSucceed: true,
		},
		{
			name:          "basic auth - invalid secret",
			authInfo:      basicAuthInfo,
			secret:        bearerAuthSecret,
			shouldSucceed: false,
		},
		{
			name:          "basic auth - secret not found",
			authInfo:      basicAuthInfo,
			secret:        nil,
			shouldSucceed: false,
		},
		{
			name:          "bearer auth - normal",
			authInfo:      bearerAuthInfo,
			secret:        bearerAuthSecret,
			shouldSucceed: true,
		},
		{
			name:          "bearer auth - invalid secret",
			authInfo:      bearerAuthInfo,
			secret:        basicAuthSecret,
			shouldSucceed: false,
		},
		{
			name:          "bearer auth - secret not found",
			authInfo:      bearerAuthInfo,
			secret:        nil,
			shouldSucceed: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testReconcileBrokerWithAuth(t, tc.authInfo, tc.secret, tc.shouldSucceed)
		})
	}
}

func testReconcileBrokerWithAuth(t *testing.T, authInfo *v1alpha1.BrokerAuthInfo, secret *v1.Secret, shouldSucceed bool) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t, fakeosb.FakeClientConfiguration{})

	broker := getTestBrokerWithAuth(authInfo)
	if secret != nil {
		addGetSecretReaction(fakeKubeClient, secret)
	} else {
		addGetSecretNotFoundReaction(fakeKubeClient)
	}
	testServiceClass := getTestServiceClass()
	fakeBrokerClient.CatalogReaction = &fakeosb.CatalogReaction{
		Response: &osb.CatalogResponse{
			Services: []osb.Service{
				{
					ID:   testServiceClass.ExternalID,
					Name: testServiceClass.Name,
				},
			},
		},
	}

	err := testController.reconcileBroker(broker)
	if shouldSucceed && err != nil {
		t.Fatal("Should have succeeded to get the catalog for the broker. got error: ", err)
	}

	brokerActions := fakeBrokerClient.Actions()
	if shouldSucceed {
		// GetCatalog
		assertNumberOfBrokerActions(t, brokerActions, 1)
		assertGetCatalog(t, brokerActions[0])
	} else {
		assertNumberOfBrokerActions(t, brokerActions, 0)
	}

	actions := fakeCatalogClient.Actions()
	if shouldSucceed {
		assertNumberOfActions(t, actions, 2)
		assertCreate(t, actions[0], testServiceClass)
		updatedBroker := assertUpdateStatus(t, actions[1], broker)
		assertBrokerReadyTrue(t, updatedBroker)
	} else {
		assertNumberOfActions(t, actions, 1)
		updatedBroker := assertUpdateStatus(t, actions[0], broker)
		assertBrokerReadyFalse(t, updatedBroker)
	}

	// verify one kube action occurred
	kubeActions := fakeKubeClient.Actions()
	assertNumberOfActions(t, kubeActions, 1)

	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	if e, a := "secrets", getAction.GetResource().Resource; e != a {
		t.Fatalf("Unexpected resource on action; expected %v, got %v", e, a)
	}

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	var expectedEvent string
	if shouldSucceed {
		expectedEvent = api.EventTypeNormal + " " + successFetchedCatalogReason + " " + successFetchedCatalogMessage
	} else {
		expectedEvent = api.EventTypeWarning + " " + errorAuthCredentialsReason + " " + "Error getting broker auth credentials for broker \"test-broker\""
	}
	if e, a := expectedEvent, events[0]; !strings.HasPrefix(a, e) {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

// TestReconcileBrokerWithReconcileError simulates broker reconciliation where
// creation of a service class causes an error which causes ReconcileBroker to
// return an error
func TestReconcileBrokerWithReconcileError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t, getTestCatalogConfig())

	broker := getTestBroker()

	fakeCatalogClient.AddReactor("create", "serviceclasses", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error creating serviceclass")
	})

	if err := testController.reconcileBroker(broker); err == nil {
		t.Fatal("There should have been an error.")
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertGetCatalog(t, brokerActions[0])

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 2)

	createSCAction := actions[0].(clientgotesting.CreateAction)
	createdSC, ok := createSCAction.GetObject().(*v1alpha1.ServiceClass)
	if !ok {
		t.Fatalf("couldn't convert to a ServiceClass: %+v", createSCAction.GetObject())
	}
	if e, a := getTestServiceClass(), createdSC; !reflect.DeepEqual(e, a) {
		t.Fatalf("unexpected diff for created ServiceClass: %v,\n\nEXPECTED: %+v\n\nACTUAL:  %+v", diff.ObjectReflectDiff(e, a), e, a)
	}
	updatedBroker := assertUpdateStatus(t, actions[1], broker)
	assertBrokerReadyFalse(t, updatedBroker)

	kubeActions := fakeKubeClient.Actions()
	assertNumberOfActions(t, kubeActions, 0)

	events := getRecordedEvents(testController)
	assertNumEvents(t, events, 1)

	expectedEvent := api.EventTypeWarning + " " + errorSyncingCatalogReason + ` Error reconciling serviceClass "test-serviceclass" (broker "test-broker"): error creating serviceclass`
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

// TestUpdateBrokerCondition ensures that with specific conditions
// the broker correctly reflects the changes during updateBrokerCondition().
//
// The test cases are proving:
// - broker transitions from unset status to not ready results in status change and new time
// - broker transitions from not ready to not ready results in no changes
// - broker transitions from not ready to not ready and with reason & msg updates results in no time change, but reflects new reason & msg
// - broker transitions from not ready to ready results in status change & new time
// - broker transitions from ready to ready results in no status change
// - broker transitions from ready to not ready results in status change & new time
// - condition reason & message should always be updated
func TestUpdateBrokerCondition(t *testing.T) {
	// Anonymous struct fields:
	// name: short description of the test
	// input: broker object to test
	// status: new condition status
	// reason: condition reason
	// message: condition message
	// transitionTimeChanged: true if the test conditions should result in transition time change
	cases := []struct {
		name                  string
		input                 *v1alpha1.Broker
		status                v1alpha1.ConditionStatus
		reason                string
		message               string
		transitionTimeChanged bool
	}{

		{
			name:                  "initially unset",
			input:                 getTestBroker(),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
		{
			name:                  "not ready -> not ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: false,
		},
		{
			name:                  "not ready -> not ready with reason and message change",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionFalse,
			reason:                "foo",
			message:               "bar",
			transitionTimeChanged: false,
		},
		{
			name:                  "not ready -> ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: true,
		},
		{
			name:                  "ready -> ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: false,
		},
		{
			name:                  "ready -> not ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
	}

	for _, tc := range cases {
		_, fakeCatalogClient, _, testController, _ := newTestController(t, getTestCatalogConfig())

		clone, err := api.Scheme.DeepCopy(tc.input)
		if err != nil {
			t.Errorf("%v: deep copy failed", tc.name)
			continue
		}

		inputClone := clone.(*v1alpha1.Broker)

		err = testController.updateBrokerCondition(tc.input, v1alpha1.BrokerConditionReady, tc.status, tc.reason, tc.message)
		if err != nil {
			t.Errorf("%v: error updating broker condition: %v", tc.name, err)
			continue
		}

		if !reflect.DeepEqual(tc.input, inputClone) {
			t.Errorf("%v: updating broker condition mutated input: expected %v, got %v", tc.name, inputClone, tc.input)
			continue
		}

		actions := fakeCatalogClient.Actions()
		if ok := expectNumberOfActions(t, tc.name, actions, 1); !ok {
			continue
		}

		updatedBroker, ok := expectUpdateStatus(t, tc.name, actions[0], tc.input)
		if !ok {
			continue
		}

		updateActionObject, ok := updatedBroker.(*v1alpha1.Broker)
		if !ok {
			t.Errorf("%v: couldn't convert to broker", tc.name)
			continue
		}

		var initialTs metav1.Time
		if len(inputClone.Status.Conditions) != 0 {
			initialTs = inputClone.Status.Conditions[0].LastTransitionTime
		}

		if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
			t.Errorf("%v: expected %v condition(s), got %v", tc.name, e, a)
		}

		outputCondition := updateActionObject.Status.Conditions[0]
		newTs := outputCondition.LastTransitionTime

		if tc.transitionTimeChanged && initialTs == newTs {
			t.Errorf("%v: transition time didn't change when it should have", tc.name)
			continue
		} else if !tc.transitionTimeChanged && initialTs != newTs {
			t.Errorf("%v: transition time changed when it shouldn't have", tc.name)
			continue
		}
		if e, a := tc.reason, outputCondition.Reason; e != "" && e != a {
			t.Errorf("%v: condition reasons didn't match; expected %v, got %v", tc.name, e, a)
			continue
		}
		if e, a := tc.message, outputCondition.Message; e != "" && e != a {
			t.Errorf("%v: condition message didn't match; expected %v, got %v", tc.name, e, a)
		}
	}
}
