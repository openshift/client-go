// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// ConsoleSampleGitImportSourceApplyConfiguration represents a declarative configuration of the ConsoleSampleGitImportSource type for use
// with apply.
type ConsoleSampleGitImportSourceApplyConfiguration struct {
	Repository *ConsoleSampleGitImportSourceRepositoryApplyConfiguration `json:"repository,omitempty"`
	Service    *ConsoleSampleGitImportSourceServiceApplyConfiguration    `json:"service,omitempty"`
}

// ConsoleSampleGitImportSourceApplyConfiguration constructs a declarative configuration of the ConsoleSampleGitImportSource type for use with
// apply.
func ConsoleSampleGitImportSource() *ConsoleSampleGitImportSourceApplyConfiguration {
	return &ConsoleSampleGitImportSourceApplyConfiguration{}
}

// WithRepository sets the Repository field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Repository field is set to the value of the last call.
func (b *ConsoleSampleGitImportSourceApplyConfiguration) WithRepository(value *ConsoleSampleGitImportSourceRepositoryApplyConfiguration) *ConsoleSampleGitImportSourceApplyConfiguration {
	b.Repository = value
	return b
}

// WithService sets the Service field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Service field is set to the value of the last call.
func (b *ConsoleSampleGitImportSourceApplyConfiguration) WithService(value *ConsoleSampleGitImportSourceServiceApplyConfiguration) *ConsoleSampleGitImportSourceApplyConfiguration {
	b.Service = value
	return b
}
