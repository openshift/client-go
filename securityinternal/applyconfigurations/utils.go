// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfigurations

import (
	v1 "github.com/openshift/api/securityinternal/v1"
	securityinternalv1 "github.com/openshift/client-go/securityinternal/applyconfigurations/securityinternal/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=security.internal.openshift.io, Version=v1
	case v1.SchemeGroupVersion.WithKind("RangeAllocation"):
		return &securityinternalv1.RangeAllocationApplyConfiguration{}

	}
	return nil
}
