package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectRequestLimitConfig is the configuration for the project request limit plug-in
// It contains an ordered list of limits based on user label selectors. Selectors will
// be checked in order and the first one that applies will be used as the limit.
type ProjectRequestLimitConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Limits are the project request limits
	Limits []ProjectLimitBySelector `json:"limits" description:"project request limits"`

	// MaxProjectsForSystemUsers controls how many projects a certificate user may have.  Certificate
	// users do not have any labels associated with them for more fine grained control
	MaxProjectsForSystemUsers *int `json:"maxProjectsForSystemUsers"`

	// MaxProjectsForServiceAccounts controls how many projects a service account may have.  Service
	// accounts can't create projects by default, but if they are allowed to create projects, you cannot
	// trust any labels placed on them since project editors can manipulate those labels
	MaxProjectsForServiceAccounts *int `json:"maxProjectsForServiceAccounts"`
}

// ProjectLimitBySelector specifies the maximum number of projects allowed for a given user label selector
type ProjectLimitBySelector struct {
	// Selector is a user label selector. An empty selector selects everything.
	Selector map[string]string `json:"selector" description:"user label selector"`
	// MaxProjects is the number of projects allowed for this class of users. If MaxProjects is nil,
	// there is no limit to the number of projects users can request. An unlimited number of projects
	// is useful in the case a limit is specified as the default for all users and only users with a
	// specific set of labels should be allowed unlimited project creation.
	MaxProjects *int `json:"maxProjects,omitempty" description:"maximum number of projects, unlimited if nil"`
}
