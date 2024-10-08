// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfigurations

import (
	v1 "github.com/openshift/api/apiserver/v1"
	apiserverv1 "github.com/openshift/client-go/apiserver/applyconfigurations/apiserver/v1"
	internal "github.com/openshift/client-go/apiserver/applyconfigurations/internal"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=apiserver.openshift.io, Version=v1
	case v1.SchemeGroupVersion.WithKind("APIRequestCount"):
		return &apiserverv1.APIRequestCountApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("APIRequestCountSpec"):
		return &apiserverv1.APIRequestCountSpecApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("APIRequestCountStatus"):
		return &apiserverv1.APIRequestCountStatusApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("PerNodeAPIRequestLog"):
		return &apiserverv1.PerNodeAPIRequestLogApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("PerResourceAPIRequestLog"):
		return &apiserverv1.PerResourceAPIRequestLogApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("PerUserAPIRequestCount"):
		return &apiserverv1.PerUserAPIRequestCountApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("PerVerbAPIRequestCount"):
		return &apiserverv1.PerVerbAPIRequestCountApplyConfiguration{}

	}
	return nil
}

func NewTypeConverter(scheme *runtime.Scheme) *testing.TypeConverter {
	return &testing.TypeConverter{Scheme: scheme, TypeResolver: internal.Parser()}
}
