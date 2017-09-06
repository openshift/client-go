package controller

import (
	"math/rand"
	"time"

	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	quotacontroller "github.com/openshift/origin/pkg/quota/controller"
	"github.com/openshift/origin/pkg/quota/controller/clusterquotamapping"
	"github.com/openshift/origin/pkg/quota/controller/clusterquotareconciliation"
	"k8s.io/kubernetes/pkg/controller"
	kresourcequota "k8s.io/kubernetes/pkg/controller/resourcequota"

	"github.com/openshift/origin/pkg/quota"
)

func RunResourceQuotaManager(ctx ControllerContext) (bool, error) {
	concurrentResourceQuotaSyncs := int(ctx.KubeControllerContext.Options.ConcurrentResourceQuotaSyncs)
	resourceQuotaSyncPeriod := ctx.KubeControllerContext.Options.ResourceQuotaSyncPeriod.Duration
	replenishmentSyncPeriodFunc := calculateResyncPeriod(ctx.KubeControllerContext.Options.MinResyncPeriod.Duration)
	saName := "resourcequota-controller"

	resourceQuotaRegistry := quota.NewOriginQuotaRegistry(
		ctx.ImageInformers.Image().InternalVersion().ImageStreams(),
		ctx.ClientBuilder.DeprecatedOpenshiftClientOrDie(saName),
	)

	resourceQuotaControllerOptions := &kresourcequota.ResourceQuotaControllerOptions{
		QuotaClient:           ctx.ClientBuilder.ClientOrDie(saName).Core(),
		ResourceQuotaInformer: ctx.ExternalKubeInformers.Core().V1().ResourceQuotas(),
		ResyncPeriod:          controller.StaticResyncPeriodFunc(resourceQuotaSyncPeriod),
		Registry:              resourceQuotaRegistry,
		GroupKindsToReplenish: quota.AllEvaluatedGroupKinds,
		ControllerFactory: quotacontroller.NewAllResourceReplenishmentControllerFactory(
			ctx.ExternalKubeInformers,
			ctx.ImageInformers.Image().InternalVersion().ImageStreams(),
			ctx.ClientBuilder.DeprecatedOpenshiftClientOrDie(saName),
		),
		ReplenishmentResyncPeriod: replenishmentSyncPeriodFunc,
	}
	go kresourcequota.NewResourceQuotaController(resourceQuotaControllerOptions).Run(concurrentResourceQuotaSyncs, ctx.Stop)

	return true, nil
}

type ClusterQuotaReconciliationControllerConfig struct {
	DefaultResyncPeriod            time.Duration
	DefaultReplenishmentSyncPeriod time.Duration
}

func (c *ClusterQuotaReconciliationControllerConfig) RunController(ctx ControllerContext) (bool, error) {
	saName := bootstrappolicy.InfraClusterQuotaReconciliationControllerServiceAccountName
	resourceQuotaRegistry := quota.NewAllResourceQuotaRegistry(
		ctx.ExternalKubeInformers,
		ctx.ImageInformers.Image().InternalVersion().ImageStreams(),
		ctx.ClientBuilder.DeprecatedOpenshiftClientOrDie(saName),
		ctx.ClientBuilder.ClientOrDie(saName),
	)
	groupKindsToReplenish := quota.AllEvaluatedGroupKinds

	clusterQuotaMappingController := clusterquotamapping.NewClusterQuotaMappingController(
		ctx.ExternalKubeInformers.Core().V1().Namespaces(),
		ctx.QuotaInformers.Quota().InternalVersion().ClusterResourceQuotas())
	options := clusterquotareconciliation.ClusterQuotaReconcilationControllerOptions{
		ClusterQuotaInformer: ctx.QuotaInformers.Quota().InternalVersion().ClusterResourceQuotas(),
		ClusterQuotaMapper:   clusterQuotaMappingController.GetClusterQuotaMapper(),
		ClusterQuotaClient:   ctx.ClientBuilder.DeprecatedOpenshiftClientOrDie(saName),

		Registry:     resourceQuotaRegistry,
		ResyncPeriod: c.DefaultResyncPeriod,
		ControllerFactory: quotacontroller.NewAllResourceReplenishmentControllerFactory(
			ctx.ExternalKubeInformers,
			ctx.ImageInformers.Image().InternalVersion().ImageStreams(),
			ctx.ClientBuilder.DeprecatedOpenshiftClientOrDie(saName),
		),
		ReplenishmentResyncPeriod: controller.StaticResyncPeriodFunc(c.DefaultReplenishmentSyncPeriod),
		GroupKindsToReplenish:     groupKindsToReplenish,
	}
	clusterQuotaReconciliationController := clusterquotareconciliation.NewClusterQuotaReconcilationController(options)
	clusterQuotaMappingController.GetClusterQuotaMapper().AddListener(clusterQuotaReconciliationController)

	go clusterQuotaMappingController.Run(5, ctx.Stop)
	go clusterQuotaReconciliationController.Run(5, ctx.Stop)

	return true, nil
}

func calculateResyncPeriod(period time.Duration) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(period.Nanoseconds()) * factor)
	}
}
