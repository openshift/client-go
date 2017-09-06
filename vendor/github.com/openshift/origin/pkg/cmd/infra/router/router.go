package router

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	kcoreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"

	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/variable"
	projectclient "github.com/openshift/origin/pkg/project/generated/internalclientset/typed/project/internalversion"
	routeapi "github.com/openshift/origin/pkg/route/apis/route"
	routeclient "github.com/openshift/origin/pkg/route/generated/internalclientset/typed/route/internalversion"
	"github.com/openshift/origin/pkg/router/controller"
	controllerfactory "github.com/openshift/origin/pkg/router/controller/factory"
)

// RouterSelection controls what routes and resources on the server are considered
// part of this router.
type RouterSelection struct {
	ResyncInterval time.Duration

	HostnameTemplate string
	OverrideHostname bool

	LabelSelector string
	Labels        labels.Selector
	FieldSelector string
	Fields        fields.Selector

	Namespace              string
	NamespaceLabelSelector string
	NamespaceLabels        labels.Selector

	ProjectLabelSelector string
	ProjectLabels        labels.Selector

	IncludeUDP bool

	DeniedDomains      []string
	BlacklistedDomains sets.String

	AllowedDomains     []string
	WhitelistedDomains sets.String

	AllowWildcardRoutes bool

	DisableNamespaceOwnershipCheck bool

	EnableIngress bool

	ListenAddr string
}

// Bind sets the appropriate labels
func (o *RouterSelection) Bind(flag *pflag.FlagSet) {
	flag.DurationVar(&o.ResyncInterval, "resync-interval", 10*time.Minute, "The interval at which the route list should be fully refreshed")
	flag.StringVar(&o.HostnameTemplate, "hostname-template", cmdutil.Env("ROUTER_SUBDOMAIN", ""), "If specified, a template that should be used to generate the hostname for a route without spec.host (e.g. '${name}-${namespace}.myapps.mycompany.com')")
	flag.BoolVar(&o.OverrideHostname, "override-hostname", cmdutil.Env("ROUTER_OVERRIDE_HOSTNAME", "") == "true", "Override the spec.host value for a route with --hostname-template")
	flag.StringVar(&o.LabelSelector, "labels", cmdutil.Env("ROUTE_LABELS", ""), "A label selector to apply to the routes to watch")
	flag.StringVar(&o.FieldSelector, "fields", cmdutil.Env("ROUTE_FIELDS", ""), "A field selector to apply to routes to watch")
	flag.StringVar(&o.ProjectLabelSelector, "project-labels", cmdutil.Env("PROJECT_LABELS", ""), "A label selector to apply to projects to watch; if '*' watches all projects the client can access")
	flag.StringVar(&o.NamespaceLabelSelector, "namespace-labels", cmdutil.Env("NAMESPACE_LABELS", ""), "A label selector to apply to namespaces to watch")
	flag.BoolVar(&o.IncludeUDP, "include-udp-endpoints", false, "If true, UDP endpoints will be considered as candidates for routing")
	flag.StringSliceVar(&o.DeniedDomains, "denied-domains", envVarAsStrings("ROUTER_DENIED_DOMAINS", "", ","), "List of comma separated domains to deny in routes")
	flag.StringSliceVar(&o.AllowedDomains, "allowed-domains", envVarAsStrings("ROUTER_ALLOWED_DOMAINS", "", ","), "List of comma separated domains to allow in routes. If specified, only the domains in this list will be allowed routes. Note that domains in the denied list take precedence over the ones in the allowed list")
	flag.BoolVar(&o.AllowWildcardRoutes, "allow-wildcard-routes", cmdutil.Env("ROUTER_ALLOW_WILDCARD_ROUTES", "") == "true", "Allow wildcard host names for routes")
	flag.BoolVar(&o.DisableNamespaceOwnershipCheck, "disable-namespace-ownership-check", cmdutil.Env("ROUTER_DISABLE_NAMESPACE_OWNERSHIP_CHECK", "") == "true", "Disables the namespace ownership checks for a route host with different paths or for overlapping host names in the case of wildcard routes. Please be aware that if namespace ownership checks are disabled, routes in a different namespace can use this mechanism to 'steal' sub-paths for existing domains. This is only safe if route creation privileges are restricted, or if all the users can be trusted.")
	flag.BoolVar(&o.EnableIngress, "enable-ingress", cmdutil.Env("ROUTER_ENABLE_INGRESS", "") == "true", "Enable configuration via ingress resources")
	flag.StringVar(&o.ListenAddr, "listen-addr", cmdutil.Env("ROUTER_LISTEN_ADDR", ""), "The name of an interface to listen on to expose metrics and health checking. If not specified, will not listen. Overrides stats port.")
}

// RouteSelectionFunc returns a func that identifies the host for a route.
func (o *RouterSelection) RouteSelectionFunc() controller.RouteHostFunc {
	if len(o.HostnameTemplate) == 0 {
		return controller.HostForRoute
	}
	return func(route *routeapi.Route) string {
		if !o.OverrideHostname && len(route.Spec.Host) > 0 {
			return route.Spec.Host
		}
		// GetNameForHost returns the ingress name for a generated route, and the route route
		// name otherwise.  When a route and ingress in the same namespace share a name, the
		// route and the ingress' rules should receive the same generated host.
		nameForHost := controller.GetNameForHost(route.Name)
		s, err := variable.ExpandStrict(o.HostnameTemplate, func(key string) (string, bool) {
			switch key {
			case "name":
				return nameForHost, true
			case "namespace":
				return route.Namespace, true
			default:
				return "", false
			}
		})
		if err != nil {
			return ""
		}
		return strings.Trim(s, "\"'")
	}
}

func (o *RouterSelection) AdmissionCheck(route *routeapi.Route) error {
	if len(route.Spec.Host) < 1 {
		return nil
	}

	if hostInDomainList(route.Spec.Host, o.BlacklistedDomains) {
		glog.V(4).Infof("host %s in list of denied domains", route.Spec.Host)
		return fmt.Errorf("host in list of denied domains")
	}

	if o.WhitelistedDomains.Len() > 0 {
		glog.V(4).Infof("Checking if host %s is in the list of allowed domains", route.Spec.Host)
		if hostInDomainList(route.Spec.Host, o.WhitelistedDomains) {
			glog.V(4).Infof("host %s admitted - in the list of allowed domains", route.Spec.Host)
			return nil
		}

		glog.V(4).Infof("host %s rejected - not in the list of allowed domains", route.Spec.Host)
		return fmt.Errorf("host not in the allowed list of domains")
	}

	glog.V(4).Infof("host %s admitted", route.Spec.Host)
	return nil
}

// RouteAdmissionFunc returns a func that checks if a route can be admitted
// based on blacklist & whitelist checks and wildcard routes policy setting.
// Note: The blacklist settings trumps the whitelist ones.
func (o *RouterSelection) RouteAdmissionFunc() controller.RouteAdmissionFunc {
	return func(route *routeapi.Route) error {
		if err := o.AdmissionCheck(route); err != nil {
			return err
		}

		switch route.Spec.WildcardPolicy {
		case routeapi.WildcardPolicyNone:
			return nil

		case routeapi.WildcardPolicySubdomain:
			if o.AllowWildcardRoutes {
				return nil
			}
			return fmt.Errorf("wildcard routes are not allowed")
		}

		return fmt.Errorf("unknown wildcard policy %v", route.Spec.WildcardPolicy)
	}
}

// Complete converts string representations of field and label selectors to their parsed equivalent, or
// returns an error.
func (o *RouterSelection) Complete() error {
	if len(o.HostnameTemplate) == 0 && o.OverrideHostname {
		return fmt.Errorf("--override-hostname requires that --hostname-template be specified")
	}
	if len(o.LabelSelector) > 0 {
		s, err := labels.Parse(o.LabelSelector)
		if err != nil {
			return fmt.Errorf("label selector is not valid: %v", err)
		}
		o.Labels = s
	} else {
		o.Labels = labels.Everything()
	}

	if len(o.FieldSelector) > 0 {
		s, err := fields.ParseSelector(o.FieldSelector)
		if err != nil {
			return fmt.Errorf("field selector is not valid: %v", err)
		}
		o.Fields = s
	} else {
		o.Fields = fields.Everything()
	}

	if len(o.ProjectLabelSelector) > 0 {
		if len(o.Namespace) > 0 {
			return fmt.Errorf("only one of --project-labels and --namespace may be used")
		}
		if len(o.NamespaceLabelSelector) > 0 {
			return fmt.Errorf("only one of --namespace-labels and --project-labels may be used")
		}

		if o.ProjectLabelSelector == "*" {
			o.ProjectLabels = labels.Everything()
		} else {
			s, err := labels.Parse(o.ProjectLabelSelector)
			if err != nil {
				return fmt.Errorf("--project-labels selector is not valid: %v", err)
			}
			o.ProjectLabels = s
		}
	}

	if len(o.NamespaceLabelSelector) > 0 {
		if len(o.Namespace) > 0 {
			return fmt.Errorf("only one of --namespace-labels and --namespace may be used")
		}
		s, err := labels.Parse(o.NamespaceLabelSelector)
		if err != nil {
			return fmt.Errorf("--namespace-labels selector is not valid: %v", err)
		}
		o.NamespaceLabels = s
	}

	o.BlacklistedDomains = sets.NewString(o.DeniedDomains...)
	o.WhitelistedDomains = sets.NewString(o.AllowedDomains...)

	return nil
}

// NewFactory initializes a factory that will watch the requested routes
func (o *RouterSelection) NewFactory(routeclient routeclient.RoutesGetter, projectclient projectclient.ProjectResourceInterface, kc kclientset.Interface) *controllerfactory.RouterControllerFactory {
	factory := controllerfactory.NewDefaultRouterControllerFactory(routeclient, kc)
	factory.Labels = o.Labels
	factory.Fields = o.Fields
	factory.Namespace = o.Namespace
	factory.ResyncInterval = o.ResyncInterval
	switch {
	case o.NamespaceLabels != nil:
		glog.Infof("Router is only using routes in namespaces matching %s", o.NamespaceLabels)
		factory.Namespaces = namespaceNames{kc.Core().Namespaces(), o.NamespaceLabels}
	case o.ProjectLabels != nil:
		glog.Infof("Router is only using routes in projects matching %s", o.ProjectLabels)
		factory.Namespaces = projectNames{projectclient, o.ProjectLabels}
	case len(factory.Namespace) > 0:
		glog.Infof("Router is only using resources in namespace %s", factory.Namespace)
	default:
		glog.Infof("Router is including routes in all namespaces")
	}
	return factory
}

// projectNames returns the names of projects matching the label selector
type projectNames struct {
	client   projectclient.ProjectResourceInterface
	selector labels.Selector
}

func (n projectNames) NamespaceNames() (sets.String, error) {
	all, err := n.client.List(metav1.ListOptions{LabelSelector: n.selector.String()})
	if err != nil {
		return nil, err
	}
	names := make(sets.String, len(all.Items))
	for i := range all.Items {
		names.Insert(all.Items[i].Name)
	}
	return names, nil
}

// namespaceNames returns the names of namespaces matching the label selector
type namespaceNames struct {
	client   kcoreclient.NamespaceInterface
	selector labels.Selector
}

func (n namespaceNames) NamespaceNames() (sets.String, error) {
	all, err := n.client.List(metav1.ListOptions{LabelSelector: n.selector.String()})
	if err != nil {
		return nil, err
	}
	names := make(sets.String, len(all.Items))
	for i := range all.Items {
		names.Insert(all.Items[i].Name)
	}
	return names, nil
}

func envVarAsStrings(name, defaultValue, separator string) []string {
	strlist := []string{}
	if env := cmdutil.Env(name, defaultValue); env != "" {
		values := strings.Split(env, separator)
		for i := range values {
			if val := strings.TrimSpace(values[i]); val != "" {
				strlist = append(strlist, val)
			}
		}
	}
	return strlist
}

func hostInDomainList(host string, domains sets.String) bool {
	if domains.Has(host) {
		return true
	}

	if idx := strings.IndexRune(host, '.'); idx > 0 {
		return hostInDomainList(host[idx+1:], domains)
	}

	return false
}
