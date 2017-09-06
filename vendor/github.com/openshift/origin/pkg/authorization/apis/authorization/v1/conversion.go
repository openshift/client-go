package v1

import (
	"sort"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	oapi "github.com/openshift/origin/pkg/api"
	"github.com/openshift/origin/pkg/api/extension"
	newer "github.com/openshift/origin/pkg/authorization/apis/authorization"
	uservalidation "github.com/openshift/origin/pkg/user/apis/user/validation"
)

func Convert_v1_SubjectAccessReview_To_authorization_SubjectAccessReview(in *SubjectAccessReview, out *newer.SubjectAccessReview, s conversion.Scope) error {
	if err := autoConvert_v1_SubjectAccessReview_To_authorization_SubjectAccessReview(in, out, s); err != nil {
		return err
	}

	out.Groups = sets.NewString(in.GroupsSlice...)
	out.Scopes = []string(in.Scopes)
	return nil
}

func Convert_authorization_SubjectAccessReview_To_v1_SubjectAccessReview(in *newer.SubjectAccessReview, out *SubjectAccessReview, s conversion.Scope) error {
	if err := autoConvert_authorization_SubjectAccessReview_To_v1_SubjectAccessReview(in, out, s); err != nil {
		return err
	}

	out.GroupsSlice = in.Groups.List()
	out.Scopes = OptionalScopes(in.Scopes)
	return nil
}

func Convert_v1_SelfSubjectRulesReviewSpec_To_authorization_SelfSubjectRulesReviewSpec(in *SelfSubjectRulesReviewSpec, out *newer.SelfSubjectRulesReviewSpec, s conversion.Scope) error {
	if err := autoConvert_v1_SelfSubjectRulesReviewSpec_To_authorization_SelfSubjectRulesReviewSpec(in, out, s); err != nil {
		return err
	}

	out.Scopes = []string(in.Scopes)
	return nil
}

func Convert_authorization_SelfSubjectRulesReviewSpec_To_v1_SelfSubjectRulesReviewSpec(in *newer.SelfSubjectRulesReviewSpec, out *SelfSubjectRulesReviewSpec, s conversion.Scope) error {
	if err := autoConvert_authorization_SelfSubjectRulesReviewSpec_To_v1_SelfSubjectRulesReviewSpec(in, out, s); err != nil {
		return err
	}

	out.Scopes = OptionalScopes(in.Scopes)
	return nil
}

func Convert_v1_LocalSubjectAccessReview_To_authorization_LocalSubjectAccessReview(in *LocalSubjectAccessReview, out *newer.LocalSubjectAccessReview, s conversion.Scope) error {
	if err := autoConvert_v1_LocalSubjectAccessReview_To_authorization_LocalSubjectAccessReview(in, out, s); err != nil {
		return err
	}

	out.Groups = sets.NewString(in.GroupsSlice...)
	out.Scopes = []string(in.Scopes)
	return nil
}

func Convert_authorization_LocalSubjectAccessReview_To_v1_LocalSubjectAccessReview(in *newer.LocalSubjectAccessReview, out *LocalSubjectAccessReview, s conversion.Scope) error {
	if err := autoConvert_authorization_LocalSubjectAccessReview_To_v1_LocalSubjectAccessReview(in, out, s); err != nil {
		return err
	}

	out.GroupsSlice = in.Groups.List()
	out.Scopes = OptionalScopes(in.Scopes)
	return nil
}

func Convert_v1_ResourceAccessReviewResponse_To_authorization_ResourceAccessReviewResponse(in *ResourceAccessReviewResponse, out *newer.ResourceAccessReviewResponse, s conversion.Scope) error {
	if err := autoConvert_v1_ResourceAccessReviewResponse_To_authorization_ResourceAccessReviewResponse(in, out, s); err != nil {
		return err
	}

	out.Users = sets.NewString(in.UsersSlice...)
	out.Groups = sets.NewString(in.GroupsSlice...)
	return nil
}

func Convert_authorization_ResourceAccessReviewResponse_To_v1_ResourceAccessReviewResponse(in *newer.ResourceAccessReviewResponse, out *ResourceAccessReviewResponse, s conversion.Scope) error {
	if err := autoConvert_authorization_ResourceAccessReviewResponse_To_v1_ResourceAccessReviewResponse(in, out, s); err != nil {
		return err
	}

	out.UsersSlice = in.Users.List()
	out.GroupsSlice = in.Groups.List()
	return nil
}

func Convert_v1_PolicyRule_To_authorization_PolicyRule(in *PolicyRule, out *newer.PolicyRule, s conversion.Scope) error {
	SetDefaults_PolicyRule(in)
	if err := oapi.Convert_runtime_RawExtension_To_runtime_Object(&in.AttributeRestrictions, &out.AttributeRestrictions, s); err != nil {
		return err
	}

	out.APIGroups = in.APIGroups

	out.Resources = sets.String{}
	out.Resources.Insert(in.Resources...)

	out.Verbs = sets.String{}
	out.Verbs.Insert(in.Verbs...)

	out.ResourceNames = sets.NewString(in.ResourceNames...)

	out.NonResourceURLs = sets.NewString(in.NonResourceURLsSlice...)

	return nil
}

func Convert_authorization_PolicyRule_To_v1_PolicyRule(in *newer.PolicyRule, out *PolicyRule, s conversion.Scope) error {
	if err := oapi.Convert_runtime_Object_To_runtime_RawExtension(&in.AttributeRestrictions, &out.AttributeRestrictions, s); err != nil {
		return err
	}

	out.APIGroups = in.APIGroups

	out.Resources = []string{}
	out.Resources = append(out.Resources, in.Resources.List()...)

	out.Verbs = []string{}
	out.Verbs = append(out.Verbs, in.Verbs.List()...)

	out.ResourceNames = in.ResourceNames.List()

	out.NonResourceURLsSlice = in.NonResourceURLs.List()

	return nil
}

func Convert_v1_Policy_To_authorization_Policy(in *Policy, out *newer.Policy, s conversion.Scope) error {
	if err := autoConvert_v1_Policy_To_authorization_Policy(in, out, s); err != nil {
		return err
	}
	if out.Roles == nil {
		out.Roles = make(map[string]*newer.Role)
	}
	return nil
}

func Convert_v1_RoleBinding_To_authorization_RoleBinding(in *RoleBinding, out *newer.RoleBinding, s conversion.Scope) error {
	if err := autoConvert_v1_RoleBinding_To_authorization_RoleBinding(in, out, s); err != nil {
		return err
	}

	// if the users and groups fields are cleared, then respect only subjects.  The field was set in the DefaultConvert above
	if in.UserNames == nil && in.GroupNames == nil {
		return nil
	}

	out.Subjects = newer.BuildSubjects(in.UserNames, in.GroupNames, uservalidation.ValidateUserName, uservalidation.ValidateGroupName)

	return nil
}

func Convert_authorization_RoleBinding_To_v1_RoleBinding(in *newer.RoleBinding, out *RoleBinding, s conversion.Scope) error {
	if err := autoConvert_authorization_RoleBinding_To_v1_RoleBinding(in, out, s); err != nil {
		return err
	}

	out.UserNames, out.GroupNames = newer.StringSubjectsFor(in.Namespace, in.Subjects)

	return nil
}

func Convert_v1_PolicyBinding_To_authorization_PolicyBinding(in *PolicyBinding, out *newer.PolicyBinding, s conversion.Scope) error {
	if err := autoConvert_v1_PolicyBinding_To_authorization_PolicyBinding(in, out, s); err != nil {
		return err
	}
	if out.RoleBindings == nil {
		out.RoleBindings = make(map[string]*newer.RoleBinding)
	}
	return nil
}

// and now the globals
func Convert_v1_ClusterPolicy_To_authorization_ClusterPolicy(in *ClusterPolicy, out *newer.ClusterPolicy, s conversion.Scope) error {
	if err := autoConvert_v1_ClusterPolicy_To_authorization_ClusterPolicy(in, out, s); err != nil {
		return err
	}
	if out.Roles == nil {
		out.Roles = make(map[string]*newer.ClusterRole)
	}
	return nil
}

func Convert_v1_ClusterRoleBinding_To_authorization_ClusterRoleBinding(in *ClusterRoleBinding, out *newer.ClusterRoleBinding, s conversion.Scope) error {
	if err := autoConvert_v1_ClusterRoleBinding_To_authorization_ClusterRoleBinding(in, out, s); err != nil {
		return err
	}

	// if the users and groups fields are cleared, then respect only subjects.  The field was set in the DefaultConvert above
	if in.UserNames == nil && in.GroupNames == nil {
		return nil
	}

	out.Subjects = newer.BuildSubjects(in.UserNames, in.GroupNames, uservalidation.ValidateUserName, uservalidation.ValidateGroupName)

	return nil
}

func Convert_authorization_ClusterRoleBinding_To_v1_ClusterRoleBinding(in *newer.ClusterRoleBinding, out *ClusterRoleBinding, s conversion.Scope) error {
	if err := autoConvert_authorization_ClusterRoleBinding_To_v1_ClusterRoleBinding(in, out, s); err != nil {
		return err
	}

	out.UserNames, out.GroupNames = newer.StringSubjectsFor(in.Namespace, in.Subjects)

	return nil
}

func Convert_v1_ClusterPolicyBinding_To_authorization_ClusterPolicyBinding(in *ClusterPolicyBinding, out *newer.ClusterPolicyBinding, s conversion.Scope) error {
	if err := autoConvert_v1_ClusterPolicyBinding_To_authorization_ClusterPolicyBinding(in, out, s); err != nil {
		return err
	}
	if out.RoleBindings == nil {
		out.RoleBindings = make(map[string]*newer.ClusterRoleBinding)
	}
	return nil
}

func Convert_v1_NamedRoles_To_authorization_RolesByName(in *NamedRoles, out *newer.RolesByName, s conversion.Scope) error {
	if *out == nil {
		*out = make(newer.RolesByName)
	}

	for _, curr := range *in {
		newRole := &newer.Role{}
		if err := Convert_v1_Role_To_authorization_Role(&curr.Role, newRole, s); err != nil {
			return err
		}
		(*out)[curr.Name] = newRole
	}

	return nil
}
func Convert_authorization_RolesByName_To_v1_NamedRoles(in *newer.RolesByName, out *NamedRoles, s conversion.Scope) error {
	allKeys := make([]string, 0, len(*in))
	for key := range *in {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)

	for _, key := range allKeys {
		newRole := (*in)[key]
		oldRole := &Role{}
		if err := Convert_authorization_Role_To_v1_Role(newRole, oldRole, s); err != nil {
			return err
		}

		namedRole := NamedRole{key, *oldRole}
		*out = append(*out, namedRole)
	}

	return nil
}

func Convert_v1_NamedRoleBindings_To_authorization_RoleBindingsByName(in *NamedRoleBindings, out *newer.RoleBindingsByName, s conversion.Scope) error {
	if *out == nil {
		*out = make(newer.RoleBindingsByName)
	}
	for _, curr := range *in {
		newRoleBinding := &newer.RoleBinding{}
		if err := Convert_v1_RoleBinding_To_authorization_RoleBinding(&curr.RoleBinding, newRoleBinding, s); err != nil {
			return err
		}
		(*out)[curr.Name] = newRoleBinding
	}

	return nil
}
func Convert_authorization_RoleBindingsByName_To_v1_NamedRoleBindings(in *newer.RoleBindingsByName, out *NamedRoleBindings, s conversion.Scope) error {
	allKeys := make([]string, 0, len(*in))
	for key := range *in {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)

	for _, key := range allKeys {
		newRoleBinding := (*in)[key]
		oldRoleBinding := &RoleBinding{}
		if err := Convert_authorization_RoleBinding_To_v1_RoleBinding(newRoleBinding, oldRoleBinding, s); err != nil {
			return err
		}

		namedRoleBinding := NamedRoleBinding{key, *oldRoleBinding}
		*out = append(*out, namedRoleBinding)
	}

	return nil
}

func Convert_v1_NamedClusterRoles_To_authorization_ClusterRolesByName(in *NamedClusterRoles, out *newer.ClusterRolesByName, s conversion.Scope) error {
	if *out == nil {
		*out = make(newer.ClusterRolesByName)
	}
	for _, curr := range *in {
		newRole := &newer.ClusterRole{}
		if err := Convert_v1_ClusterRole_To_authorization_ClusterRole(&curr.Role, newRole, s); err != nil {
			return err
		}
		(*out)[curr.Name] = newRole
	}

	return nil
}
func Convert_authorization_ClusterRolesByName_To_v1_NamedClusterRoles(in *newer.ClusterRolesByName, out *NamedClusterRoles, s conversion.Scope) error {
	allKeys := make([]string, 0, len(*in))
	for key := range *in {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)

	for _, key := range allKeys {
		newRole := (*in)[key]
		oldRole := &ClusterRole{}
		if err := Convert_authorization_ClusterRole_To_v1_ClusterRole(newRole, oldRole, s); err != nil {
			return err
		}

		namedRole := NamedClusterRole{key, *oldRole}
		*out = append(*out, namedRole)
	}

	return nil
}
func Convert_v1_NamedClusterRoleBindings_To_authorization_ClusterRoleBindingsByName(in *NamedClusterRoleBindings, out *newer.ClusterRoleBindingsByName, s conversion.Scope) error {
	if *out == nil {
		*out = make(newer.ClusterRoleBindingsByName)
	}
	for _, curr := range *in {
		newRoleBinding := &newer.ClusterRoleBinding{}
		if err := Convert_v1_ClusterRoleBinding_To_authorization_ClusterRoleBinding(&curr.RoleBinding, newRoleBinding, s); err != nil {
			return err
		}
		(*out)[curr.Name] = newRoleBinding
	}
	return nil
}
func Convert_authorization_ClusterRoleBindingsByName_To_v1_NamedClusterRoleBindings(in *newer.ClusterRoleBindingsByName, out *NamedClusterRoleBindings, s conversion.Scope) error {
	allKeys := make([]string, 0, len(*in))
	for key := range *in {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)

	for _, key := range allKeys {
		newRoleBinding := (*in)[key]
		oldRoleBinding := &ClusterRoleBinding{}
		if err := Convert_authorization_ClusterRoleBinding_To_v1_ClusterRoleBinding(newRoleBinding, oldRoleBinding, s); err != nil {
			return err
		}

		namedRoleBinding := NamedClusterRoleBinding{key, *oldRoleBinding}
		*out = append(*out, namedRoleBinding)
	}

	return nil
}

func addConversionFuncs(scheme *runtime.Scheme) error {
	err := scheme.AddConversionFuncs(
		Convert_v1_SubjectAccessReview_To_authorization_SubjectAccessReview,
		Convert_authorization_SubjectAccessReview_To_v1_SubjectAccessReview,
		Convert_v1_LocalSubjectAccessReview_To_authorization_LocalSubjectAccessReview,
		Convert_authorization_LocalSubjectAccessReview_To_v1_LocalSubjectAccessReview,
		Convert_v1_ResourceAccessReview_To_authorization_ResourceAccessReview,
		Convert_authorization_ResourceAccessReview_To_v1_ResourceAccessReview,
		Convert_v1_LocalResourceAccessReview_To_authorization_LocalResourceAccessReview,
		Convert_authorization_LocalResourceAccessReview_To_v1_LocalResourceAccessReview,
		Convert_v1_ResourceAccessReviewResponse_To_authorization_ResourceAccessReviewResponse,
		Convert_authorization_ResourceAccessReviewResponse_To_v1_ResourceAccessReviewResponse,
		Convert_v1_PolicyRule_To_authorization_PolicyRule,
		Convert_authorization_PolicyRule_To_v1_PolicyRule,
		Convert_v1_Policy_To_authorization_Policy,
		Convert_authorization_Policy_To_v1_Policy,
		Convert_v1_RoleBinding_To_authorization_RoleBinding,
		Convert_authorization_RoleBinding_To_v1_RoleBinding,
		Convert_v1_PolicyBinding_To_authorization_PolicyBinding,
		Convert_authorization_PolicyBinding_To_v1_PolicyBinding,
		Convert_v1_ClusterPolicy_To_authorization_ClusterPolicy,
		Convert_authorization_ClusterPolicy_To_v1_ClusterPolicy,
		Convert_v1_ClusterRoleBinding_To_authorization_ClusterRoleBinding,
		Convert_authorization_ClusterRoleBinding_To_v1_ClusterRoleBinding,
		Convert_v1_ClusterPolicyBinding_To_authorization_ClusterPolicyBinding,
		Convert_authorization_ClusterPolicyBinding_To_v1_ClusterPolicyBinding,
	)
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		return err
	}

	if err := scheme.AddFieldLabelConversionFunc("v1", "PolicyBinding",
		oapi.GetFieldLabelConversionFunc(newer.PolicyBindingToSelectableFields(&newer.PolicyBinding{}), nil),
	); err != nil {
		return err
	}

	return nil
}

var _ runtime.NestedObjectDecoder = &PolicyRule{}
var _ runtime.NestedObjectEncoder = &PolicyRule{}

func (c *PolicyRule) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	extension.DecodeNestedRawExtensionOrUnknown(d, &c.AttributeRestrictions)
	return nil
}
func (c *PolicyRule) EncodeNestedObjects(e runtime.Encoder) error {
	return extension.EncodeNestedRawExtension(e, &c.AttributeRestrictions)
}

var _ runtime.NestedObjectDecoder = &SelfSubjectRulesReview{}
var _ runtime.NestedObjectEncoder = &SelfSubjectRulesReview{}

func (c *SelfSubjectRulesReview) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Status.Rules {
		c.Status.Rules[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *SelfSubjectRulesReview) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Status.Rules {
		if err := c.Status.Rules[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &SubjectRulesReview{}
var _ runtime.NestedObjectEncoder = &SubjectRulesReview{}

func (c *SubjectRulesReview) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Status.Rules {
		c.Status.Rules[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *SubjectRulesReview) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Status.Rules {
		if err := c.Status.Rules[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &ClusterRole{}
var _ runtime.NestedObjectEncoder = &ClusterRole{}

func (c *ClusterRole) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Rules {
		c.Rules[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *ClusterRole) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Rules {
		if err := c.Rules[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &Role{}
var _ runtime.NestedObjectEncoder = &Role{}

func (c *Role) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Rules {
		c.Rules[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *Role) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Rules {
		if err := c.Rules[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &ClusterRoleList{}
var _ runtime.NestedObjectEncoder = &ClusterRoleList{}

func (c *ClusterRoleList) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Items {
		c.Items[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *ClusterRoleList) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Items {
		if err := c.Items[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &RoleList{}
var _ runtime.NestedObjectEncoder = &RoleList{}

func (c *RoleList) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Items {
		c.Items[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *RoleList) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Items {
		if err := c.Items[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &ClusterPolicy{}
var _ runtime.NestedObjectEncoder = &ClusterPolicy{}

func (c *ClusterPolicy) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Roles {
		c.Roles[i].Role.DecodeNestedObjects(d)
	}
	return nil
}
func (c *ClusterPolicy) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Roles {
		if err := c.Roles[i].Role.EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &Policy{}
var _ runtime.NestedObjectEncoder = &Policy{}

func (c *Policy) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Roles {
		c.Roles[i].Role.DecodeNestedObjects(d)
	}
	return nil
}
func (c *Policy) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Roles {
		if err := c.Roles[i].Role.EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &ClusterPolicyList{}
var _ runtime.NestedObjectEncoder = &ClusterPolicyList{}

func (c *ClusterPolicyList) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Items {
		c.Items[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *ClusterPolicyList) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Items {
		if err := c.Items[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}

var _ runtime.NestedObjectDecoder = &PolicyList{}
var _ runtime.NestedObjectEncoder = &PolicyList{}

func (c *PolicyList) DecodeNestedObjects(d runtime.Decoder) error {
	// decoding failures result in a runtime.Unknown object being created in Object and passed
	// to conversion
	for i := range c.Items {
		c.Items[i].DecodeNestedObjects(d)
	}
	return nil
}
func (c *PolicyList) EncodeNestedObjects(e runtime.Encoder) error {
	for i := range c.Items {
		if err := c.Items[i].EncodeNestedObjects(e); err != nil {
			return err
		}
	}
	return nil
}
