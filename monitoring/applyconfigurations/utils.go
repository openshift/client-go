// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfigurations

import (
	v1 "github.com/openshift/api/monitoring/v1"
	internal "github.com/openshift/client-go/monitoring/applyconfigurations/internal"
	monitoringv1 "github.com/openshift/client-go/monitoring/applyconfigurations/monitoring/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=monitoring.openshift.io, Version=v1
	case v1.SchemeGroupVersion.WithKind("AlertingRule"):
		return &monitoringv1.AlertingRuleApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("AlertingRuleSpec"):
		return &monitoringv1.AlertingRuleSpecApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("AlertingRuleStatus"):
		return &monitoringv1.AlertingRuleStatusApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("AlertRelabelConfig"):
		return &monitoringv1.AlertRelabelConfigApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("AlertRelabelConfigSpec"):
		return &monitoringv1.AlertRelabelConfigSpecApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("AlertRelabelConfigStatus"):
		return &monitoringv1.AlertRelabelConfigStatusApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("PrometheusRuleRef"):
		return &monitoringv1.PrometheusRuleRefApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("RelabelConfig"):
		return &monitoringv1.RelabelConfigApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("Rule"):
		return &monitoringv1.RuleApplyConfiguration{}
	case v1.SchemeGroupVersion.WithKind("RuleGroup"):
		return &monitoringv1.RuleGroupApplyConfiguration{}

	}
	return nil
}

func NewTypeConverter(scheme *runtime.Scheme) *testing.TypeConverter {
	return &testing.TypeConverter{Scheme: scheme, TypeResolver: internal.Parser()}
}
