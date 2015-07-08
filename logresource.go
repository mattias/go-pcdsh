package main

import (
	"database/sql"
	"github.com/emicklei/go-restful"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type Log struct {
	Id, Index, Refid, Participantid int64
	Name                            string
	Time                            time.Time
	Attributes                      map[string]string
}

type LogResource struct {
	logs []Log
}

func (l LogResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/log").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(l.getLatestLogs))
	ws.Route(ws.GET("/session/{id}").To(l.getLogRangeBySessionId))

	container.Add(ws)
}

func (l LogResource) getLogRangeBySessionId(request *restful.Request, response *restful.Response) {
	start := time.Now()
	sessionId := request.PathParameter("id")
	var startId, endId int64
	var logCache interface{}

	logCache, err := GetCache("/log/session/" + sessionId)
	if err == nil {
		l = logCache.(LogResource)
		response.WriteEntity(l.logs)
		elapsed := time.Since(start)
		log.Printf("Render getLogRangeBySessionId took %s", elapsed)
		return
	} else {
		log.Println(err)
	}

	configuration := readConfiguration()

	var (
		logId             int64
		logIndex          int64
		logTime           time.Time
		logName           string
		logRefid          int64
		logParticipantid  int64
		logAttributes     map[string]string
		logAttributeKey   string
		logAttributeValue string
	)

	l.logs = make([]Log, 0)

	db, err := sql.Open("mysql", configuration.Datasource)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	sessionsOut, err := db.Prepare("SELECT start_log_id, end_log_id FROM sessions WHERE id = ? AND valid = 1")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsOut.Close()

	logAttributesOut, err := db.Prepare("SELECT `key`, value FROM log_attributes WHERE log_id = ?")
	if err != nil {
		log.Println(err.Error())
	}
	defer logAttributesOut.Close()

	logsOut, err := db.Prepare("SELECT * FROM logs WHERE id >= ? AND id <= ? ORDER BY `id` ASC")
	if err != nil {
		log.Println(err.Error())
	}
	defer logsOut.Close()

	err = sessionsOut.QueryRow(sessionId).Scan(&startId, &endId)
	if err != nil {
		log.Println(err.Error())
	}

	logRows, err := logsOut.Query(startId, endId)
	if err != nil {
		log.Println(err.Error())
	}

	for logRows.Next() {
		err = logRows.Scan(&logId, &logIndex, &logTime, &logName, &logRefid, &logParticipantid)
		if err != nil {
			log.Println(err.Error())
		}

		logAttributeRows, err := logAttributesOut.Query(logId)
		if err != nil {
			log.Println(err.Error())
		}
		logAttributes = make(map[string]string)
		for logAttributeRows.Next() {
			err = logAttributeRows.Scan(&logAttributeKey, &logAttributeValue)
			if err != nil {
				log.Println(err.Error())
			}

			logAttributes[logAttributeKey] = logAttributeValue
		}

		logAttributeRows.Close()

		log := Log{Id: logId, Index: logIndex, Time: logTime, Name: logName, Refid: logRefid, Participantid: logParticipantid, Attributes: logAttributes}
		l.logs = append(l.logs, log)
	}

	logRows.Close()

	response.WriteEntity(l.logs)
	SetCache("/log/session/"+sessionId, l)

	elapsed := time.Since(start)
	log.Printf("Render getLogRangeBySessionId took %s", elapsed)
}

func (l LogResource) getLatestLogs(request *restful.Request, response *restful.Response) {
	// TODO: Render all or have a limit + pagination?
	start := time.Now()
	configuration := readConfiguration()

	var (
		logId             int64
		logIndex          int64
		logTime           time.Time
		logName           string
		logRefid          int64
		logParticipantid  int64
		logAttributes     map[string]string
		logAttributeKey   string
		logAttributeValue string
	)

	l.logs = make([]Log, 0)

	db, err := sql.Open("mysql", configuration.Datasource)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	logAttributesOut, err := db.Prepare("SELECT `key`, value FROM log_attributes WHERE log_id = ?")
	if err != nil {
		log.Println(err.Error())
	}
	defer logAttributesOut.Close()

	logsOut, err := db.Prepare("SELECT * FROM logs ORDER BY `id` DESC LIMIT 10")
	if err != nil {
		log.Println(err.Error())
	}
	defer logsOut.Close()

	logRows, err := logsOut.Query()
	if err != nil {
		log.Println(err.Error())
	}

	for logRows.Next() {
		err = logRows.Scan(&logId, &logIndex, &logTime, &logName, &logRefid, &logParticipantid)
		if err != nil {
			log.Println(err.Error())
		}

		logAttributeRows, err := logAttributesOut.Query(logId)
		if err != nil {
			log.Println(err.Error())
		}
		logAttributes = make(map[string]string)
		for logAttributeRows.Next() {
			err = logAttributeRows.Scan(&logAttributeKey, &logAttributeValue)
			if err != nil {
				log.Println(err.Error())
			}

			logAttributes[logAttributeKey] = logAttributeValue
		}

		logAttributeRows.Close()

		log := Log{Id: logId, Index: logIndex, Time: logTime, Name: logName, Refid: logRefid, Participantid: logParticipantid, Attributes: logAttributes}
		l.logs = append(l.logs, log)
	}

	logRows.Close()

	response.WriteEntity(l.logs)
	elapsed := time.Since(start)
	log.Printf("Render getLatestLogs took %s", elapsed)
}
