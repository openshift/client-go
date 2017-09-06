package allocation

import (
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"

	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/route"
)

// RouteAllocationControllerFactory creates a RouteAllocationController
// that allocates router shards to specific routes.
type RouteAllocationControllerFactory struct {
	// OSClient is is an OpenShift client.
	OSClient osclient.Interface

	// KubeClient is a Kubernetes client.
	KubeClient kclientset.Interface
}

// Create a RouteAllocationController instance.
func (factory *RouteAllocationControllerFactory) Create(plugin route.AllocationPlugin) *RouteAllocationController {
	return &RouteAllocationController{Plugin: plugin}
}
