package reaper

import (
	"github.com/golang/glog"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/openshift/origin/pkg/client"
)

// reapClusterBindings removes the subject from cluster-level role bindings
func reapClusterBindings(removedSubject kapi.ObjectReference, c client.ClusterRoleBindingsInterface) error {
	clusterBindings, err := c.ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, binding := range clusterBindings.Items {
		retainedSubjects := []kapi.ObjectReference{}
		for _, subject := range binding.Subjects {
			if subject != removedSubject {
				retainedSubjects = append(retainedSubjects, subject)
			}
		}
		if len(retainedSubjects) != len(binding.Subjects) {
			updatedBinding := binding
			updatedBinding.Subjects = retainedSubjects
			if _, err := c.ClusterRoleBindings().Update(&updatedBinding); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update clusterrolebinding/%s: %v", binding.Name, err)
			}
		}
	}
	return nil
}

// reapNamespacedBindings removes the subject from namespaced role bindings
func reapNamespacedBindings(removedSubject kapi.ObjectReference, c client.RoleBindingsNamespacer) error {
	namespacedBindings, err := c.RoleBindings(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, binding := range namespacedBindings.Items {
		retainedSubjects := []kapi.ObjectReference{}
		for _, subject := range binding.Subjects {
			if subject != removedSubject {
				retainedSubjects = append(retainedSubjects, subject)
			}
		}
		if len(retainedSubjects) != len(binding.Subjects) {
			updatedBinding := binding
			updatedBinding.Subjects = retainedSubjects
			if _, err := c.RoleBindings(binding.Namespace).Update(&updatedBinding); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update rolebinding/%s in %s: %v", binding.Name, binding.Namespace, err)
			}
		}
	}
	return nil
}
