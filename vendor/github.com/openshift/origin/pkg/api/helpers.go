package api

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation/path"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/api/validation"
)

// GetNameValidationFunc returns a name validation function that includes the standard restrictions we want for all types
func GetNameValidationFunc(nameFunc validation.ValidateNameFunc) validation.ValidateNameFunc {
	return func(name string, prefix bool) []string {
		if reasons := path.ValidatePathSegmentName(name, prefix); len(reasons) != 0 {
			return reasons
		}

		return nameFunc(name, prefix)
	}
}

// GetFieldLabelConversionFunc returns a field label conversion func, which does the following:
// * returns overrideLabels[label], value, nil if the specified label exists in the overrideLabels map
// * returns label, value, nil if the specified label exists as a key in the supportedLabels map (values in this map are unused, it is intended to be a prototypical label/value map)
// * otherwise, returns an error
func GetFieldLabelConversionFunc(supportedLabels map[string]string, overrideLabels map[string]string) func(label, value string) (string, string, error) {
	return func(label, value string) (string, string, error) {
		if label, overridden := overrideLabels[label]; overridden {
			return label, value, nil
		}
		if _, supported := supportedLabels[label]; supported {
			return label, value, nil
		}
		return "", "", fmt.Errorf("field label not supported: %s", label)
	}
}

// GetResourceKey returns a string of the form [namespace]/[name] for
// the given resource.  This is a common way of ensuring a key for a
// resource that is unique across the cluster.
func GetResourceKey(obj metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)
}

// OriginObjects returns the origin types.
func OriginObjects(objs []runtime.Object) []runtime.Object {
	ret := []runtime.Object{}
	for _, obj := range objs {
		if !upstreamType(obj) {
			ret = append(ret, obj)
		}
	}
	return ret
}

// UpstreamObjects returns the non-origin types.
func UpstreamObjects(objs []runtime.Object) []runtime.Object {
	ret := []runtime.Object{}
	for _, obj := range objs {
		if upstreamType(obj) {
			ret = append(ret, obj)
		}
	}
	return ret
}

// upstreamType returns true for Kubernetes types.
func upstreamType(obj runtime.Object) bool {
	t := reflect.TypeOf(obj).Elem()
	return strings.Contains(t.PkgPath(), "k8s.io/kubernetes/")
}
