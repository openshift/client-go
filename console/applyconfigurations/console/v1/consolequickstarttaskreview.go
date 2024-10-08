// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// ConsoleQuickStartTaskReviewApplyConfiguration represents a declarative configuration of the ConsoleQuickStartTaskReview type for use
// with apply.
type ConsoleQuickStartTaskReviewApplyConfiguration struct {
	Instructions   *string `json:"instructions,omitempty"`
	FailedTaskHelp *string `json:"failedTaskHelp,omitempty"`
}

// ConsoleQuickStartTaskReviewApplyConfiguration constructs a declarative configuration of the ConsoleQuickStartTaskReview type for use with
// apply.
func ConsoleQuickStartTaskReview() *ConsoleQuickStartTaskReviewApplyConfiguration {
	return &ConsoleQuickStartTaskReviewApplyConfiguration{}
}

// WithInstructions sets the Instructions field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Instructions field is set to the value of the last call.
func (b *ConsoleQuickStartTaskReviewApplyConfiguration) WithInstructions(value string) *ConsoleQuickStartTaskReviewApplyConfiguration {
	b.Instructions = &value
	return b
}

// WithFailedTaskHelp sets the FailedTaskHelp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FailedTaskHelp field is set to the value of the last call.
func (b *ConsoleQuickStartTaskReviewApplyConfiguration) WithFailedTaskHelp(value string) *ConsoleQuickStartTaskReviewApplyConfiguration {
	b.FailedTaskHelp = &value
	return b
}
