package rest

import (
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/api"

	buildapi "github.com/openshift/origin/pkg/build/apis/build"
)

// HookHandler is a Kubernetes API compatible webhook that is able to get access to the raw request
// and response. Used when adapting existing webhook code to the Kubernetes patterns.
type HookHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, ctx apirequest.Context, name, subpath string) error
}

type httpHookHandler struct {
	http.Handler
}

func (h httpHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, ctx apirequest.Context, name, subpath string) error {
	h.Handler.ServeHTTP(w, r)
	return nil
}

// WebHook provides a reusable rest.Storage implementation for linking a generic WebHook handler
// into the Kube API pattern. It is intended to be used with GET or POST against a resource's
// named path, possibly as a subresource. The handler has access to the extracted information
// from the Kube apiserver including the context, the name, and the subpath.
type WebHook struct {
	h        HookHandler
	allowGet bool
}

var _ rest.Connecter = &WebHook{}

// NewWebHook creates an adapter that implements rest.Connector for the given HookHandler.
func NewWebHook(handler HookHandler, allowGet bool) *WebHook {
	return &WebHook{
		h:        handler,
		allowGet: allowGet,
	}
}

// NewHTTPWebHook creates an adapter that implements rest.Connector for the given http.Handler.
func NewHTTPWebHook(handler http.Handler, allowGet bool) *WebHook {
	return &WebHook{
		h:        httpHookHandler{handler},
		allowGet: allowGet,
	}
}

// New() responds with the status object.
func (h *WebHook) New() runtime.Object {
	return &buildapi.Build{}
}

// Connect responds to connections with a ConnectHandler
func (h *WebHook) Connect(ctx apirequest.Context, name string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	return &WebHookHandler{
		handler:   h.h,
		ctx:       ctx,
		name:      name,
		options:   options.(*api.PodProxyOptions),
		responder: responder,
	}, nil
}

// NewConnectionOptions identifies the options that should be passed to this hook
func (h *WebHook) NewConnectOptions() (runtime.Object, bool, string) {
	return &api.PodProxyOptions{}, true, "path"
}

// ConnectMethods returns the supported web hook types.
func (h *WebHook) ConnectMethods() []string {
	if h.allowGet {
		return []string{"GET", "POST"}
	}
	return []string{"POST"}
}

// WebHookHandler responds to web hook requests from the master.
type WebHookHandler struct {
	handler   HookHandler
	ctx       apirequest.Context
	name      string
	options   *api.PodProxyOptions
	responder rest.Responder
}

var _ http.Handler = &WebHookHandler{}

func (h *WebHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.handler.ServeHTTP(w, r, h.ctx, h.name, h.options.Path); err != nil {
		h.responder.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
