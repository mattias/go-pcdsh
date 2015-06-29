package main

import (
	"github.com/emicklei/go-restful"
	"time"
	"log"
	"io"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func (s SessionResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
	Path("/session").
	Consumes(restful.MIME_JSON).
	Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(s.getAllSessions))
	ws.Route(ws.GET("/{id}").To(s.getSessionById))
	ws.Route(ws.GET("/date/{date}").To(s.getSessionByDate))

	container.Add(ws)
}

func (s SessionResource) getAllSessions(request *restful.Request, response *restful.Response) {
	configuration := readConfiguration()

	db, err := sql.Open("mysql", configuration.Datasource)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	sessionsOut, err := db.Prepare("SELECT * FROM sessions ORDER BY `id` DESC")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsOut.Close()

	sessionRows, err := sessionsOut.Query()
	if err != nil {
		log.Println(err.Error())
	}

	var (
		sessionId int64
		logStartId int64
		logEndId int64
		logStartTime time.Time
		logEndTime time.Time
		logCount int64
	)

	for sessionRows.Next() {
		err = sessionRows.Scan(&sessionId, &logStartId, &logEndId, &logStartTime, &logEndTime, &logCount)
		if err != nil {
			log.Println(err.Error())
		}

		session := Session{Id: sessionId, StartLogId: logStartId, EndLogId: logEndId, StartTime: logStartTime, EndTime: logEndTime, LogCount: logCount}
		// TODO: try to feed session variable into a slice or something, then have WriteEntity write out the entire slice
		// TODO: As it is now, this won't do, fix!
		response.WriteEntity(session)
	}
}

func (s SessionResource) getSessionById(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	configuration := readConfiguration()

	db, err := sql.Open("mysql", configuration.Datasource)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	sessionsOut, err := db.Prepare("SELECT * FROM sessions WHERE id = ?")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsOut.Close()

	var (
		sessionId int64
		logStartId int64
		logEndId int64
		logStartTime time.Time
		logEndTime time.Time
		logCount int64
	)

	err = sessionsOut.QueryRow(id).Scan(&sessionId, &logStartId, &logEndId, &logStartTime, &logEndTime, &logCount)
	if err != nil {
		log.Println(err.Error())
	}

	response.WriteEntity(Session{Id: sessionId, StartLogId: logStartId, EndLogId: logEndId, StartTime: logStartTime, EndTime: logEndTime, LogCount: logCount })
}

func (s SessionResource) getSessionByDate(request *restful.Request, response *restful.Response) {
	start := request.QueryParameter("start")
	end := request.QueryParameter("end")
	io.WriteString(response.ResponseWriter, start + " : " + end)
}
