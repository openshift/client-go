package user

import "k8s.io/apimachinery/pkg/fields"

// IdentityToSelectableFields returns a label set that represents the object
// changes to the returned keys require registering conversions for existing versions using Scheme.AddFieldLabelConversionFunc
func IdentityToSelectableFields(identity *Identity) fields.Set {
	return fields.Set{
		"metadata.name":    identity.Name,
		"providerName":     identity.ProviderName,
		"providerUserName": identity.ProviderName,
		"user.name":        identity.User.Name,
		"user.uid":         string(identity.User.UID),
	}
}
