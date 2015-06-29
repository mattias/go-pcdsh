package main

import (
	"github.com/emicklei/go-restful"
	"io"
)

type SessionResource struct {}

func (s SessionResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/session").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/{id}").To(s.getSessionById))
	ws.Route(ws.GET("/date").To(s.getSessionByDate))

	container.Add(ws)
}

func (s SessionResource) getSessionById(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	io.WriteString(response.ResponseWriter, id)
}

func (s SessionResource) getSessionByDate(request *restful.Request, response *restful.Response) {
	start := request.QueryParameter("start")
	end := request.QueryParameter("end")
	io.WriteString(response.ResponseWriter, start + " : " + end)
}
