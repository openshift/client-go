package diagnostics

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	"github.com/openshift/origin/pkg/client"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	clustdiags "github.com/openshift/origin/pkg/diagnostics/cluster"
	agldiags "github.com/openshift/origin/pkg/diagnostics/cluster/aggregated_logging"
	"github.com/openshift/origin/pkg/diagnostics/types"
)

var (
	// availableClusterDiagnostics contains the names of cluster diagnostics that can be executed
	// during a single run of diagnostics. Add more diagnostics to the list as they are defined.
	availableClusterDiagnostics = sets.NewString(
		agldiags.AggregatedLoggingName,
		clustdiags.ClusterRegistryName,
		clustdiags.ClusterRouterName,
		clustdiags.ClusterRolesName,
		clustdiags.ClusterRoleBindingsName,
		clustdiags.MasterNodeName,
		clustdiags.MetricsApiProxyName,
		clustdiags.NodeDefinitionsName,
		clustdiags.RouteCertificateValidationName,
		clustdiags.ServiceExternalIPsName,
	)
)

// buildClusterDiagnostics builds cluster Diagnostic objects if a cluster-admin client can be extracted from the rawConfig passed in.
// Returns the Diagnostics built, "ok" bool for whether to proceed or abort, and an error if any was encountered during the building of diagnostics.) {
func (o DiagnosticsOptions) buildClusterDiagnostics(rawConfig *clientcmdapi.Config) ([]types.Diagnostic, bool, error) {
	requestedDiagnostics := availableClusterDiagnostics.Intersection(sets.NewString(o.RequestedDiagnostics...)).List()
	if len(requestedDiagnostics) == 0 { // no diagnostics to run here
		return nil, true, nil // don't waste time on discovery
	}

	var (
		clusterClient  *client.Client
		kclusterClient kclientset.Interface
	)

	config, clusterClient, kclusterClient, found, serverUrl, err := o.findClusterClients(rawConfig)
	if !found {
		o.Logger.Notice("CED1002", "Could not configure a client with cluster-admin permissions for the current server, so cluster diagnostics will be skipped")
		return nil, true, err
	}
	if err != nil {
		return nil, false, err
	}

	diagnostics := []types.Diagnostic{}
	for _, diagnosticName := range requestedDiagnostics {
		var d types.Diagnostic
		switch diagnosticName {
		case agldiags.AggregatedLoggingName:
			d = agldiags.NewAggregatedLogging(o.MasterConfigLocation, kclusterClient, clusterClient)
		case clustdiags.NodeDefinitionsName:
			d = &clustdiags.NodeDefinitions{KubeClient: kclusterClient, OsClient: clusterClient}
		case clustdiags.MasterNodeName:
			d = &clustdiags.MasterNode{KubeClient: kclusterClient, OsClient: clusterClient, ServerUrl: serverUrl, MasterConfigFile: o.MasterConfigLocation}
		case clustdiags.ClusterRegistryName:
			d = &clustdiags.ClusterRegistry{KubeClient: kclusterClient, OsClient: clusterClient, PreventModification: o.PreventModification}
		case clustdiags.ClusterRouterName:
			d = &clustdiags.ClusterRouter{KubeClient: kclusterClient, OsClient: clusterClient}
		case clustdiags.ClusterRolesName:
			d = &clustdiags.ClusterRoles{ClusterRolesClient: clusterClient, SARClient: clusterClient}
		case clustdiags.ClusterRoleBindingsName:
			d = &clustdiags.ClusterRoleBindings{ClusterRoleBindingsClient: clusterClient, SARClient: clusterClient}
		case clustdiags.MetricsApiProxyName:
			d = &clustdiags.MetricsApiProxy{KubeClient: kclusterClient}
		case clustdiags.ServiceExternalIPsName:
			d = &clustdiags.ServiceExternalIPs{MasterConfigFile: o.MasterConfigLocation, KclusterClient: kclusterClient}
		case clustdiags.RouteCertificateValidationName:
			d = &clustdiags.RouteCertificateValidation{OsClient: clusterClient, RESTConfig: config}
		default:
			return nil, false, fmt.Errorf("unknown diagnostic: %v", diagnosticName)
		}
		diagnostics = append(diagnostics, d)
	}
	return diagnostics, true, nil
}

// attempts to find which context in the config might be a cluster-admin for the server in the current context.
func (o DiagnosticsOptions) findClusterClients(rawConfig *clientcmdapi.Config) (*rest.Config, *client.Client, kclientset.Interface, bool, string, error) {
	if o.ClientClusterContext != "" { // user has specified cluster context to use
		if context, exists := rawConfig.Contexts[o.ClientClusterContext]; exists {
			configErr := fmt.Errorf("Specified '%s' as cluster-admin context, but it was not found in your client configuration.", o.ClientClusterContext)
			o.Logger.Error("CED1003", configErr.Error())
			return nil, nil, nil, false, "", configErr
		} else if config, os, kube, found, serverUrl, err := o.makeClusterClients(rawConfig, o.ClientClusterContext, context); found {
			return config, os, kube, true, serverUrl, err
		} else {
			return nil, nil, nil, false, "", err
		}
	}
	currentContext, exists := rawConfig.Contexts[rawConfig.CurrentContext]
	if !exists { // config specified cluster admin context that doesn't exist; complain and quit
		configErr := fmt.Errorf("Current context '%s' not found in client configuration; will not attempt cluster diagnostics.", rawConfig.CurrentContext)
		o.Logger.Error("CED1004", configErr.Error())
		return nil, nil, nil, false, "", configErr
	}
	// check if current context is already cluster admin
	if config, os, kube, found, serverUrl, err := o.makeClusterClients(rawConfig, rawConfig.CurrentContext, currentContext); found {
		return config, os, kube, true, serverUrl, err
	}
	// otherwise, for convenience, search for a context with the same server but with the system:admin user
	for name, context := range rawConfig.Contexts {
		if context.Cluster == currentContext.Cluster && name != rawConfig.CurrentContext && strings.HasPrefix(context.AuthInfo, "system:admin/") {
			if config, os, kube, found, serverUrl, err := o.makeClusterClients(rawConfig, name, context); found {
				return config, os, kube, true, serverUrl, err
			} else {
				return nil, nil, nil, false, "", err // don't try more than one such context, they'll probably fail the same
			}
		}
	}
	return nil, nil, nil, false, "", nil
}

// makes the client from the specified context and determines whether it is a cluster-admin.
func (o DiagnosticsOptions) makeClusterClients(rawConfig *clientcmdapi.Config, contextName string, context *clientcmdapi.Context) (*rest.Config, *client.Client, kclientset.Interface, bool, string, error) {
	overrides := &clientcmd.ConfigOverrides{Context: *context}
	clientConfig := clientcmd.NewDefaultClientConfig(*rawConfig, overrides)
	serverUrl := rawConfig.Clusters[context.Cluster].Server
	factory := osclientcmd.NewFactory(clientConfig)
	config, err := factory.ClientConfig()
	if err != nil {
		o.Logger.Debug("CED1006", fmt.Sprintf("Error creating client for context '%s':\n%v", contextName, err))
		return nil, nil, nil, false, "", nil
	}
	o.Logger.Debug("CED1005", fmt.Sprintf("Checking if context is cluster-admin: '%s'", contextName))
	if osClient, kubeClient, err := factory.Clients(); err != nil {
		o.Logger.Debug("CED1006", fmt.Sprintf("Error creating client for context '%s':\n%v", contextName, err))
		return nil, nil, nil, false, "", nil
	} else {
		subjectAccessReview := authorizationapi.SubjectAccessReview{Action: authorizationapi.Action{
			// if you can do everything, you're the cluster admin.
			Verb:     "*",
			Group:    "*",
			Resource: "*",
		}}
		if resp, err := osClient.SubjectAccessReviews().Create(&subjectAccessReview); err != nil {
			if regexp.MustCompile(`User "[\w:]+" cannot create \w+ at the cluster scope`).MatchString(err.Error()) {
				o.Logger.Debug("CED1007", fmt.Sprintf("Context '%s' does not have cluster-admin access:\n%v", contextName, err))
				return nil, nil, nil, false, "", nil
			} else {
				o.Logger.Error("CED1008", fmt.Sprintf("Unknown error testing cluster-admin access for context '%s':\n%v", contextName, err))
				return nil, nil, nil, false, "", err
			}
		} else if resp.Allowed {
			o.Logger.Info("CED1009", fmt.Sprintf("Using context for cluster-admin access: '%s'", contextName))
			return config, osClient, kubeClient, true, serverUrl, nil
		}
	}
	o.Logger.Debug("CED1010", fmt.Sprintf("Context does not have cluster-admin access: '%s'", contextName))
	return nil, nil, nil, false, "", nil
}
