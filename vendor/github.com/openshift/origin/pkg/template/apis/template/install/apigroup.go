package install

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	kapi "k8s.io/kubernetes/pkg/api"

	templateapi "github.com/openshift/origin/pkg/template/apis/template"
	templateapiv1 "github.com/openshift/origin/pkg/template/apis/template/v1"
)

func installApiGroup() {
	Install(kapi.GroupFactoryRegistry, kapi.Registry, kapi.Scheme)
}

// Install registers the API group and adds types to a scheme
func Install(groupFactoryRegistry announced.APIGroupFactoryRegistry, registry *registered.APIRegistrationManager, scheme *runtime.Scheme) {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:                  templateapi.GroupName,
			VersionPreferenceOrder:     []string{templateapiv1.LegacySchemeGroupVersion.Version},
			ImportPrefix:               importPrefix,
			AddInternalObjectsToScheme: templateapi.AddToScheme,
			RootScopedKinds:            sets.NewString("BrokerTemplateInstance"),
		},
		announced.VersionToSchemeFunc{
			templateapiv1.LegacySchemeGroupVersion.Version: templateapiv1.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme); err != nil {
		panic(err)
	}
}
