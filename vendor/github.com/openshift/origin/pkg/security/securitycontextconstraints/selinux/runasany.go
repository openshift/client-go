package selinux

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/api"

	securityapi "github.com/openshift/origin/pkg/security/apis/security"
)

// runAsAny implements the SELinuxSecurityContextConstraintsStrategy interface.
type runAsAny struct{}

var _ SELinuxSecurityContextConstraintsStrategy = &runAsAny{}

// NewRunAsAny provides a strategy that will return the configured se linux context or nil.
func NewRunAsAny(options *securityapi.SELinuxContextStrategyOptions) (SELinuxSecurityContextConstraintsStrategy, error) {
	return &runAsAny{}, nil
}

// Generate creates the SELinuxOptions based on constraint rules.
func (s *runAsAny) Generate(pod *api.Pod, container *api.Container) (*api.SELinuxOptions, error) {
	return nil, nil
}

// Validate ensures that the specified values fall within the range of the strategy.
func (s *runAsAny) Validate(pod *api.Pod, container *api.Container) field.ErrorList {
	return field.ErrorList{}
}
