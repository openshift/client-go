package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/api/validation"

	buildvalidation "github.com/openshift/origin/pkg/build/apis/build/validation"
	"github.com/openshift/origin/pkg/build/controller/build/overrides/api"
)

func ValidateBuildOverridesConfig(config *api.BuildOverridesConfig) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, buildvalidation.ValidateImageLabels(config.ImageLabels, field.NewPath("imageLabels"))...)
	allErrs = append(allErrs, buildvalidation.ValidateNodeSelector(config.NodeSelector, field.NewPath("nodeSelector"))...)
	allErrs = append(allErrs, validation.ValidateAnnotations(config.Annotations, field.NewPath("annotations"))...)

	return allErrs
}
