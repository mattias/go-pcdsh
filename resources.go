package main

import (
	"github.com/emicklei/go-restful"
	"io"
)

type LogResource struct {}

func (l LogResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/logs").
		Consumes("*/*").
		Produces("*/*")

	ws.Route(ws.GET("/{index}").To(l.nop))

	container.Add(ws)
}

func (l LogResource) nop(request *restful.Request, response *restful.Response) {
	index := request.PathParameter("index")
	io.WriteString(response.ResponseWriter, index)
	io.WriteString(response.ResponseWriter, "this would be a normal response")
}

