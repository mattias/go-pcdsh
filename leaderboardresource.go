package main

import (
	"github.com/emicklei/go-restful"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type LeaderboardResource struct {
}

func (l LeaderboardResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/leaderboard").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// TODO: Get all tracks, cars and different users available (which has a record)
	ws.Route(ws.GET("/").To(l.getAll))
	// TODO: List all records for track
	// TODO: List all records for track + car combo
	// TODO: List all records for user

	container.Add(ws)
}

func (l LeaderboardResource) getAll(request *restful.Request, response *restful.Response) {
	start := time.Now()

	elapsed := time.Since(start)
	log.Printf("Render LeaderboardResource getAll took %s", elapsed)
}
