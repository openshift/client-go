package origin

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"

	genericmux "k8s.io/apiserver/pkg/server/mux"

	"github.com/openshift/origin/pkg/cmd/util/plug"
	oauthutil "github.com/openshift/origin/pkg/oauth/util"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

type OpenshiftNonAPIConfig struct {
	GenericConfig *genericapiserver.Config

	// these are only needed for the controller endpoint which should be moved out and made an optional
	// add-on in the chain (as the final delegate) when running an all-in-one
	ControllerPlug plug.Plug

	MasterPublicURL string
	EnableOAuth     bool
}

// OpenshiftNonAPIServer serves non-API endpoints for openshift.
type OpenshiftNonAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedOpenshiftNonAPIConfig struct {
	*OpenshiftNonAPIConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *OpenshiftNonAPIConfig) Complete() completedOpenshiftNonAPIConfig {
	c.GenericConfig.Complete()

	return completedOpenshiftNonAPIConfig{c}
}

// SkipComplete provides a way to construct a server instance without config completion.
func (c *OpenshiftNonAPIConfig) SkipComplete() completedOpenshiftNonAPIConfig {
	return completedOpenshiftNonAPIConfig{c}
}

func (c completedOpenshiftNonAPIConfig) New(delegationTarget genericapiserver.DelegationTarget) (*OpenshiftNonAPIServer, error) {
	genericServer, err := c.OpenshiftNonAPIConfig.GenericConfig.SkipComplete().New("openshift-non-api-routes", delegationTarget) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}

	s := &OpenshiftNonAPIServer{
		GenericAPIServer: genericServer,
	}

	// TODO punt this out to its own "unrelated gorp" delegation target.  It is not related to API
	initControllerRoutes(s.GenericAPIServer.Handler.GoRestfulContainer, "/controllers", c.ControllerPlug)

	// TODO move this up to the spot where we wire the oauth endpoint
	// Set up OAuth metadata only if we are configured to use OAuth
	if c.EnableOAuth {
		initOAuthAuthorizationServerMetadataRoute(s.GenericAPIServer.Handler.NonGoRestfulMux, oauthMetadataEndpoint, c.MasterPublicURL)
	}

	return s, nil
}

const (
	// Discovery endpoint for OAuth 2.0 Authorization Server Metadata
	// See IETF Draft:
	// https://tools.ietf.org/html/draft-ietf-oauth-discovery-04#section-2
	oauthMetadataEndpoint = "/.well-known/oauth-authorization-server"
)

// initOAuthAuthorizationServerMetadataRoute initializes an HTTP endpoint for OAuth 2.0 Authorization Server Metadata discovery
// https://tools.ietf.org/id/draft-ietf-oauth-discovery-04.html#rfc.section.2
// masterPublicURL should be internally and externally routable to allow all users to discover this information
func initOAuthAuthorizationServerMetadataRoute(mux *genericmux.PathRecorderMux, path, masterPublicURL string) {
	// Build OAuth metadata once
	metadata, err := json.MarshalIndent(oauthutil.GetOauthMetadata(masterPublicURL), "", "  ")
	if err != nil {
		glog.Errorf("Unable to initialize OAuth authorization server metadata route: %v", err)
		return
	}

	mux.UnlistedHandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(metadata)
	})
}
