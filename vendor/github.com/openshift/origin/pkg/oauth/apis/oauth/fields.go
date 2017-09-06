package oauth

import "k8s.io/apimachinery/pkg/fields"

// OAuthAccessTokenToSelectableFields returns a label set that represents the object
func OAuthAccessTokenToSelectableFields(obj *OAuthAccessToken) fields.Set {
	return fields.Set{
		"metadata.name":  obj.Name,
		"clientName":     obj.ClientName,
		"userName":       obj.UserName,
		"userUID":        obj.UserUID,
		"authorizeToken": obj.AuthorizeToken,
	}
}

// OAuthAuthorizeTokenToSelectableFields returns a label set that represents the object
func OAuthAuthorizeTokenToSelectableFields(obj *OAuthAuthorizeToken) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
		"clientName":    obj.ClientName,
		"userName":      obj.UserName,
		"userUID":       obj.UserUID,
	}
}

// OAuthClientAuthorizationToSelectableFields returns a label set that represents the object
func OAuthClientAuthorizationToSelectableFields(obj *OAuthClientAuthorization) fields.Set {
	return fields.Set{
		"metadata.name": obj.Name,
		"clientName":    obj.ClientName,
		"userName":      obj.UserName,
		"userUID":       obj.UserUID,
	}
}
