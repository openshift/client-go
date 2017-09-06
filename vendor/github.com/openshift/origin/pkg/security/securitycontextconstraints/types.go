package securitycontextconstraints

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/api"
)

// SecurityContextConstraintsProvider provides the implementation to generate a new security
// context based on constraints or validate an existing security context against constraints.
type SecurityContextConstraintsProvider interface {
	// Create a PodSecurityContext based on the given constraints.
	CreatePodSecurityContext(pod *api.Pod) (*api.PodSecurityContext, map[string]string, error)
	// Create a container SecurityContext based on the given constraints
	CreateContainerSecurityContext(pod *api.Pod, container *api.Container) (*api.SecurityContext, error)
	// Ensure a pod's SecurityContext is in compliance with the given constraints.
	ValidatePodSecurityContext(pod *api.Pod, fldPath *field.Path) field.ErrorList
	// Ensure a container's SecurityContext is in compliance with the given constraints
	ValidateContainerSecurityContext(pod *api.Pod, container *api.Container, fldPath *field.Path) field.ErrorList
	// Get the name of the SCC that this provider was initialized with.
	GetSCCName() string
}
