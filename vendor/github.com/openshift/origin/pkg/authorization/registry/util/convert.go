package util

import (
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kubernetes/pkg/apis/rbac"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
)

// ClusterRoleToRBAC turns an OpenShift ClusterRole into a Kubernetes RBAC
// ClusterRole, the returned object is safe to mutate
func ClusterRoleToRBAC(obj *authorizationapi.ClusterRole) (*rbac.ClusterRole, error) {
	convertedObj := &rbac.ClusterRole{}
	if err := authorizationapi.Convert_authorization_ClusterRole_To_rbac_ClusterRole(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &rbac.ClusterRole{}
	if err := rbac.DeepCopy_rbac_ClusterRole(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// ClusterRoleBindingToRBAC turns an OpenShift ClusterRoleBinding into a Kubernetes
// RBAC ClusterRoleBinding, the returned object is safe to mutate
func ClusterRoleBindingToRBAC(obj *authorizationapi.ClusterRoleBinding) (*rbac.ClusterRoleBinding, error) {
	convertedObj := &rbac.ClusterRoleBinding{}
	if err := authorizationapi.Convert_authorization_ClusterRoleBinding_To_rbac_ClusterRoleBinding(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &rbac.ClusterRoleBinding{}
	if err := rbac.DeepCopy_rbac_ClusterRoleBinding(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// RoleToRBAC turns an OpenShift Role into a Kubernetes RBAC Role,
// the returned object is safe to mutate
func RoleToRBAC(obj *authorizationapi.Role) (*rbac.Role, error) {
	convertedObj := &rbac.Role{}
	if err := authorizationapi.Convert_authorization_Role_To_rbac_Role(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &rbac.Role{}
	if err := rbac.DeepCopy_rbac_Role(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// RoleBindingToRBAC turns an OpenShift RoleBinding into a Kubernetes RBAC
// Rolebinding, the returned object is safe to mutate
func RoleBindingToRBAC(obj *authorizationapi.RoleBinding) (*rbac.RoleBinding, error) {
	convertedObj := &rbac.RoleBinding{}
	if err := authorizationapi.Convert_authorization_RoleBinding_To_rbac_RoleBinding(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &rbac.RoleBinding{}
	if err := rbac.DeepCopy_rbac_RoleBinding(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// ClusterRoleFromRBAC turns a Kubernetes RBAC ClusterRole into an Openshift
// ClusterRole, the returned object is safe to mutate
func ClusterRoleFromRBAC(obj *rbac.ClusterRole) (*authorizationapi.ClusterRole, error) {
	convertedObj := &authorizationapi.ClusterRole{}
	if err := authorizationapi.Convert_rbac_ClusterRole_To_authorization_ClusterRole(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &authorizationapi.ClusterRole{}
	if err := authorizationapi.DeepCopy_authorization_ClusterRole(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// ClusterRoleBindingFromRBAC turns a Kuberenets RBAC ClusterRoleBinding into
// an Openshift ClusterRoleBinding, the returned object is safe to mutate
func ClusterRoleBindingFromRBAC(obj *rbac.ClusterRoleBinding) (*authorizationapi.ClusterRoleBinding, error) {
	convertedObj := &authorizationapi.ClusterRoleBinding{}
	if err := authorizationapi.Convert_rbac_ClusterRoleBinding_To_authorization_ClusterRoleBinding(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &authorizationapi.ClusterRoleBinding{}
	if err := authorizationapi.DeepCopy_authorization_ClusterRoleBinding(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// RoleFromRBAC turns a Kubernetes RBAC Role into an OpenShift Role,
// the returned object is safe to mutate
func RoleFromRBAC(obj *rbac.Role) (*authorizationapi.Role, error) {
	convertedObj := &authorizationapi.Role{}
	if err := authorizationapi.Convert_rbac_Role_To_authorization_Role(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &authorizationapi.Role{}
	if err := authorizationapi.DeepCopy_authorization_Role(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

// RoleBindingFromRBAC turns a Kubernetes RBAC RoleBinding into an OpenShift
// Rolebinding, the returned object is safe to mutate
func RoleBindingFromRBAC(obj *rbac.RoleBinding) (*authorizationapi.RoleBinding, error) {
	convertedObj := &authorizationapi.RoleBinding{}
	if err := authorizationapi.Convert_rbac_RoleBinding_To_authorization_RoleBinding(obj, convertedObj, nil); err != nil {
		return nil, err
	}
	// do a deep copy here since conversion does not guarantee a new object.
	deepcopiedObj := &authorizationapi.RoleBinding{}
	if err := authorizationapi.DeepCopy_authorization_RoleBinding(convertedObj, deepcopiedObj, cloner); err != nil {
		return nil, err
	}

	return deepcopiedObj, nil
}

var cloner = conversion.NewCloner()
