package util

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kapi "k8s.io/kubernetes/pkg/api"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"

	"github.com/openshift/origin/pkg/cmd/util"
)

// Namespace returns the test namespace. The default namespace is set to
// 'integration-test'. You can override it by setting the 'OS_TEST_NAMESPACE'
// environment variable
func Namespace() string {
	return util.Env("OS_TEST_NAMESPACE", "integration")
}

// RandomNamespace provides random Kubernetes namespace name based on the UNIX
// timestamp. Optionally you can set the prefix.
func RandomNamespace(prefix string) string {
	return prefix + string([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))[3:12])
}

// CreateNamespace creates a namespace with the specified name using the provided kubeconfig
// DO NOT USE, use create project instead
func CreateNamespace(clusterAdminKubeConfig, name string) (err error) {
	clusterAdminKubeClient, err := GetClusterAdminKubeClient(clusterAdminKubeConfig)
	if err != nil {
		return err
	}
	_, err = clusterAdminKubeClient.Core().Namespaces().Create(&kapi.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	})
	return err
}

func DeleteAndWaitForNamespaceTermination(c kclientset.Interface, name string) error {
	w, err := c.Core().Namespaces().Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer w.Stop()

	if err := c.Core().Namespaces().Delete(name, nil); err != nil {
		return err
	}
	_, err = watch.Until(30*time.Second, w, func(event watch.Event) (bool, error) {
		if event.Type != watch.Deleted {
			return false, nil
		}
		namespace, ok := event.Object.(*kapi.Namespace)
		if !ok {
			return false, nil
		}
		return namespace.Name == name, nil
	})
	return err
}
