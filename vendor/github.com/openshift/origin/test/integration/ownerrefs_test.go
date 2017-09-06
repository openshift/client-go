package integration

import (
	"strings"
	"testing"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	testutil "github.com/openshift/origin/test/util"
	testserver "github.com/openshift/origin/test/util/server"
)

func TestOwnerRefRestriction(t *testing.T) {
	// functionality of the plugin has a unit test, we just need to make sure its called.
	masterConfig, clusterAdminKubeConfig, err := testserver.StartTestMasterAPI()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer testserver.CleanupMasterEtcd(t, masterConfig)

	clientConfig, err := testutil.GetClusterAdminClientConfig(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	originClient, err := testutil.GetClusterAdminClient(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = originClient.ClusterRoles().Create(&authorizationapi.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "create-svc",
		},
		Rules: []authorizationapi.PolicyRule{
			authorizationapi.NewRule("create").Groups(kapi.GroupName).Resources("services").RuleOrDie(),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := testserver.CreateNewProject(originClient, *clientConfig, "foo", "admin-user"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, creatorClient, _, err := testutil.GetClientForUser(*clientConfig, "creator")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = originClient.RoleBindings("foo").Create(&authorizationapi.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "create-svc",
		},
		RoleRef:  kapi.ObjectReference{Name: "create-svc"},
		Subjects: []kapi.ObjectReference{{Kind: authorizationapi.UserKind, Name: "creator"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := testutil.WaitForPolicyUpdate(originClient, "foo", "create", kapi.Resource("services"), true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = creatorClient.Core().Services("foo").Create(&kapi.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-service",
			OwnerReferences: []metav1.OwnerReference{{}},
		},
	})
	if err == nil {
		t.Fatalf("missing err")
	}
	if !kapierrors.IsForbidden(err) || !strings.Contains(err.Error(), "cannot set an ownerRef on a resource you can't delete") {
		t.Fatalf("expecting cannot set an ownerRef on a resource you can't delete, got %v", err)
	}
}
