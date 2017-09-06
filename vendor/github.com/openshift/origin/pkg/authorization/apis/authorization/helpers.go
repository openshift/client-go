package authorization

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/validation"
	// uservalidation "github.com/openshift/origin/pkg/user/apis/user/validation"
)

func (r PolicyRule) String() string {
	return "PolicyRule" + r.CompactString()
}

// CompactString exposes a compact string representation for use in escalation error messages
func (r PolicyRule) CompactString() string {
	formatStringParts := []string{}
	formatArgs := []interface{}{}
	if len(r.Verbs) > 0 {
		formatStringParts = append(formatStringParts, "Verbs:%q")
		formatArgs = append(formatArgs, r.Verbs.List())
	}
	if len(r.APIGroups) > 0 {
		formatStringParts = append(formatStringParts, "APIGroups:%q")
		formatArgs = append(formatArgs, r.APIGroups)
	}
	if len(r.Resources) > 0 {
		formatStringParts = append(formatStringParts, "Resources:%q")
		formatArgs = append(formatArgs, r.Resources.List())
	}
	if len(r.ResourceNames) > 0 {
		formatStringParts = append(formatStringParts, "ResourceNames:%q")
		formatArgs = append(formatArgs, r.ResourceNames.List())
	}
	if r.AttributeRestrictions != nil {
		formatStringParts = append(formatStringParts, "Restrictions:%q")
		formatArgs = append(formatArgs, r.AttributeRestrictions)
	}
	if len(r.NonResourceURLs) > 0 {
		formatStringParts = append(formatStringParts, "NonResourceURLs:%q")
		formatArgs = append(formatArgs, r.NonResourceURLs.List())
	}
	formatString := "{" + strings.Join(formatStringParts, ", ") + "}"
	return fmt.Sprintf(formatString, formatArgs...)
}

type PolicyBindingSorter []PolicyBinding

func (s PolicyBindingSorter) Len() int {
	return len(s)
}
func (s PolicyBindingSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s PolicyBindingSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type RoleBindingSorter []RoleBinding

func (s RoleBindingSorter) Len() int {
	return len(s)
}
func (s RoleBindingSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s RoleBindingSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func GetPolicyBindingName(policyRefNamespace string) string {
	return fmt.Sprintf("%s:%s", policyRefNamespace, PolicyName)
}

var ClusterPolicyBindingName = GetPolicyBindingName("")

func BuildSubjects(users, groups []string, userNameValidator, groupNameValidator validation.ValidateNameFunc) []kapi.ObjectReference {
	subjects := []kapi.ObjectReference{}

	for _, user := range users {
		saNamespace, saName, err := serviceaccount.SplitUsername(user)
		if err == nil {
			subjects = append(subjects, kapi.ObjectReference{Kind: ServiceAccountKind, Namespace: saNamespace, Name: saName})
			continue
		}

		kind := determineUserKind(user, userNameValidator)
		subjects = append(subjects, kapi.ObjectReference{Kind: kind, Name: user})
	}

	for _, group := range groups {
		kind := determineGroupKind(group, groupNameValidator)
		subjects = append(subjects, kapi.ObjectReference{Kind: kind, Name: group})
	}

	return subjects
}

func determineUserKind(user string, userNameValidator validation.ValidateNameFunc) string {
	kind := UserKind
	if len(userNameValidator(user, false)) != 0 {
		kind = SystemUserKind
	}
	return kind
}

func determineGroupKind(group string, groupNameValidator validation.ValidateNameFunc) string {
	kind := GroupKind
	if len(groupNameValidator(group, false)) != 0 {
		kind = SystemGroupKind
	}
	return kind
}

// StringSubjectsFor returns users and groups for comparison against user.Info.  currentNamespace is used to
// to create usernames for service accounts where namespace=="".
func StringSubjectsFor(currentNamespace string, subjects []kapi.ObjectReference) ([]string, []string) {
	// these MUST be nil to indicate empty
	var users, groups []string

	for _, subject := range subjects {
		switch subject.Kind {
		case ServiceAccountKind:
			namespace := currentNamespace
			if len(subject.Namespace) > 0 {
				namespace = subject.Namespace
			}
			if len(namespace) > 0 {
				users = append(users, serviceaccount.MakeUsername(namespace, subject.Name))
			}

		case UserKind, SystemUserKind:
			users = append(users, subject.Name)

		case GroupKind, SystemGroupKind:
			groups = append(groups, subject.Name)
		}
	}

	return users, groups
}

// SubjectsStrings returns users, groups, serviceaccounts, unknown for display purposes.  currentNamespace is used to
// hide the subject.Namespace for ServiceAccounts in the currentNamespace
func SubjectsStrings(currentNamespace string, subjects []kapi.ObjectReference) ([]string, []string, []string, []string) {
	users := []string{}
	groups := []string{}
	sas := []string{}
	others := []string{}

	for _, subject := range subjects {
		switch subject.Kind {
		case ServiceAccountKind:
			if len(subject.Namespace) > 0 && currentNamespace != subject.Namespace {
				sas = append(sas, subject.Namespace+"/"+subject.Name)
			} else {
				sas = append(sas, subject.Name)
			}

		case UserKind, SystemUserKind:
			users = append(users, subject.Name)

		case GroupKind, SystemGroupKind:
			groups = append(groups, subject.Name)

		default:
			others = append(others, fmt.Sprintf("%s/%s/%s", subject.Kind, subject.Namespace, subject.Name))

		}
	}

	return users, groups, sas, others
}

// SubjectsContainUser returns true if the provided subjects contain the named user. currentNamespace
// is used to identify service accounts that are defined in a relative fashion.
func SubjectsContainUser(subjects []kapi.ObjectReference, currentNamespace string, user string) bool {
	if !strings.HasPrefix(user, serviceaccount.ServiceAccountUsernamePrefix) {
		for _, subject := range subjects {
			switch subject.Kind {
			case UserKind, SystemUserKind:
				if user == subject.Name {
					return true
				}
			}
		}
		return false
	}

	for _, subject := range subjects {
		switch subject.Kind {
		case ServiceAccountKind:
			namespace := currentNamespace
			if len(subject.Namespace) > 0 {
				namespace = subject.Namespace
			}
			if len(namespace) == 0 {
				continue
			}
			if user == serviceaccount.MakeUsername(namespace, subject.Name) {
				return true
			}

		case UserKind, SystemUserKind:
			if user == subject.Name {
				return true
			}
		}
	}
	return false
}

// SubjectsContainAnyGroup returns true if the provided subjects any of the named groups.
func SubjectsContainAnyGroup(subjects []kapi.ObjectReference, groups []string) bool {
	for _, subject := range subjects {
		switch subject.Kind {
		case GroupKind, SystemGroupKind:
			for _, group := range groups {
				if group == subject.Name {
					return true
				}
			}
		}
	}
	return false
}

func AddUserToSAR(user user.Info, sar *SubjectAccessReview) *SubjectAccessReview {
	origScopes := user.GetExtra()[ScopesKey]
	scopes := make([]string, len(origScopes), len(origScopes))
	copy(scopes, origScopes)

	sar.User = user.GetName()
	sar.Groups = sets.NewString(user.GetGroups()...)
	sar.Scopes = scopes
	return sar
}
func AddUserToLSAR(user user.Info, lsar *LocalSubjectAccessReview) *LocalSubjectAccessReview {
	origScopes := user.GetExtra()[ScopesKey]
	scopes := make([]string, len(origScopes), len(origScopes))
	copy(scopes, origScopes)

	lsar.User = user.GetName()
	lsar.Groups = sets.NewString(user.GetGroups()...)
	lsar.Scopes = scopes
	return lsar
}

// +gencopy=false
// PolicyRuleBuilder let's us attach methods.  A no-no for API types
type PolicyRuleBuilder struct {
	PolicyRule PolicyRule
}

func NewRule(verbs ...string) *PolicyRuleBuilder {
	return &PolicyRuleBuilder{
		PolicyRule: PolicyRule{
			Verbs:         sets.NewString(verbs...),
			Resources:     sets.String{},
			ResourceNames: sets.String{},
		},
	}
}

func (r *PolicyRuleBuilder) Groups(groups ...string) *PolicyRuleBuilder {
	r.PolicyRule.APIGroups = append(r.PolicyRule.APIGroups, groups...)
	return r
}

func (r *PolicyRuleBuilder) Resources(resources ...string) *PolicyRuleBuilder {
	r.PolicyRule.Resources.Insert(resources...)
	return r
}

func (r *PolicyRuleBuilder) Names(names ...string) *PolicyRuleBuilder {
	r.PolicyRule.ResourceNames.Insert(names...)
	return r
}

func (r *PolicyRuleBuilder) RuleOrDie() PolicyRule {
	ret, err := r.Rule()
	if err != nil {
		panic(err)
	}
	return ret
}

func (r *PolicyRuleBuilder) Rule() (PolicyRule, error) {
	if r.PolicyRule.AttributeRestrictions != nil {
		return PolicyRule{}, fmt.Errorf("rule may not have attributeRestrictions because they are deprecated and ignored: %#v", r.PolicyRule)
	}

	if len(r.PolicyRule.Verbs) == 0 {
		return PolicyRule{}, fmt.Errorf("verbs are required: %#v", r.PolicyRule)
	}

	switch {
	case len(r.PolicyRule.NonResourceURLs) > 0:
		if len(r.PolicyRule.APIGroups) != 0 || len(r.PolicyRule.Resources) != 0 || len(r.PolicyRule.ResourceNames) != 0 {
			return PolicyRule{}, fmt.Errorf("non-resource rule may not have apiGroups, resources, or resourceNames: %#v", r.PolicyRule)
		}
	case len(r.PolicyRule.Resources) > 0:
		if len(r.PolicyRule.NonResourceURLs) != 0 {
			return PolicyRule{}, fmt.Errorf("resource rule may not have nonResourceURLs: %#v", r.PolicyRule)
		}
		if len(r.PolicyRule.APIGroups) == 0 {
			return PolicyRule{}, fmt.Errorf("resource rule must have apiGroups: %#v", r.PolicyRule)
		}
	default:
		return PolicyRule{}, fmt.Errorf("a rule must have either nonResourceURLs or resources: %#v", r.PolicyRule)
	}

	return r.PolicyRule, nil
}

type SortableRuleSlice []PolicyRule

func (s SortableRuleSlice) Len() int      { return len(s) }
func (s SortableRuleSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortableRuleSlice) Less(i, j int) bool {
	return strings.Compare(s[i].String(), s[j].String()) < 0
}
