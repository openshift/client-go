package origin

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"

	"github.com/openshift/origin/pkg/cmd/util/plug"
)

// initControllerRoutes adds a web service endpoint for managing the execution
// state of the controllers.
func initControllerRoutes(apiContainer *restful.Container, path string, plug plug.Plug) {
	canStart := plug != nil

	ws := new(restful.WebService).
		Path(path).
		Doc("Check whether the controllers are running on this master")
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		if !canStart {
			resp.ResponseWriter.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(resp, "disabled")
			return
		}
		if plug.IsStarted() {
			resp.ResponseWriter.WriteHeader(http.StatusOK)
			fmt.Fprintf(resp, "ok")
		} else {
			resp.ResponseWriter.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(resp, "waiting")
		}
	}).Doc("Check whether the controllers are running on this master").
		Returns(http.StatusOK, "if controllers are running", nil).
		Returns(http.StatusMethodNotAllowed, "if controllers are disabled", nil).
		Returns(http.StatusAccepted, "if controllers are waiting to be started", nil).
		Produces(restful.MIME_JSON))

	ws.Route(ws.PUT(path).To(func(req *restful.Request, resp *restful.Response) {
		if !canStart {
			resp.ResponseWriter.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(resp, "disabled")
			return
		}
		plug.Start()
		resp.ResponseWriter.WriteHeader(http.StatusOK)
		fmt.Fprintf(resp, "ok")
	}).Doc("Start controllers on this master").
		Returns(http.StatusOK, "if controllers have started", nil).
		Returns(http.StatusMethodNotAllowed, "if controllers are disabled", nil).
		Produces(restful.MIME_JSON))

	ws.Route(ws.DELETE(path).To(func(req *restful.Request, resp *restful.Response) {
		resp.ResponseWriter.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(resp, "terminating")
		plug.Stop(nil)
	}).Doc("Stop the master").
		Returns(http.StatusAccepted, "if the master will stop", nil).
		Produces(restful.MIME_JSON))

	apiContainer.Add(ws)
}
