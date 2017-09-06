package generictrigger

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
	kcorelister "k8s.io/kubernetes/pkg/client/listers/core/v1"

	"github.com/golang/glog"
	oscache "github.com/openshift/origin/pkg/client/cache"
	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	appsclient "github.com/openshift/origin/pkg/deploy/generated/internalclientset/typed/apps/internalversion"
)

const (
	// maxRetryCount is the number of times a deployment config will be retried before it is dropped
	// out of the queue.
	maxRetryCount = 15
)

// DeploymentTriggerController processes all triggers for a deployment config
// and kicks new deployments whenever possible.
type DeploymentTriggerController struct {
	// triggerFromImages is true if image changes should be processed by the instantiate
	// endpoint
	triggerFromImages bool

	// dn is used to update deployment configs.
	dn appsclient.DeploymentConfigsGetter

	// queue contains deployment configs that need to be synced.
	queue workqueue.RateLimitingInterface

	// dcLister provides a local cache for deployment configs.
	dcLister oscache.StoreToDeploymentConfigLister
	// dcListerSynced makes sure the dc store is synced before reconcling any deployment config.
	dcListerSynced func() bool
	// rcLister provides a local cache for replication controllers.
	rcLister kcorelister.ReplicationControllerLister
	// rcListerSynced makes sure the dc store is synced before reconcling any replication controller.
	rcListerSynced func() bool

	// codec is used for decoding a config out of a deployment.
	codec runtime.Codec
}

// Handle processes deployment triggers for a deployment config.
func (c *DeploymentTriggerController) Handle(config *deployapi.DeploymentConfig) error {
	if len(config.Spec.Triggers) == 0 || config.Spec.Paused {
		return nil
	}

	request := &deployapi.DeploymentRequest{
		Name:   config.Name,
		Latest: true,
		Force:  false,
	}
	if !c.triggerFromImages {
		request.ExcludeTriggers = []deployapi.DeploymentTriggerType{deployapi.DeploymentTriggerOnImageChange}
	}

	_, err := c.dn.DeploymentConfigs(config.Namespace).Instantiate(config.Name, request)
	return err
}

func (c *DeploymentTriggerController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetryCount {
		glog.V(2).Infof("Error instantiating deployment config %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	glog.V(2).Infof("Dropping deployment config %q out of the trigger queue: %v", key, err)
	c.queue.Forget(key)
}
