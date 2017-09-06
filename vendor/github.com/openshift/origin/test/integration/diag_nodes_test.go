package integration

import (
	"testing"
	"time"

	kapierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kapi "k8s.io/kubernetes/pkg/api"

	clusterdiags "github.com/openshift/origin/pkg/diagnostics/cluster"
	diagtype "github.com/openshift/origin/pkg/diagnostics/types"
	testutil "github.com/openshift/origin/test/util"
	testserver "github.com/openshift/origin/test/util/server"
)

func TestDiagNodeConditions(t *testing.T) {
	masterConfig, nodeConfig, clientFile, err := testserver.StartTestAllInOne()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer testserver.CleanupMasterEtcd(t, masterConfig)
	client, err := testutil.GetClusterAdminKubeClient(clientFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodeDiag := clusterdiags.NodeDefinitions{KubeClient: client}
	err = wait.Poll(200*time.Millisecond, 5*time.Second, func() (bool, error) {
		if _, err := client.Core().Nodes().Get(nodeConfig.NodeName, metav1.GetOptions{}); kapierror.IsNotFound(err) {
			return false, nil
		}
		return true, err
	})
	if err != nil {
		t.Errorf("unexpected error waiting for all-in-one node: %v", err)
	}

	// start by testing that the diagnostic passes with all-in-one up.
	result := nodeDiag.Check()
	if warnings := result.Warnings(); len(warnings) > 0 {
		t.Fatalf("expected no warnings with one node ready, but: %#v", warnings)
	} else if errors := result.Errors(); len(errors) > 0 {
		t.Fatalf("expected no errors with one node ready, but: %#v", errors)
	}

	// Make the node unschedulable and verify diagnostics notices
	err = wait.Poll(200*time.Millisecond, time.Second, func() (bool, error) {
		node, err := client.Core().Nodes().Get(nodeConfig.NodeName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		node.Spec.Unschedulable = true
		if _, err := client.Core().Nodes().Update(node); kapierror.IsConflict(err) {
			return false, nil
		}
		return true, err
	})
	if err != nil {
		t.Errorf("unexpected error making node unschedulable: %v", err)
	}
	result = nodeDiag.Check()
	if errors := result.Errors(); len(errors) != 1 ||
		!diagtype.MatchesDiagError(errors[0], "DClu0004") {
		t.Fatalf("expected 1 error about not having nodes, but: %#v", errors)
	} else if warnings := result.Warnings(); len(warnings) < 1 || !diagtype.MatchesDiagError(warnings[0], "DClu0003") {
		t.Fatalf("expected a warning about test-node not being schedulable, but: %#v", warnings)
	}

	// delete it and check with no nodes defined; should get an error about that.
	if err := client.Core().Nodes().Delete(nodeConfig.NodeName, nil); err != nil {
		t.Errorf("unexpected error deleting node: %v", err)
	}
	if errors := nodeDiag.Check().Errors(); len(errors) != 1 ||
		!diagtype.MatchesDiagError(errors[0], "DClu0004") {
		t.Fatalf("expected 1 error about not having nodes, not: %#v", errors)
	}

	// Next create a node and leave it in NotReady state. Should get a warning
	// about that, plus the previous error as there are still no nodes available.
	_, err = client.Core().Nodes().Create(&kapi.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}})
	if err != nil {
		t.Fatalf("expected no errors creating a node: %#v", err)
	}
	result = nodeDiag.Check()
	if errors := result.Errors(); len(errors) != 1 ||
		!diagtype.MatchesDiagError(errors[0], "DClu0004") {
		t.Fatalf("expected 1 error about not having nodes, not: %#v", errors)
	} else if warnings := result.Warnings(); len(warnings) < 1 || !diagtype.MatchesDiagError(warnings[0], "DClu0002") {
		t.Fatalf("expected a warning about test-node not being ready, not: %#v", warnings)
	}
}
