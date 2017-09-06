package install

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	kapi "k8s.io/kubernetes/pkg/api"

	projectapi "github.com/openshift/origin/pkg/project/apis/project"
	projectapiv1 "github.com/openshift/origin/pkg/project/apis/project/v1"
)

func installApiGroup() {
	Install(kapi.GroupFactoryRegistry, kapi.Registry, kapi.Scheme)
}

// Install registers the API group and adds types to a scheme
func Install(groupFactoryRegistry announced.APIGroupFactoryRegistry, registry *registered.APIRegistrationManager, scheme *runtime.Scheme) {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:                  projectapi.GroupName,
			VersionPreferenceOrder:     []string{projectapiv1.SchemeGroupVersion.Version},
			ImportPrefix:               importPrefix,
			AddInternalObjectsToScheme: projectapi.AddToScheme,
			RootScopedKinds:            sets.NewString("Project", "ProjectRequest"),
		},
		announced.VersionToSchemeFunc{
			projectapiv1.SchemeGroupVersion.Version: projectapiv1.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme); err != nil {
		panic(err)
	}
}
