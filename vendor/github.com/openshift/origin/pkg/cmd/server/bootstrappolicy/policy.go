package bootstrappolicy

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/apps"
	kauthenticationapi "k8s.io/kubernetes/pkg/apis/authentication"
	kauthorizationapi "k8s.io/kubernetes/pkg/apis/authorization"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/apis/settings"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac/bootstrappolicy"

	oapi "github.com/openshift/origin/pkg/api"
	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	buildapi "github.com/openshift/origin/pkg/build/apis/build"
	deployapi "github.com/openshift/origin/pkg/deploy/apis/apps"
	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	networkapi "github.com/openshift/origin/pkg/network/apis/network"
	oauthapi "github.com/openshift/origin/pkg/oauth/apis/oauth"
	projectapi "github.com/openshift/origin/pkg/project/apis/project"
	quotaapi "github.com/openshift/origin/pkg/quota/apis/quota"
	routeapi "github.com/openshift/origin/pkg/route/apis/route"
	securityapi "github.com/openshift/origin/pkg/security/apis/security"
	templateapi "github.com/openshift/origin/pkg/template/apis/template"
	userapi "github.com/openshift/origin/pkg/user/apis/user"
)

const (
	// roleSystemOnly is an annotation key that determines if a role is system only
	roleSystemOnly = "authorization.openshift.io/system-only"
	// roleIsSystemOnly is an annotation value that denotes roleSystemOnly, and thus excludes the role from the UI
	roleIsSystemOnly = "true"
)

var (
	readWrite = []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"}
	read      = []string{"get", "list", "watch"}

	kapiGroup            = kapi.GroupName
	appsGroup            = apps.GroupName
	autoscalingGroup     = autoscaling.GroupName
	apiExtensionsGroup   = "apiextensions.k8s.io"
	apiRegistrationGroup = "apiregistration.k8s.io"
	batchGroup           = batch.GroupName
	certificatesGroup    = certificates.GroupName
	extensionsGroup      = extensions.GroupName
	networkingGroup      = "networking.k8s.io"
	policyGroup          = policy.GroupName
	rbacGroup            = rbac.GroupName
	securityGroup        = securityapi.GroupName
	legacySecurityGroup  = securityapi.LegacyGroupName
	storageGroup         = storage.GroupName
	settingsGroup        = settings.GroupName

	authzGroup          = authorizationapi.GroupName
	kAuthzGroup         = kauthorizationapi.GroupName
	kAuthnGroup         = kauthenticationapi.GroupName
	legacyAuthzGroup    = authorizationapi.LegacyGroupName
	buildGroup          = buildapi.GroupName
	legacyBuildGroup    = buildapi.LegacyGroupName
	deployGroup         = deployapi.GroupName
	legacyDeployGroup   = deployapi.LegacyGroupName
	imageGroup          = imageapi.GroupName
	legacyImageGroup    = imageapi.LegacyGroupName
	projectGroup        = projectapi.GroupName
	legacyProjectGroup  = projectapi.LegacyGroupName
	quotaGroup          = quotaapi.GroupName
	legacyQuotaGroup    = quotaapi.LegacyGroupName
	routeGroup          = routeapi.GroupName
	legacyRouteGroup    = routeapi.LegacyGroupName
	templateGroup       = templateapi.GroupName
	legacyTemplateGroup = templateapi.LegacyGroupName
	userGroup           = userapi.GroupName
	legacyUserGroup     = userapi.LegacyGroupName
	oauthGroup          = oauthapi.GroupName
	legacyOauthGroup    = oauthapi.LegacyGroupName
	networkGroup        = networkapi.GroupName
	legacyNetworkGroup  = networkapi.LegacyGroupName
)

func GetBootstrapOpenshiftRoles(openshiftNamespace string) []rbac.Role {
	return []rbac.Role{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      OpenshiftSharedResourceViewRoleName,
				Namespace: openshiftNamespace,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(read...).
					Groups(templateGroup, legacyTemplateGroup).
					Resources("templates").
					RuleOrDie(),
				rbac.NewRule(read...).
					Groups(imageGroup, legacyImageGroup).
					Resources("imagestreams", "imagestreamtags", "imagestreamimages").
					RuleOrDie(),
				// so anyone can pull from openshift/* image streams
				rbac.NewRule("get").
					Groups(imageGroup, legacyImageGroup).
					Resources("imagestreams/layers").
					RuleOrDie(),
			},
		},
	}
}

func GetOpenshiftBootstrapClusterRoles() []rbac.ClusterRole {
	// four resource can be a single line
	// up to ten-ish resources per line otherwise

	roles := []rbac.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ClusterAdminRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A super-user that can perform any action in the cluster. When granted to a user within a project, they have full control over quota and membership and can perform every action on every resource in the project.",
				},
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(rbac.VerbAll).Groups(rbac.APIGroupAll).Resources(rbac.ResourceAll).RuleOrDie(),
				rbac.NewRule(rbac.VerbAll).URLs(rbac.NonResourceAll).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SudoerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("impersonate").Groups(userGroup, legacyUserGroup).Resources(authorizationapi.SystemUserResource).Names(SystemAdminUsername).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ClusterReaderRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(read...).Groups(kapiGroup).Resources("bindings", "componentstatuses", "configmaps", "endpoints", "events", "limitranges",
					"namespaces", "namespaces/status", "nodes", "nodes/status", "persistentvolumeclaims", "persistentvolumeclaims/status", "persistentvolumes",
					"persistentvolumes/status", "pods", "pods/binding", "pods/eviction", "pods/log", "pods/status", "podtemplates", "replicationcontrollers", "replicationcontrollers/scale",
					"replicationcontrollers/status", "resourcequotas", "resourcequotas/status", "securitycontextconstraints", "serviceaccounts", "services",
					"services/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(appsGroup).Resources("statefulsets", "statefulsets/status", "deployments", "deployments/scale", "deployments/status", "controllerrevisions").RuleOrDie(),

				rbac.NewRule(read...).Groups(apiExtensionsGroup).Resources("customresourcedefinitions", "customresourcedefinitions/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(apiRegistrationGroup).Resources("apiservices", "apiservices/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(autoscalingGroup).Resources("horizontalpodautoscalers", "horizontalpodautoscalers/status").RuleOrDie(),

				// TODO do we still need scheduledjobs?
				rbac.NewRule(read...).Groups(batchGroup).Resources("jobs", "jobs/status", "scheduledjobs", "scheduledjobs/status", "cronjobs", "cronjobs/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(extensionsGroup).Resources("daemonsets", "daemonsets/status", "deployments", "deployments/scale",
					"deployments/status", "horizontalpodautoscalers", "horizontalpodautoscalers/status", "ingresses", "ingresses/status", "jobs", "jobs/status",
					"networkpolicies", "podsecuritypolicies", "replicasets", "replicasets/scale", "replicasets/status", "replicationcontrollers",
					"replicationcontrollers/scale", "storageclasses", "thirdpartyresources").RuleOrDie(),

				rbac.NewRule(read...).Groups(networkingGroup).Resources("networkpolicies").RuleOrDie(),

				rbac.NewRule(read...).Groups(policyGroup).Resources("poddisruptionbudgets", "poddisruptionbudgets/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(rbacGroup).Resources("roles", "rolebindings", "clusterroles", "clusterrolebindings").RuleOrDie(),

				rbac.NewRule(read...).Groups(settingsGroup).Resources("podpresets").RuleOrDie(),

				rbac.NewRule(read...).Groups(storageGroup).Resources("storageclasses").RuleOrDie(),

				rbac.NewRule(read...).Groups(certificatesGroup).Resources("certificatesigningrequests", "certificatesigningrequests/approval", "certificatesigningrequests/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(authzGroup, legacyAuthzGroup).Resources("clusterpolicies", "clusterpolicybindings", "clusterroles", "clusterrolebindings",
					"policies", "policybindings", "roles", "rolebindings", "rolebindingrestrictions").RuleOrDie(),

				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "builds/details", "buildconfigs", "buildconfigs/webhooks", "builds/log").RuleOrDie(),

				rbac.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale", "deploymentconfigs/log",
					"deploymentconfigs/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("images", "imagesignatures", "imagestreams", "imagestreamtags", "imagestreamimages",
					"imagestreams/status").RuleOrDie(),
				// pull images
				rbac.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbac.NewRule(read...).Groups(oauthGroup, legacyOauthGroup).Resources("oauthclientauthorizations").RuleOrDie(),

				rbac.NewRule(read...).Groups(projectGroup, legacyProjectGroup).Resources("projectrequests", "projects").RuleOrDie(),

				rbac.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas", "clusterresourcequotas", "clusterresourcequotas/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes", "routes/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(networkGroup, legacyNetworkGroup).Resources("clusternetworks", "egressnetworkpolicies", "hostsubnets", "netnamespaces").RuleOrDie(),

				rbac.NewRule(read...).Groups(securityGroup, legacySecurityGroup).Resources("securitycontextconstraints").RuleOrDie(),

				rbac.NewRule(read...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),
				rbac.NewRule(read...).Groups(templateGroup, legacyTemplateGroup).Resources("brokertemplateinstances", "templateinstances/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(userGroup, legacyUserGroup).Resources("groups", "identities", "useridentitymappings", "users").RuleOrDie(),

				// permissions to check access.  These creates are non-mutating
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("localresourceaccessreviews", "localsubjectaccessreviews", "resourceaccessreviews",
					"selfsubjectrulesreviews", "subjectrulesreviews", "subjectaccessreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthzGroup).Resources("selfsubjectaccessreviews", "subjectaccessreviews", "localsubjectaccessreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthnGroup).Resources("tokenreviews").RuleOrDie(),
				// permissions to check PSP, these creates are non-mutating
				rbac.NewRule("create").Groups(securityGroup, legacySecurityGroup).Resources("podsecuritypolicysubjectreviews", "podsecuritypolicyselfsubjectreviews", "podsecuritypolicyreviews").RuleOrDie(),
				// Allow read access to node metrics
				rbac.NewRule("get").Groups(kapiGroup).Resources("nodes/"+authorizationapi.NodeMetricsSubresource, "nodes/"+authorizationapi.NodeSpecSubresource).RuleOrDie(),
				// Allow read access to stats
				// Node stats requests are submitted as POSTs.  These creates are non-mutating
				rbac.NewRule("get", "create").Groups(kapiGroup).Resources("nodes/" + authorizationapi.NodeStatsSubresource).RuleOrDie(),

				rbac.NewRule("get").URLs(rbac.NonResourceAll).RuleOrDie(),

				// backwards compatibility
				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ClusterDebuggerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get").URLs("/metrics", "/debug/pprof", "/debug/pprof/*").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategyDockerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(authorizationapi.DockerBuildResource, authorizationapi.OptimizedDockerBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategyCustomRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(authorizationapi.CustomBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategySourceRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(authorizationapi.SourceBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BuildStrategyJenkinsPipelineRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources(authorizationapi.JenkinsPipelineBuildResource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: StorageAdminRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("persistentvolumes").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(storageGroup).Resources("storageclasses").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("persistentvolumeclaims", "events").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: AdminRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A user that has edit rights within the project and can change the project's membership.",
				},
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("pods", "pods/attach", "pods/proxy", "pods/exec", "pods/portforward").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("replicationcontrollers", "replicationcontrollers/scale", "serviceaccounts",
					"services", "services/proxy", "endpoints", "persistentvolumeclaims", "configmaps", "secrets").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("limitranges", "resourcequotas", "bindings", "events",
					"namespaces", "pods/status", "resourcequotas/status", "namespaces/status", "replicationcontrollers/status", "pods/log").RuleOrDie(),
				rbac.NewRule("impersonate").Groups(kapiGroup).Resources("serviceaccounts").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(autoscalingGroup).Resources("horizontalpodautoscalers").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(batchGroup).Resources("jobs", "scheduledjobs", "cronjobs").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(extensionsGroup).Resources("horizontalpodautoscalers", "replicationcontrollers/scale",
					"replicasets", "replicasets/scale", "deployments", "deployments/scale", "deployments/rollback", "networkpolicies").RuleOrDie(),
				rbac.NewRule(read...).Groups(extensionsGroup).Resources("daemonsets").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(appsGroup).Resources("statefulsets", "deployments", "deployments/scale", "deployments/status").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(authzGroup, legacyAuthzGroup).Resources("roles", "rolebindings").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(rbacGroup).Resources("roles", "rolebindings").RuleOrDie(),
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("localresourceaccessreviews", "localsubjectaccessreviews", "subjectrulesreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(securityGroup, legacySecurityGroup).Resources("podsecuritypolicysubjectreviews", "podsecuritypolicyselfsubjectreviews", "podsecuritypolicyreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthzGroup).Resources("localsubjectaccessreviews").RuleOrDie(),

				rbac.NewRule(read...).Groups(authzGroup, legacyAuthzGroup).Resources("policies", "policybindings", "rolebindingrestrictions").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "buildconfigs", "buildconfigs/webhooks").RuleOrDie(),
				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/log").RuleOrDie(),
				rbac.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs/instantiate", "buildconfigs/instantiatebinary", "builds/clone").RuleOrDie(),
				rbac.NewRule("update").Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),
				// access to jenkins.  multiple values to ensure that covers relationships
				rbac.NewRule("admin", "edit", "view").Groups(buildapi.GroupName).Resources("jenkins").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale").RuleOrDie(),
				rbac.NewRule("create").Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigrollbacks", "deploymentconfigs/rollback", "deploymentconfigs/instantiate").RuleOrDie(),
				rbac.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs/log", "deploymentconfigs/status").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams", "imagestreammappings", "imagestreamtags", "imagestreamimages", "imagestreams/secrets").RuleOrDie(),
				rbac.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
				// push and pull images
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				rbac.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),

				rbac.NewRule("get", "patch", "update", "delete").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbac.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				// admins can create routes with custom hosts
				rbac.NewRule("create").Groups(routeGroup, legacyRouteGroup).Resources("routes/custom-host").RuleOrDie(),
				rbac.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),
				// an admin can run routers that write back conditions to the route
				rbac.NewRule("update").Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),

				// backwards compatibility
				rbac.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("resourceaccessreviews", "subjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: EditRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A user that can create and edit most objects in a project, but can not update the project's membership.",
				},
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("pods", "pods/attach", "pods/proxy", "pods/exec", "pods/portforward").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("replicationcontrollers", "replicationcontrollers/scale", "serviceaccounts",
					"services", "services/proxy", "endpoints", "persistentvolumeclaims", "configmaps", "secrets").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("limitranges", "resourcequotas", "bindings", "events",
					"namespaces", "pods/status", "resourcequotas/status", "namespaces/status", "replicationcontrollers/status", "pods/log").RuleOrDie(),
				rbac.NewRule("impersonate").Groups(kapiGroup).Resources("serviceaccounts").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(autoscalingGroup).Resources("horizontalpodautoscalers").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(batchGroup).Resources("jobs", "scheduledjobs", "cronjobs").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(extensionsGroup).Resources("horizontalpodautoscalers", "replicationcontrollers/scale",
					"replicasets", "replicasets/scale", "deployments", "deployments/scale", "deployments/rollback").RuleOrDie(),
				rbac.NewRule(read...).Groups(extensionsGroup).Resources("daemonsets").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(appsGroup).Resources("statefulsets", "deployments", "deployments/scale", "deployments/status").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "buildconfigs", "buildconfigs/webhooks").RuleOrDie(),
				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/log").RuleOrDie(),
				rbac.NewRule("create").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs/instantiate", "buildconfigs/instantiatebinary", "builds/clone").RuleOrDie(),
				rbac.NewRule("update").Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),
				// access to jenkins.  multiple values to ensure that covers relationships
				rbac.NewRule("edit", "view").Groups(buildapi.GroupName).Resources("jenkins").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale").RuleOrDie(),
				rbac.NewRule("create").Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigrollbacks", "deploymentconfigs/rollback", "deploymentconfigs/instantiate").RuleOrDie(),
				rbac.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs/log", "deploymentconfigs/status").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams", "imagestreammappings", "imagestreamtags", "imagestreamimages", "imagestreams/secrets").RuleOrDie(),
				rbac.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
				// push and pull images
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				rbac.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),

				rbac.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbac.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				// editors can create routes with custom hosts
				rbac.NewRule("create").Groups(routeGroup, legacyRouteGroup).Resources("routes/custom-host").RuleOrDie(),
				rbac.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),

				rbac.NewRule(readWrite...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),

				// backwards compatibility
				rbac.NewRule(readWrite...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ViewRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A user who can view but not edit any resources within the project. They can not view secrets or membership.",
				},
			},
			Rules: []rbac.PolicyRule{
				// TODO add "replicationcontrollers/scale" here
				rbac.NewRule(read...).Groups(kapiGroup).Resources("pods", "replicationcontrollers", "serviceaccounts",
					"services", "endpoints", "persistentvolumeclaims", "configmaps").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("limitranges", "resourcequotas", "bindings", "events",
					"namespaces", "pods/status", "resourcequotas/status", "namespaces/status", "replicationcontrollers/status", "pods/log").RuleOrDie(),

				rbac.NewRule(read...).Groups(autoscalingGroup).Resources("horizontalpodautoscalers").RuleOrDie(),

				rbac.NewRule(read...).Groups(batchGroup).Resources("jobs", "scheduledjobs", "cronjobs").RuleOrDie(),

				rbac.NewRule(read...).Groups(extensionsGroup).Resources("horizontalpodautoscalers", "replicasets", "replicasets/scale",
					"deployments", "deployments/scale").RuleOrDie(),
				rbac.NewRule(read...).Groups(extensionsGroup).Resources("daemonsets").RuleOrDie(),

				rbac.NewRule(read...).Groups(appsGroup).Resources("statefulsets", "deployments", "deployments/scale").RuleOrDie(),

				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds", "buildconfigs", "buildconfigs/webhooks").RuleOrDie(),
				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("builds/log").RuleOrDie(),
				// access to jenkins
				rbac.NewRule("view").Groups(buildapi.GroupName).Resources("jenkins").RuleOrDie(),

				rbac.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs", "deploymentconfigs/scale").RuleOrDie(),
				rbac.NewRule(read...).Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs/log", "deploymentconfigs/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams", "imagestreammappings", "imagestreamtags", "imagestreamimages").RuleOrDie(),
				rbac.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
				// TODO let them pull images?
				// pull images
				// rbac.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbac.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				rbac.NewRule(read...).Groups(quotaGroup, legacyQuotaGroup).Resources("appliedclusterresourcequotas").RuleOrDie(),

				rbac.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				rbac.NewRule(read...).Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),

				rbac.NewRule(read...).Groups(templateGroup, legacyTemplateGroup).Resources("templates", "templateconfigs", "processedtemplates", "templateinstances").RuleOrDie(),

				// backwards compatibility
				rbac.NewRule(read...).Groups(buildGroup, legacyBuildGroup).Resources("buildlogs").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("resourcequotausages").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: BasicUserRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A user that can get basic information about projects.",
				},
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get").Groups(userGroup, legacyUserGroup).Resources("users").Names("~").RuleOrDie(),
				rbac.NewRule("list").Groups(projectGroup, legacyProjectGroup).Resources("projectrequests").RuleOrDie(),
				rbac.NewRule("get", "list").Groups(authzGroup, legacyAuthzGroup).Resources("clusterroles").RuleOrDie(),
				rbac.NewRule(read...).Groups(rbacGroup).Resources("clusterroles").RuleOrDie(),
				rbac.NewRule("get", "list").Groups(storageGroup).Resources("storageclasses").RuleOrDie(),
				rbac.NewRule("list", "watch").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("selfsubjectrulesreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthzGroup).Resources("selfsubjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SelfAccessReviewerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("selfsubjectrulesreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthzGroup).Resources("selfsubjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SelfProvisionerRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A user that can request projects.",
				},
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("create").Groups(projectGroup, legacyProjectGroup).Resources("projectrequests").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: StatusCheckerRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "A user that can get basic cluster status information.",
				},
			},
			Rules: []rbac.PolicyRule{
				// Health
				rbac.NewRule("get").URLs("/healthz", "/healthz/*").RuleOrDie(),
				authorizationapi.RbacDiscoveryRule,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImageAuditorRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get", "list", "watch", "patch", "update").Groups(imageGroup, legacyImageGroup).Resources("images").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImagePullerRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "Grants the right to pull images from within a project.",
				},
			},
			Rules: []rbac.PolicyRule{
				// pull images
				rbac.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
			},
		},
		{
			// This role looks like a duplicate of ImageBuilderRole, but the ImageBuilder role is specifically for our builder service accounts
			// if we found another permission needed by them, we'd add it there so the intent is different if you used the ImageBuilderRole
			// you could end up accidentally granting more permissions than you intended.  This is intended to only grant enough powers to
			// push an image to our registry
			ObjectMeta: metav1.ObjectMeta{
				Name: ImagePusherRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "Grants the right to push and pull images from within a project.",
				},
			},
			Rules: []rbac.PolicyRule{
				// push and pull images
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImageBuilderRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "Grants the right to build, push and pull images from within a project.  Used primarily with service accounts for builds.",
				},
			},
			Rules: []rbac.PolicyRule{
				// push and pull images
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				// allow auto-provisioning when pushing an image that doesn't have an imagestream yet
				rbac.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreams").RuleOrDie(),
				rbac.NewRule("update").Groups(buildGroup, legacyBuildGroup).Resources("builds/details").RuleOrDie(),
				rbac.NewRule("get").Groups(buildGroup, legacyBuildGroup).Resources("builds").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImagePrunerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get", "list").Groups(kapiGroup).Resources("pods", "replicationcontrollers").RuleOrDie(),
				rbac.NewRule("list").Groups(kapiGroup).Resources("limitranges").RuleOrDie(),
				rbac.NewRule("get", "list").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs", "builds").RuleOrDie(),
				rbac.NewRule("get", "list").Groups(deployGroup, legacyDeployGroup).Resources("deploymentconfigs").RuleOrDie(),

				rbac.NewRule("delete").Groups(imageGroup, legacyImageGroup).Resources("images").RuleOrDie(),
				rbac.NewRule("get", "list").Groups(imageGroup, legacyImageGroup).Resources("images", "imagestreams").RuleOrDie(),
				rbac.NewRule("update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/status").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: ImageSignerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("images", "imagestreams/layers").RuleOrDie(),
				rbac.NewRule("create", "delete").Groups(imageGroup, legacyImageGroup).Resources("imagesignatures").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: DeployerRoleName,
				Annotations: map[string]string{
					oapi.OpenShiftDescription: "Grants the right to deploy within a project.  Used primarily with service accounts for automated deployments.",
				},
			},
			Rules: []rbac.PolicyRule{
				// "delete" is required here for compatibility with older deployer images
				// (see https://github.com/openshift/origin/pull/14322#issuecomment-303968976)
				// TODO: remove "delete" rule few releases after 3.6
				rbac.NewRule("delete").Groups(kapiGroup).Resources("replicationcontrollers").RuleOrDie(),
				rbac.NewRule("get", "list", "watch", "update").Groups(kapiGroup).Resources("replicationcontrollers").RuleOrDie(),
				rbac.NewRule("get", "list", "watch", "create").Groups(kapiGroup).Resources("pods").RuleOrDie(),
				rbac.NewRule("get").Groups(kapiGroup).Resources("pods/log").RuleOrDie(),
				rbac.NewRule("create", "list").Groups(kapiGroup).Resources("events").RuleOrDie(),

				rbac.NewRule("update").Groups(imageGroup, legacyImageGroup).Resources("imagestreamtags").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: MasterRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(rbac.VerbAll).Groups(rbac.APIGroupAll).Resources(rbac.ResourceAll).RuleOrDie(),
				rbac.NewRule(rbac.VerbAll).URLs(rbac.NonResourceAll).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: OAuthTokenDeleterRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("delete").Groups(oauthGroup, legacyOauthGroup).Resources("oauthaccesstokens", "oauthauthorizetokens").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: RouterRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("list", "watch").Groups(kapiGroup).Resources("endpoints").RuleOrDie(),
				rbac.NewRule("list", "watch").Groups(kapiGroup).Resources("services").RuleOrDie(),

				rbac.NewRule("list", "watch").Groups(routeGroup, legacyRouteGroup).Resources("routes").RuleOrDie(),
				rbac.NewRule("update").Groups(routeGroup, legacyRouteGroup).Resources("routes/status").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: RegistryRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("list").Groups(kapiGroup).Resources("limitranges", "resourcequotas").RuleOrDie(),

				rbac.NewRule("get", "delete").Groups(imageGroup, legacyImageGroup).Resources("images", "imagestreamtags").RuleOrDie(),
				rbac.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreams/secrets").RuleOrDie(),
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("images", "imagestreams").RuleOrDie(),
				rbac.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreammappings").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeProxierRoleName,
			},
			Rules: []rbac.PolicyRule{
				// Used to build serviceLister
				rbac.NewRule("list", "watch").Groups(kapiGroup).Resources("services", "endpoints").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeAdminRoleName,
			},
			Rules: []rbac.PolicyRule{
				// Allow read-only access to the API objects
				rbac.NewRule(read...).Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				// Allow all API calls to the nodes
				rbac.NewRule("proxy").Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				rbac.NewRule("*").Groups(kapiGroup).Resources("nodes/proxy", "nodes/"+authorizationapi.NodeMetricsSubresource, "nodes/"+authorizationapi.NodeSpecSubresource, "nodes/"+authorizationapi.NodeStatsSubresource, "nodes/"+authorizationapi.NodeLogSubresource).RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeReaderRoleName,
			},
			Rules: []rbac.PolicyRule{
				// Allow read-only access to the API objects
				rbac.NewRule(read...).Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				// Allow read access to node metrics
				rbac.NewRule("get").Groups(kapiGroup).Resources("nodes/"+authorizationapi.NodeMetricsSubresource, "nodes/"+authorizationapi.NodeSpecSubresource).RuleOrDie(),
				// Allow read access to stats
				// Node stats requests are submitted as POSTs.  These creates are non-mutating
				rbac.NewRule("get", "create").Groups(kapiGroup).Resources("nodes/" + authorizationapi.NodeStatsSubresource).RuleOrDie(),
				// TODO: expose other things like /healthz on the node once we figure out non-resource URL policy across systems
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeRoleName,
			},
			Rules: []rbac.PolicyRule{
				// Needed to check API access.  These creates are non-mutating
				rbac.NewRule("create").Groups(kAuthnGroup).Resources("tokenreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("subjectaccessreviews", "localsubjectaccessreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthzGroup).Resources("subjectaccessreviews", "localsubjectaccessreviews").RuleOrDie(),
				// Needed to build serviceLister, to populate env vars for services
				rbac.NewRule(read...).Groups(kapiGroup).Resources("services").RuleOrDie(),
				// Nodes can register themselves
				// TODO: restrict to creating a node with the same name they announce
				rbac.NewRule("create", "get", "list", "watch").Groups(kapiGroup).Resources("nodes").RuleOrDie(),
				// TODO: restrict to the bound node once supported
				rbac.NewRule("update", "patch").Groups(kapiGroup).Resources("nodes/status").RuleOrDie(),

				// TODO: restrict to the bound node as creator once supported
				rbac.NewRule("create", "update", "patch").Groups(kapiGroup).Resources("events").RuleOrDie(),

				// TODO: restrict to pods scheduled on the bound node once supported
				rbac.NewRule(read...).Groups(kapiGroup).Resources("pods").RuleOrDie(),

				// TODO: remove once mirror pods are removed
				// TODO: restrict deletion to mirror pods created by the bound node once supported
				// Needed for the node to create/delete mirror pods
				rbac.NewRule("get", "create", "delete").Groups(kapiGroup).Resources("pods").RuleOrDie(),
				// TODO: restrict to pods scheduled on the bound node once supported
				rbac.NewRule("update").Groups(kapiGroup).Resources("pods/status").RuleOrDie(),

				// TODO: restrict to secrets and configmaps used by pods scheduled on bound node once supported
				// Needed for imagepullsecrets, rbd/ceph and secret volumes, and secrets in envs
				// Needed for configmap volume and envs
				rbac.NewRule("get").Groups(kapiGroup).Resources("secrets", "configmaps").RuleOrDie(),
				// TODO: restrict to claims/volumes used by pods scheduled on bound node once supported
				// Needed for persistent volumes
				rbac.NewRule("get").Groups(kapiGroup).Resources("persistentvolumeclaims", "persistentvolumes").RuleOrDie(),
				// TODO: restrict to namespaces of pods scheduled on bound node once supported
				// TODO: change glusterfs to use DNS lookup so this isn't needed?
				// Needed for glusterfs volumes
				rbac.NewRule("get").Groups(kapiGroup).Resources("endpoints").RuleOrDie(),
				// Nodes are allowed to request CSRs (specifically, request serving certs)
				rbac.NewRule("get", "create").Groups(certificates.GroupName).Resources("certificatesigningrequests").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SDNReaderRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(read...).Groups(networkGroup, legacyNetworkGroup).Resources("egressnetworkpolicies", "hostsubnets", "netnamespaces").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("nodes", "namespaces").RuleOrDie(),
				rbac.NewRule(read...).Groups(extensionsGroup).Resources("networkpolicies").RuleOrDie(),
				rbac.NewRule("get").Groups(networkGroup, legacyNetworkGroup).Resources("clusternetworks").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: SDNManagerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get", "list", "watch", "create", "delete").Groups(networkGroup, legacyNetworkGroup).Resources("hostsubnets", "netnamespaces").RuleOrDie(),
				rbac.NewRule("get", "create").Groups(networkGroup, legacyNetworkGroup).Resources("clusternetworks").RuleOrDie(),
				rbac.NewRule(read...).Groups(kapiGroup).Resources("nodes").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: WebHooksRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get", "create").Groups(buildGroup, legacyBuildGroup).Resources("buildconfigs/webhooks").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: DiscoveryRoleName,
			},
			Rules: []rbac.PolicyRule{
				authorizationapi.RbacDiscoveryRule,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: PersistentVolumeProvisionerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get", "list", "watch", "create", "delete").Groups(kapiGroup).Resources("persistentvolumes").RuleOrDie(),
				// update is needed in addition to read access for setting lock annotations on PVCs
				rbac.NewRule("get", "list", "watch", "update").Groups(kapiGroup).Resources("persistentvolumeclaims").RuleOrDie(),
				rbac.NewRule(read...).Groups(storageGroup).Resources("storageclasses").RuleOrDie(),
				// Needed for watching provisioning success and failure events
				rbac.NewRule("create", "update", "patch", "list", "watch").Groups(kapiGroup).Resources("events").RuleOrDie(),
			},
		},

		{
			ObjectMeta: metav1.ObjectMeta{
				Name: RegistryAdminRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("serviceaccounts", "secrets").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreammappings", "imagestreams", "imagestreams/secrets", "imagestreamtags").RuleOrDie(),
				rbac.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(authzGroup, legacyAuthzGroup).Resources("rolebindings", "roles").RuleOrDie(),
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("localresourceaccessreviews", "localsubjectaccessreviews", "subjectrulesreviews").RuleOrDie(),
				rbac.NewRule("create").Groups(kAuthzGroup).Resources("localsubjectaccessreviews").RuleOrDie(),
				rbac.NewRule(read...).Groups(authzGroup, legacyAuthzGroup).Resources("policies", "policybindings").RuleOrDie(),

				rbac.NewRule("get").Groups(kapiGroup).Resources("namespaces").RuleOrDie(),
				rbac.NewRule("get", "delete").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),

				// backwards compatibility
				rbac.NewRule("create").Groups(authzGroup, legacyAuthzGroup).Resources("resourceaccessreviews", "subjectaccessreviews").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: RegistryEditorRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(readWrite...).Groups(kapiGroup).Resources("serviceaccounts", "secrets").RuleOrDie(),
				rbac.NewRule(readWrite...).Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreammappings", "imagestreams", "imagestreams/secrets", "imagestreamtags").RuleOrDie(),
				rbac.NewRule("create").Groups(imageGroup, legacyImageGroup).Resources("imagestreamimports").RuleOrDie(),
				rbac.NewRule("get", "update").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbac.NewRule("get").Groups(kapiGroup).Resources("namespaces").RuleOrDie(),
				rbac.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: RegistryViewerRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule(read...).Groups(imageGroup, legacyImageGroup).Resources("imagestreamimages", "imagestreammappings", "imagestreams", "imagestreamtags").RuleOrDie(),
				rbac.NewRule("get").Groups(imageGroup, legacyImageGroup).Resources("imagestreams/layers").RuleOrDie(),

				rbac.NewRule("get").Groups(kapiGroup).Resources("namespaces").RuleOrDie(),
				rbac.NewRule("get").Groups(projectGroup, legacyProjectGroup).Resources("projects").RuleOrDie(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: TemplateServiceBrokerClientRoleName,
			},
			Rules: []rbac.PolicyRule{
				rbac.NewRule("get", "put", "update", "delete").URLs(templateapi.ServiceBrokerRoot + "/*").RuleOrDie(),
			},
		},
	}

	return roles
}

func GetBootstrapClusterRoles() []rbac.ClusterRole {
	openshiftClusterRoles := GetOpenshiftBootstrapClusterRoles()
	// dead cluster roles need to be checked for conflicts (in case something new comes up)
	// so add them to this list.
	openshiftClusterRoles = append(openshiftClusterRoles, GetDeadClusterRoles()...)
	kubeClusterRoles := bootstrappolicy.ClusterRoles()
	kubeSAClusterRoles := bootstrappolicy.ControllerRoles()
	openshiftControllerRoles := ControllerRoles()

	// Eventually openshift controllers and kube controllers have different prefixes
	// so we will only need to check conflicts on the "normal" cluster roles
	// for now, deconflict with all names
	openshiftClusterRoleNames := sets.NewString()
	kubeClusterRoleNames := sets.NewString()
	for _, clusterRole := range openshiftClusterRoles {
		openshiftClusterRoleNames.Insert(clusterRole.Name)
	}
	for _, clusterRole := range kubeClusterRoles {
		kubeClusterRoleNames.Insert(clusterRole.Name)
	}

	conflictingNames := kubeClusterRoleNames.Intersection(openshiftClusterRoleNames)
	extraRBACConflicts := conflictingNames.Difference(clusterRoleConflicts)
	extraWhitelistEntries := clusterRoleConflicts.Difference(conflictingNames)
	switch {
	case len(extraRBACConflicts) > 0 && len(extraWhitelistEntries) > 0:
		panic(fmt.Sprintf("kube ClusterRoles conflict with openshift ClusterRoles: %v and ClusterRole whitelist contains a extraneous entries: %v ", extraRBACConflicts.List(), extraWhitelistEntries.List()))
	case len(extraRBACConflicts) > 0:
		panic(fmt.Sprintf("kube ClusterRoles conflict with openshift ClusterRoles: %v", extraRBACConflicts.List()))
	case len(extraWhitelistEntries) > 0:
		panic(fmt.Sprintf("ClusterRole whitelist contains a extraneous entries: %v", extraWhitelistEntries.List()))
	}

	finalClusterRoles := []rbac.ClusterRole{}
	finalClusterRoles = append(finalClusterRoles, openshiftClusterRoles...)
	finalClusterRoles = append(finalClusterRoles, openshiftControllerRoles...)
	finalClusterRoles = append(finalClusterRoles, kubeSAClusterRoles...)
	for i := range kubeClusterRoles {
		if !clusterRoleConflicts.Has(kubeClusterRoles[i].Name) {
			finalClusterRoles = append(finalClusterRoles, kubeClusterRoles[i])
		}
	}

	// conditionally add the web console annotations
	for i := range finalClusterRoles {
		role := &finalClusterRoles[i]
		// adding annotation to any role not explicitly in the whitelist below
		if !rolesToShow.Has(role.Name) {
			if role.Annotations == nil {
				role.Annotations = map[string]string{}
			}
			role.Annotations[roleSystemOnly] = roleIsSystemOnly
		}
	}

	return finalClusterRoles
}

func GetBootstrapOpenshiftRoleBindings(openshiftNamespace string) []rbac.RoleBinding {
	return []rbac.RoleBinding{
		newOriginRoleBinding(OpenshiftSharedResourceViewRoleBindingName, OpenshiftSharedResourceViewRoleName, openshiftNamespace).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
	}
}

func newOriginRoleBinding(bindingName, roleName, namespace string) *rbac.RoleBindingBuilder {
	builder := rbac.NewRoleBinding(roleName, namespace)
	builder.RoleBinding.Name = bindingName
	return builder
}

func newOriginClusterBinding(bindingName, roleName string) *rbac.ClusterRoleBindingBuilder {
	builder := rbac.NewClusterBinding(roleName)
	builder.ClusterRoleBinding.Name = bindingName
	return builder
}

func GetOpenshiftBootstrapClusterRoleBindings() []rbac.ClusterRoleBinding {
	return []rbac.ClusterRoleBinding{
		newOriginClusterBinding(MasterRoleBindingName, MasterRoleName).
			Groups(MastersGroup).
			BindingOrDie(),
		newOriginClusterBinding(NodeAdminRoleBindingName, NodeAdminRoleName).
			Users(LegacyMasterKubeletAdminClientUsername).
			Groups(NodeAdminsGroup).
			BindingOrDie(),
		newOriginClusterBinding(ClusterAdminRoleBindingName, ClusterAdminRoleName).
			Groups(ClusterAdminGroup).
			// add system:admin to this binding so that members of the
			// sudoer group can use --as=system:admin to run a command
			// as a cluster-admin
			Users(SystemAdminUsername).
			BindingOrDie(),
		newOriginClusterBinding(ClusterReaderRoleBindingName, ClusterReaderRoleName).
			Groups(ClusterReaderGroup).
			BindingOrDie(),
		newOriginClusterBinding(BasicUserRoleBindingName, BasicUserRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(SelfAccessReviewerRoleBindingName, SelfAccessReviewerRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(SelfProvisionerRoleBindingName, SelfProvisionerRoleName).
			Groups(AuthenticatedOAuthGroup).
			BindingOrDie(),
		newOriginClusterBinding(OAuthTokenDeleterRoleBindingName, OAuthTokenDeleterRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(StatusCheckerRoleBindingName, StatusCheckerRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(NodeRoleBindingName, NodeRoleName).
			Groups(NodesGroup).
			BindingOrDie(),
		newOriginClusterBinding(NodeProxierRoleBindingName, NodeProxierRoleName).
			// Allow node identities to run node proxies
			Groups(NodesGroup).
			BindingOrDie(),
		newOriginClusterBinding(SDNReaderRoleBindingName, SDNReaderRoleName).
			// Allow node identities to run SDN plugins
			Groups(NodesGroup).
			BindingOrDie(),
		newOriginClusterBinding(WebHooksRoleBindingName, WebHooksRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(DiscoveryRoleBindingName, DiscoveryRoleName).
			Groups(AuthenticatedGroup, UnauthenticatedGroup).
			BindingOrDie(),
		// Allow all build strategies by default.
		// Cluster admins can remove these role bindings, and the reconcile-cluster-role-bindings command
		// run during an upgrade won't re-add the "system:authenticated" group
		newOriginClusterBinding(BuildStrategyDockerRoleBindingName, BuildStrategyDockerRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(BuildStrategySourceRoleBindingName, BuildStrategySourceRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		newOriginClusterBinding(BuildStrategyJenkinsPipelineRoleBindingName, BuildStrategyJenkinsPipelineRoleName).
			Groups(AuthenticatedGroup).
			BindingOrDie(),
		// Allow node-bootstrapper SA to bootstrap nodes by default.
		rbac.NewClusterBinding(NodeBootstrapRoleName).
			SAs(DefaultOpenShiftInfraNamespace, InfraNodeBootstrapServiceAccountName).
			BindingOrDie(),
	}
}

func GetBootstrapClusterRoleBindings() []rbac.ClusterRoleBinding {
	openshiftClusterRoleBindings := GetOpenshiftBootstrapClusterRoleBindings()
	kubeClusterRoleBindings := bootstrappolicy.ClusterRoleBindings()
	kubeControllerClusterRoleBindings := bootstrappolicy.ControllerRoleBindings()
	openshiftControllerClusterRoleBindings := ControllerRoleBindings()

	// openshift controllers and kube controllers have different prefixes
	// so we only need to check conflicts on the "normal" cluster rolebindings
	openshiftClusterRoleBindingNames := sets.NewString()
	kubeClusterRoleBindingNames := sets.NewString()
	for _, clusterRoleBinding := range openshiftClusterRoleBindings {
		openshiftClusterRoleBindingNames.Insert(clusterRoleBinding.Name)
	}
	for _, clusterRoleBinding := range kubeClusterRoleBindings {
		kubeClusterRoleBindingNames.Insert(clusterRoleBinding.Name)
	}

	conflictingNames := kubeClusterRoleBindingNames.Intersection(openshiftClusterRoleBindingNames)
	extraRBACConflicts := conflictingNames.Difference(clusterRoleBindingConflicts)
	extraWhitelistEntries := clusterRoleBindingConflicts.Difference(conflictingNames)
	switch {
	case len(extraRBACConflicts) > 0 && len(extraWhitelistEntries) > 0:
		panic(fmt.Sprintf("kube ClusterRoleBindings conflict with openshift ClusterRoleBindings: %v and ClusterRoleBinding whitelist contains a extraneous entries: %v ", extraRBACConflicts.List(), extraWhitelistEntries.List()))
	case len(extraRBACConflicts) > 0:
		panic(fmt.Sprintf("kube ClusterRoleBindings conflict with openshift ClusterRoleBindings: %v", extraRBACConflicts.List()))
	case len(extraWhitelistEntries) > 0:
		panic(fmt.Sprintf("ClusterRoleBinding whitelist contains a extraneous entries: %v", extraWhitelistEntries.List()))
	}

	finalClusterRoleBindings := []rbac.ClusterRoleBinding{}
	finalClusterRoleBindings = append(finalClusterRoleBindings, openshiftClusterRoleBindings...)
	finalClusterRoleBindings = append(finalClusterRoleBindings, kubeControllerClusterRoleBindings...)
	finalClusterRoleBindings = append(finalClusterRoleBindings, openshiftControllerClusterRoleBindings...)
	for i := range kubeClusterRoleBindings {
		if !clusterRoleBindingConflicts.Has(kubeClusterRoleBindings[i].Name) {
			finalClusterRoleBindings = append(finalClusterRoleBindings, kubeClusterRoleBindings[i])
		}
	}

	return finalClusterRoleBindings
}

// clusterRoleConflicts lists the roles which are known to conflict with upstream and which we have manually
// deconflicted with our own.
var clusterRoleConflicts = sets.NewString(
	// these require special treatment to handle origin resources
	"admin",
	"edit",
	"view",

	// TODO this should probably be re-swizzled to be the delta on top of the kube role
	"system:discovery",

	// TODO these should be reconsidered
	"cluster-admin",
	"system:node",
	"system:node-proxier",
	"system:persistent-volume-provisioner",
)

// clusterRoleBindingConflicts lists the roles which are known to conflict with upstream and which we have manually
// deconflicted with our own.
var clusterRoleBindingConflicts = sets.NewString()

// The current list of roles considered useful for normal users (non-admin)
var rolesToShow = sets.NewString(
	"admin",
	"basic-user",
	"edit",
	"system:deployer",
	"system:image-builder",
	"system:image-puller",
	"system:image-pusher",
	"view",
)
