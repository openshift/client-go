package template

// annotation keys
const (
	// IconClassAnnotation is the rss class of an icon representing a template
	IconClassAnnotation = "iconClass"

	// LongDescriptionAnnotation is a template's long description
	LongDescriptionAnnotation = "template.openshift.io/long-description"

	// ProviderDisplayNameAnnotation is the name of a template provider, e.g.
	// "Red Hat, Inc."
	ProviderDisplayNameAnnotation = "template.openshift.io/provider-display-name"

	// DocumentationURLAnnotation is the url where documentation associated with
	// a template can be found
	DocumentationURLAnnotation = "template.openshift.io/documentation-url"

	// SupportURLAnnotation is the url where support for a template can be found
	SupportURLAnnotation = "template.openshift.io/support-url"

	// TemplateInstanceLabel is used to label every object created by the
	// TemplateInstance API.
	TemplateInstanceLabel = "template.openshift.io/template-instance"

	//TODO - when https://github.com/kubernetes-incubator/service-catalog/pull/939 sufficiently progresses, remove this constant
	// RequesterUsernameParameterKey is the name of the key in the Open Service
	// Broker API ProvisionRequest Parameters object where we receive the user
	// name which will be impersonated during template provisioning.  The '/'
	// and '.' characters in the name happen to make this an invalid template
	// parameter name so there is no immediate overlap with passed template
	// parameters in the same object.
	RequesterUsernameParameterKey = "template.openshift.io/requester-username"

	// ServiceBrokerRoot is the API root of the template service broker.
	ServiceBrokerRoot = "/brokers/template.openshift.io"

	// ServiceMetadataIconClass is the key for the template iconClass as returned
	// in the services.metadata map from a service broker catalog response
	ServiceMetadataIconClass = "console.openshift.io/iconClass"

	// TemplateUIDIndex is the name of an index on the generated template lister,
	// initialised and used by the template service broker.
	TemplateUIDIndex = "templateuid"

	// ExposeAnnotationPrefix indicates that part of an object in a template
	// should be exposed in some way, for example implying that it should be
	// returned by the template service broker in the results of a bind call.
	// The rest of the annotation name following the prefix may be used by the
	// exposer as a key name.  The annotation value is a Kubernetes JSONPath
	// template expression which the exposer uses to calculate the exposed
	// value.  JSONPath expressions which return multiple and/or complex objects
	// are not permitted (with the exception of []byte, which is permitted).
	// Any []byte values returned are converted to strings.
	ExposeAnnotationPrefix = "template.openshift.io/expose-"

	// Base64ExposeAnnotationPrefix is as ExposeAnnotationPrefix, except that
	// any []byte values returned are base64 encoded.
	Base64ExposeAnnotationPrefix = "template.openshift.io/base64-expose-"

	// WaitForReadyAnnotation indicates that the TemplateInstance controller
	// should wait for the object to be ready before reporting the template
	// instantiation complete.
	WaitForReadyAnnotation = "template.alpha.openshift.io/wait-for-ready"
)
