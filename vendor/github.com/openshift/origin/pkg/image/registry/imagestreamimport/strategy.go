package imagestreamimport

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	kapi "k8s.io/kubernetes/pkg/api"

	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	serverapi "github.com/openshift/origin/pkg/cmd/server/api"
	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	"github.com/openshift/origin/pkg/image/apis/image/validation"
)

// strategy implements behavior for ImageStreamImports.
type strategy struct {
	runtime.ObjectTyper
	allowedRegistries     *serverapi.AllowedRegistries
	registryHostRetriever imageapi.RegistryHostnameRetriever
}

func NewStrategy(registries *serverapi.AllowedRegistries, registry imageapi.RegistryHostnameRetriever) *strategy {
	return &strategy{
		ObjectTyper:           kapi.Scheme,
		allowedRegistries:     registries,
		registryHostRetriever: registry,
	}
}

func (s *strategy) NamespaceScoped() bool {
	return true
}

func (s *strategy) GenerateName(string) string {
	return ""
}

func (s *strategy) Canonicalize(runtime.Object) {
}

func (s *strategy) ValidateAllowedRegistries(isi *imageapi.ImageStreamImport) field.ErrorList {
	errs := field.ErrorList{}
	if s.allowedRegistries == nil {
		return errs
	}
	allowedRegistries := *s.allowedRegistries
	if localRegistry, ok := s.registryHostRetriever.InternalRegistryHostname(); ok {
		allowedRegistries = append([]configapi.RegistryLocation{{DomainName: localRegistry}}, allowedRegistries...)
	}
	validate := func(path *field.Path, name string, insecure bool) field.ErrorList {
		ref, _ := imageapi.ParseDockerImageReference(name)
		registryHost, registryPort := ref.RegistryHostPort(insecure)
		return validation.ValidateRegistryAllowedForImport(path.Child("from", "name"), ref.Name, registryHost, registryPort, &allowedRegistries)
	}
	if spec := isi.Spec.Repository; spec != nil && spec.From.Kind == "DockerImage" {
		errs = append(errs, validate(field.NewPath("spec").Child("repository"), spec.From.Name, spec.ImportPolicy.Insecure)...)
	}
	if len(isi.Spec.Images) > 0 {
		for i, image := range isi.Spec.Images {
			errs = append(errs, validate(field.NewPath("spec").Child("images").Index(i), image.From.Name, image.ImportPolicy.Insecure)...)
		}
	}
	return errs
}

func (s *strategy) PrepareForCreate(ctx apirequest.Context, obj runtime.Object) {
	newIST := obj.(*imageapi.ImageStreamImport)
	newIST.Status = imageapi.ImageStreamImportStatus{}
}

func (s *strategy) PrepareImageForCreate(obj runtime.Object) {
	image := obj.(*imageapi.Image)

	// signatures can be added using "images" or "imagesignatures" resources
	image.Signatures = nil

	// Remove the raw manifest as it's very big and this leads to a large memory consumption in etcd.
	image.DockerImageManifest = ""
	image.DockerImageConfig = ""
}

func (s *strategy) Validate(ctx apirequest.Context, obj runtime.Object) field.ErrorList {
	isi := obj.(*imageapi.ImageStreamImport)
	return validation.ValidateImageStreamImport(isi)
}
