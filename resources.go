package main

import (
	"github.com/emicklei/go-restful"
	"time"
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type SessionResource struct {
	sessions []Session
}

func (s SessionResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
	Path("/session").
	Consumes(restful.MIME_JSON).
	Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(s.getAllSessions))
	ws.Route(ws.GET("/{id}").To(s.getSessionById))

	container.Add(ws)
}

func (s SessionResource) getAllSessions(request *restful.Request, response *restful.Response) {
	// TODO: Render all or have a limit + pagination?
	start := time.Now()
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

	s.sessions = make([]Session, 0)

	for sessionRows.Next() {
		err = sessionRows.Scan(&sessionId, &logStartId, &logEndId, &logStartTime, &logEndTime, &logCount)
		if err != nil {
			log.Println(err.Error())
		}

		session := Session{Id: sessionId, StartLogId: logStartId, EndLogId: logEndId, StartTime: logStartTime, EndTime: logEndTime, LogCount: logCount}
		s.sessions = append(s.sessions, session)
	}
	response.WriteEntity(s.sessions)
	elapsed := time.Since(start)
	log.Printf("Render getAllSessions took %s", elapsed)
}

func (s SessionResource) getSessionById(request *restful.Request, response *restful.Response) {
	start := time.Now()
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

		response.WriteErrorString(404, "404 page not found")
		return
	}

	response.WriteEntity(Session{Id: sessionId, StartLogId: logStartId, EndLogId: logEndId, StartTime: logStartTime, EndTime: logEndTime, LogCount: logCount })

	elapsed := time.Since(start)
	log.Printf("Render getSessionById took %s", elapsed)
}
