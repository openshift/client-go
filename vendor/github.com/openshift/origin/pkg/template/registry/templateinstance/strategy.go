package templateinstance

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/authentication/user"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/storage/names"
	kapi "k8s.io/kubernetes/pkg/api"
	kapihelper "k8s.io/kubernetes/pkg/api/helper"
	"k8s.io/kubernetes/pkg/apis/authorization"
	authorizationinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/authorization/internalversion"
	rbacregistry "k8s.io/kubernetes/pkg/registry/rbac"

	"github.com/openshift/origin/pkg/authorization/util"
	template "github.com/openshift/origin/pkg/template"
	templateapi "github.com/openshift/origin/pkg/template/apis/template"
	"github.com/openshift/origin/pkg/template/apis/template/validation"
)

// templateInstanceStrategy implements behavior for TemplateInstances
type templateInstanceStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
	authorizationClient authorizationinternalversion.AuthorizationInterface
}

func NewStrategy(authorizationClient authorizationinternalversion.AuthorizationInterface) *templateInstanceStrategy {
	return &templateInstanceStrategy{kapi.Scheme, names.SimpleNameGenerator, authorizationClient}
}

// NamespaceScoped is true for templateinstances.
func (templateInstanceStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (templateInstanceStrategy) PrepareForUpdate(ctx apirequest.Context, obj, old runtime.Object) {
	curr := obj.(*templateapi.TemplateInstance)
	prev := old.(*templateapi.TemplateInstance)

	curr.Status = prev.Status
}

// Canonicalize normalizes the object after validation.
func (templateInstanceStrategy) Canonicalize(obj runtime.Object) {
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (templateInstanceStrategy) PrepareForCreate(ctx apirequest.Context, obj runtime.Object) {
	templateInstance := obj.(*templateapi.TemplateInstance)

	//TODO - when https://github.com/kubernetes-incubator/service-catalog/pull/939 sufficiently progresses, remove the if nil check and always
	// Clear any user-specified info, then seed it from the context
	if templateInstance.Spec.Requester == nil {

		if user, ok := apirequest.UserFrom(ctx); ok {
			templateReq := template.ConvertUserToTemplateInstanceRequester(user)
			templateInstance.Spec.Requester = &templateReq
		}
	}

	templateInstance.Status = templateapi.TemplateInstanceStatus{}
}

// Validate validates a new templateinstance.
func (s *templateInstanceStrategy) Validate(ctx apirequest.Context, obj runtime.Object) field.ErrorList {
	user, ok := apirequest.UserFrom(ctx)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath(""), errors.New("user not found in context"))}
	}

	templateInstance := obj.(*templateapi.TemplateInstance)
	allErrs := validation.ValidateTemplateInstance(templateInstance)
	allErrs = append(allErrs, s.validateImpersonation(templateInstance, user)...)

	return allErrs
}

// AllowCreateOnUpdate is false for templateinstances.
func (templateInstanceStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (templateInstanceStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (s *templateInstanceStrategy) ValidateUpdate(ctx apirequest.Context, obj, old runtime.Object) field.ErrorList {
	user, ok := apirequest.UserFrom(ctx)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath(""), errors.New("user not found in context"))}
	}

	templateInstance := obj.(*templateapi.TemplateInstance)
	oldTemplateInstance := old.(*templateapi.TemplateInstance)
	allErrs := validation.ValidateTemplateInstanceUpdate(templateInstance, oldTemplateInstance)
	allErrs = append(allErrs, s.validateImpersonationUpdate(templateInstance, oldTemplateInstance, user)...)

	return allErrs
}

func (s *templateInstanceStrategy) validateImpersonationUpdate(templateInstance, oldTemplateInstance *templateapi.TemplateInstance, userinfo user.Info) field.ErrorList {
	if rbacregistry.IsOnlyMutatingGCFields(templateInstance, oldTemplateInstance, kapihelper.Semantic) {
		return nil
	}

	return s.validateImpersonation(templateInstance, userinfo)
}

func (s *templateInstanceStrategy) validateImpersonation(templateInstance *templateapi.TemplateInstance, userinfo user.Info) field.ErrorList {
	if templateInstance.Spec.Requester == nil || templateInstance.Spec.Requester.Username == "" {
		return field.ErrorList{field.Required(field.NewPath("spec.requester.username"), "")}
	}

	if templateInstance.Spec.Requester.Username != userinfo.GetName() {
		if err := util.Authorize(s.authorizationClient.SubjectAccessReviews(), userinfo, &authorization.ResourceAttributes{
			Namespace: templateInstance.Namespace,
			Verb:      "assign",
			Group:     templateapi.GroupName,
			Resource:  "templateinstances",
			Name:      templateInstance.Name,
		}); err != nil {
			return field.ErrorList{field.Forbidden(field.NewPath("spec.requester.username"), "you do not have permission to set username")}
		}
	}

	return nil
}

type statusStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var StatusStrategy = statusStrategy{kapi.Scheme, names.SimpleNameGenerator}

func (statusStrategy) NamespaceScoped() bool {
	return true
}

func (statusStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (statusStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (statusStrategy) PrepareForUpdate(ctx apirequest.Context, obj, old runtime.Object) {
	curr := obj.(*templateapi.TemplateInstance)
	prev := old.(*templateapi.TemplateInstance)

	curr.Spec = prev.Spec
}

func (statusStrategy) Canonicalize(obj runtime.Object) {
}

func (statusStrategy) ValidateUpdate(ctx apirequest.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateTemplateInstanceUpdate(obj.(*templateapi.TemplateInstance), old.(*templateapi.TemplateInstance))
}
