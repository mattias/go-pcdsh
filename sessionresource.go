package main

import (
	"database/sql"
	"github.com/emicklei/go-restful"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

type Session struct {
	Id, StartLogId, EndLogId, TrackId, LogCount, Valid int
	StartTime, EndTime                                 time.Time
}

type Lap struct {
	CountThisLapTimes, DistanceTravelled, Lap, LapTime, RacePosition, Sector1Time, Sector2Time, Sector3Time int
}

type Result struct {
	FastestLapTime, Lap, RacePosition, TotalTime, VehicleId int
	State                                                   string
}

type Impact struct {
	Name                                   string
	CollisionMagnitude, OtherParticipantId int
}

type CutTrackStart struct {
	Name                                     string
	IsMainBranch, Lap, LapTime, RacePosition int
}

type CutTrackEnd struct {
	Name                                                                string
	ElapsedTime, PenaltyThreshold, PenaltyValue, PlaceGain, SkippedTime int
}

type SessionSetup struct {
	Flags, GameMode, GridSize, MaxPlayers, Practice1Length, Practice2Length, QualifyLength, Race1Length, Race2Length, TrackId, WarmupLength int
}

type Stage struct {
	Name      string
	Laps      []Lap
	Incidents []interface{}
	Result    Result
}

type Participant struct {
	Id     int
	Name   string
	Refid  int
	Stages map[string]*Stage
}

type CompiledSession struct {
	Setup        SessionSetup
	Participants []Participant
}

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
	ws.Route(ws.GET("/compiled/{id}").To(s.getCompiledSessionById))

	container.Add(ws)
}

func (s SessionResource) getCompiledSessionById(request *restful.Request, response *restful.Response) {
	start := time.Now()
	sessionId := request.PathParameter("id")
	var compiledSessionCache interface{}
	var compiledSession CompiledSession

	compiledSessionCache, err := GetCache("/session/compiled/" + sessionId)
	if err == nil {
		compiledSession = compiledSessionCache.(CompiledSession)
		response.WriteEntity(compiledSession)
		elapsed := time.Since(start)
		log.Printf("Render getCompiledSessionById took %s", elapsed)
		return
	} else {
		log.Println(err)
	}

	var (
		logId             int
		logIndex          int
		logTime           time.Time
		logName           string
		logRefid          int
		logParticipantid  int
		logAttributes     map[string]string
		logAttributeKey   string
		logAttributeValue string
		startId, endId    int
		curSessionStage   string = "Practice1"
		sessionStages            = []string{"Practice1", "Practice2", "Qualifying", "Warmup", "Race1", "Race2"}
	)
	compiledSession.Participants = make([]Participant, 0)

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

		switch logName {
		case "StageChanged":
			curSessionStage = logAttributes["NewStage"]
		case "SessionSetup":
			if compiledSession.Setup.Flags != 0 {
				break
			}
			Flags, _ := strconv.Atoi(logAttributes["Flags"])
			GameMode, _ := strconv.Atoi(logAttributes["GameMode"])
			GridSize, _ := strconv.Atoi(logAttributes["GridSize"])
			MaxPlayers, _ := strconv.Atoi(logAttributes["MaxPlayers"])
			Practice1Length, _ := strconv.Atoi(logAttributes["Practice1Length"])
			Practice2Length, _ := strconv.Atoi(logAttributes["Practice2Length"])
			QualifyLength, _ := strconv.Atoi(logAttributes["QualifyLength"])
			Race1Length, _ := strconv.Atoi(logAttributes["Race1Length"])
			Race2Length, _ := strconv.Atoi(logAttributes["Race2Length"])
			TrackId, _ := strconv.Atoi(logAttributes["TrackId"])
			WarmupLength, _ := strconv.Atoi(logAttributes["WarmupLength"])
			compiledSession.Setup = SessionSetup{
				Flags:           Flags,
				GameMode:        GameMode,
				GridSize:        GridSize,
				MaxPlayers:      MaxPlayers,
				Practice1Length: Practice1Length,
				Practice2Length: Practice2Length,
				QualifyLength:   QualifyLength,
				Race1Length:     Race1Length,
				Race2Length:     Race2Length,
				TrackId:         TrackId,
				WarmupLength:    WarmupLength,
			}
		case "ParticipantCreated":
			compiledSession.Participants = append(compiledSession.Participants, Participant{
				Stages: make(map[string]*Stage),
				Id:     logParticipantid,
				Name:   logAttributes["Name"],
				Refid:  logRefid,
			})
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					for _, stage := range sessionStages {
						compiledSession.Participants[key].Stages[stage] = &Stage{}
						compiledSession.Participants[key].Stages[stage].Name = stage
						compiledSession.Participants[key].Stages[stage].Laps = make([]Lap, 0)
						compiledSession.Participants[key].Stages[stage].Incidents = make([]interface{}, 0)
					}
				}
			}
		case "PlayerLeft":
			fallthrough
		case "ParticipantDestroyed":
			var sliceIndex int = -1
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					sliceIndex = key
					break
				}
			}
			if sliceIndex >= 0 {
				compiledSession.Participants = append(compiledSession.Participants[:sliceIndex], compiledSession.Participants[sliceIndex+1:]...)
			}
		case "Lap":
			CountThisLapTimes, _ := strconv.Atoi(logAttributes["CountThisLapTimes"])
			DistanceTravelled, _ := strconv.Atoi(logAttributes["DistanceTravelled"])
			lap, _ := strconv.Atoi(logAttributes["Lap"])
			LapTime, _ := strconv.Atoi(logAttributes["LapTime"])
			RacePosition, _ := strconv.Atoi(logAttributes["RacePosition"])
			Sector1Time, _ := strconv.Atoi(logAttributes["Sector1Time"])
			Sector2Time, _ := strconv.Atoi(logAttributes["Sector2Time"])
			Sector3Time, _ := strconv.Atoi(logAttributes["Sector3Time"])
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					compiledSession.Participants[key].Stages[curSessionStage].Laps = append(compiledSession.Participants[key].Stages[curSessionStage].Laps, Lap{
						CountThisLapTimes: CountThisLapTimes,
						DistanceTravelled: DistanceTravelled,
						Lap:               lap,
						LapTime:           LapTime,
						RacePosition:      RacePosition,
						Sector1Time:       Sector1Time,
						Sector2Time:       Sector2Time,
						Sector3Time:       Sector3Time,
					})
				}
			}
		case "Results":
			FastestLapTime, _ := strconv.Atoi(logAttributes["FastestLapTime"])
			Lap, _ := strconv.Atoi(logAttributes["Lap"])
			RacePosition, _ := strconv.Atoi(logAttributes["RacePosition"])
			TotalTime, _ := strconv.Atoi(logAttributes["TotalTime"])
			VehicleId, _ := strconv.Atoi(logAttributes["VehicleId"])
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					compiledSession.Participants[key].Stages[curSessionStage].Result = Result{
						FastestLapTime: FastestLapTime,
						Lap:            Lap,
						RacePosition:   RacePosition,
						TotalTime:      TotalTime,
						VehicleId:      VehicleId,
						State:          logAttributes["State"],
					}
				}
			}
		case "Impact":
			CollisionMagnitude, _ := strconv.Atoi(logAttributes["CollisionMagnitude"])
			OtherParticipantId, _ := strconv.Atoi(logAttributes["OtherParticipantId"])
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					compiledSession.Participants[key].Stages[curSessionStage].Incidents = append(compiledSession.Participants[key].Stages[curSessionStage].Incidents, Impact{
						Name:               "Impact",
						CollisionMagnitude: CollisionMagnitude,
						OtherParticipantId: OtherParticipantId,
					})
				}
			}
		case "CutTrackStart":
			IsMainBranch, _ := strconv.Atoi(logAttributes["IsMainBranch"])
			Lap, _ := strconv.Atoi(logAttributes["Lap"])
			LapTime, _ := strconv.Atoi(logAttributes["LapTime"])
			RacePosition, _ := strconv.Atoi(logAttributes["RacePosition"])
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					compiledSession.Participants[key].Stages[curSessionStage].Incidents = append(compiledSession.Participants[key].Stages[curSessionStage].Incidents, CutTrackStart{
						Name:         "CutTrackStart",
						IsMainBranch: IsMainBranch,
						Lap:          Lap,
						LapTime:      LapTime,
						RacePosition: RacePosition,
					})
				}
			}
		case "CutTrackEnd":
			ElapsedTime, _ := strconv.Atoi(logAttributes["ElapsedTime"])
			PenaltyThreshold, _ := strconv.Atoi(logAttributes["PenaltyThreshold"])
			PenaltyValue, _ := strconv.Atoi(logAttributes["PenaltyValue"])
			PlaceGain, _ := strconv.Atoi(logAttributes["PlaceGain"])
			SkippedTime, _ := strconv.Atoi(logAttributes["SkippedTime"])
			for key := range compiledSession.Participants {
				if compiledSession.Participants[key].Id == logParticipantid {
					compiledSession.Participants[key].Stages[curSessionStage].Incidents = append(compiledSession.Participants[key].Stages[curSessionStage].Incidents, CutTrackEnd{
						Name:             "CutTrackEnd",
						ElapsedTime:      ElapsedTime,
						PenaltyThreshold: PenaltyThreshold,
						PenaltyValue:     PenaltyValue,
						PlaceGain:        PlaceGain,
						SkippedTime:      SkippedTime,
					})
				}
			}
		}
	}

	logRows.Close()

	response.WriteEntity(compiledSession)
	SetCache("/session/compiled/"+sessionId, compiledSession)

	elapsed := time.Since(start)
	log.Printf("Render getCompiledSessionById took %s", elapsed)
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

	sessionsOut, err := db.Prepare("SELECT * FROM sessions ORDER BY `id` DESC WHERE valid = 1")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsOut.Close()

	sessionRows, err := sessionsOut.Query()
	if err != nil {
		log.Println(err.Error())
	}

	var (
		sessionId    int
		logStartId   int
		logEndId     int
		logStartTime time.Time
		logEndTime   time.Time
		logTrackId   int
		logCount     int
	)

	s.sessions = make([]Session, 0)

	for sessionRows.Next() {
		err = sessionRows.Scan(&sessionId, &logStartId, &logEndId, &logStartTime, &logEndTime, &logTrackId, &logCount)
		if err != nil {
			log.Println(err.Error())
		}

		session := Session{Id: sessionId, StartLogId: logStartId, EndLogId: logEndId, StartTime: logStartTime, EndTime: logEndTime, TrackId: logTrackId, LogCount: logCount}
		s.sessions = append(s.sessions, session)
	}

	sessionRows.Close()

	response.WriteEntity(s.sessions)

	elapsed := time.Since(start)
	log.Printf("Render getAllSessions took %s", elapsed)
}

func (s SessionResource) getSessionById(request *restful.Request, response *restful.Response) {
	start := time.Now()
	id := request.PathParameter("id")
	var sessionCache interface{}

	sessionCache, err := GetCache("/session/" + id)
	if err == nil {
		session := sessionCache.(Session)
		response.WriteEntity(session)
		elapsed := time.Since(start)
		log.Printf("Render getSessionById took %s", elapsed)
		return
	} else {
		log.Println(err)
	}

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

	sessionsOut, err := db.Prepare("SELECT * FROM sessions WHERE id = ? AND valid = 1")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsOut.Close()

	var (
		sessionId    int
		logStartId   int
		logEndId     int
		logStartTime time.Time
		logEndTime   time.Time
		logTrackId   int
		logCount     int
	)

	err = sessionsOut.QueryRow(id).Scan(&sessionId, &logStartId, &logEndId, &logStartTime, &logEndTime, &logTrackId, &logCount)
	if err != nil {
		log.Println(err.Error())

		response.WriteErrorString(404, "404 page not found")
		return
	}

	session := Session{Id: sessionId, StartLogId: logStartId, EndLogId: logEndId, StartTime: logStartTime, EndTime: logEndTime, TrackId: logTrackId, LogCount: logCount}

	response.WriteEntity(session)
	SetCache("/session/"+id, session)

	elapsed := time.Since(start)
	log.Printf("Render getSessionById took %s", elapsed)
}
