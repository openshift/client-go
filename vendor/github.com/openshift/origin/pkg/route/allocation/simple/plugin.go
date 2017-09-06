package simple

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"

	routeapi "github.com/openshift/origin/pkg/route/apis/route"
)

// Default DNS suffix to use if no configuration is passed to this plugin.
const defaultDNSSuffix = "router.default.svc.cluster.local"

// SimpleAllocationPlugin implements the route.AllocationPlugin interface
// to provide a simple unsharded (or single sharded) allocation plugin.
type SimpleAllocationPlugin struct {
	DNSSuffix string
}

// NewSimpleAllocationPlugin creates a new SimpleAllocationPlugin.
func NewSimpleAllocationPlugin(suffix string) (*SimpleAllocationPlugin, error) {
	if len(suffix) == 0 {
		suffix = defaultDNSSuffix
	}

	glog.V(4).Infof("Route plugin initialized with suffix=%s", suffix)

	// Check that the DNS suffix is valid.
	if len(kvalidation.IsDNS1123Subdomain(suffix)) != 0 {
		return nil, fmt.Errorf("invalid DNS suffix: %s", suffix)
	}

	return &SimpleAllocationPlugin{DNSSuffix: suffix}, nil
}

// Allocate a router shard for the given route. This plugin always returns
// the "global" router shard.
// TODO: replace with per router allocation
func (p *SimpleAllocationPlugin) Allocate(route *routeapi.Route) (*routeapi.RouterShard, error) {
	glog.V(4).Infof("Allocating global shard *.%s to Route: %s", p.DNSSuffix, route.Name)

	return &routeapi.RouterShard{ShardName: "global", DNSSuffix: p.DNSSuffix}, nil
}

// GenerateHostname generates a host name for a route - using the service name,
// namespace (if provided) and the router shard dns suffix.
// TODO: move to router code, and have the routers set this back on the route status.
func (p *SimpleAllocationPlugin) GenerateHostname(route *routeapi.Route, shard *routeapi.RouterShard) string {
	if len(route.Name) == 0 || len(route.Namespace) == 0 {
		return ""
	}
	return fmt.Sprintf("%s-%s.%s", strings.Replace(route.Name, ".", "-", -1), route.Namespace, shard.DNSSuffix)
}
