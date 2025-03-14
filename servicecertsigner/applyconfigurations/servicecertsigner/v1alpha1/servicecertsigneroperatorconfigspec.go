// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	operatorv1 "github.com/openshift/api/operator/v1"
	v1 "github.com/openshift/client-go/operator/applyconfigurations/operator/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ServiceCertSignerOperatorConfigSpecApplyConfiguration represents a declarative configuration of the ServiceCertSignerOperatorConfigSpec type for use
// with apply.
type ServiceCertSignerOperatorConfigSpecApplyConfiguration struct {
	v1.OperatorSpecApplyConfiguration `json:",inline"`
}

// ServiceCertSignerOperatorConfigSpecApplyConfiguration constructs a declarative configuration of the ServiceCertSignerOperatorConfigSpec type for use with
// apply.
func ServiceCertSignerOperatorConfigSpec() *ServiceCertSignerOperatorConfigSpecApplyConfiguration {
	return &ServiceCertSignerOperatorConfigSpecApplyConfiguration{}
}

// WithManagementState sets the ManagementState field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ManagementState field is set to the value of the last call.
func (b *ServiceCertSignerOperatorConfigSpecApplyConfiguration) WithManagementState(value operatorv1.ManagementState) *ServiceCertSignerOperatorConfigSpecApplyConfiguration {
	b.OperatorSpecApplyConfiguration.ManagementState = &value
	return b
}

// WithLogLevel sets the LogLevel field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LogLevel field is set to the value of the last call.
func (b *ServiceCertSignerOperatorConfigSpecApplyConfiguration) WithLogLevel(value operatorv1.LogLevel) *ServiceCertSignerOperatorConfigSpecApplyConfiguration {
	b.OperatorSpecApplyConfiguration.LogLevel = &value
	return b
}

// WithOperatorLogLevel sets the OperatorLogLevel field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OperatorLogLevel field is set to the value of the last call.
func (b *ServiceCertSignerOperatorConfigSpecApplyConfiguration) WithOperatorLogLevel(value operatorv1.LogLevel) *ServiceCertSignerOperatorConfigSpecApplyConfiguration {
	b.OperatorSpecApplyConfiguration.OperatorLogLevel = &value
	return b
}

// WithUnsupportedConfigOverrides sets the UnsupportedConfigOverrides field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UnsupportedConfigOverrides field is set to the value of the last call.
func (b *ServiceCertSignerOperatorConfigSpecApplyConfiguration) WithUnsupportedConfigOverrides(value runtime.RawExtension) *ServiceCertSignerOperatorConfigSpecApplyConfiguration {
	b.OperatorSpecApplyConfiguration.UnsupportedConfigOverrides = &value
	return b
}

// WithObservedConfig sets the ObservedConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ObservedConfig field is set to the value of the last call.
func (b *ServiceCertSignerOperatorConfigSpecApplyConfiguration) WithObservedConfig(value runtime.RawExtension) *ServiceCertSignerOperatorConfigSpecApplyConfiguration {
	b.OperatorSpecApplyConfiguration.ObservedConfig = &value
	return b
}
