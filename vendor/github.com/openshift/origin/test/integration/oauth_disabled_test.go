package integration

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
	testutil "github.com/openshift/origin/test/util"
	testserver "github.com/openshift/origin/test/util/server"
)

func TestOAuthDisabled(t *testing.T) {
	// Build master config
	masterOptions, err := testserver.DefaultMasterOptions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer testserver.CleanupMasterEtcd(t, masterOptions)

	// Disable OAuth
	masterOptions.OAuthConfig = nil

	// Start server
	clusterAdminKubeConfig, err := testserver.StartConfiguredMaster(masterOptions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client, err := testutil.GetClusterAdminKubeClient(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clientConfig, err := testutil.GetClusterAdminClientConfig(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Make sure cert auth still works
	namespaces, err := client.Core().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(namespaces.Items) == 0 {
		t.Errorf("Expected namespaces, got none")
	}

	// Use the server and CA info
	anonConfig := restclient.Config{}
	anonConfig.Host = clientConfig.Host
	anonConfig.CAFile = clientConfig.CAFile
	anonConfig.CAData = clientConfig.CAData

	// Make sure we can't authenticate using OAuth
	if _, err := tokencmd.RequestToken(&anonConfig, nil, "username", "password"); err == nil {
		t.Error("Expected error, got none")
	}

}
