package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"
)

func hashSessions(configuration Configuration) {
	log.Println("Hasher starting...")
	db, err := sql.Open("mysql", configuration.Datasource)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	sessionsOut, err := db.Prepare("SELECT end_time FROM `sessions` ORDER BY end_time DESC LIMIT 1")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsOut.Close()

	logsOut, err := db.Prepare("SELECT id, name, time FROM `logs` WHERE `time` > ? ORDER BY `time` ASC, `index` ASC")
	if err != nil {
		log.Println(err.Error())
	}
	defer logsOut.Close()

	logAttributesOut, err := db.Prepare("SELECT `key`, value FROM log_attributes WHERE log_id = ?")
	if err != nil {
		log.Println(err.Error())
	}
	defer logAttributesOut.Close()

	sessionsIns, err := db.Prepare("INSERT INTO sessions VALUES( ?, ?, ?, ?, ?, ?, ?, ? )")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsIns.Close()

	var (
		logId             int64
		logName           string
		logTime           time.Time
		logCount          int64
		logStartId        int64
		logStartTime      time.Time
		logEndId          int64
		logEndTime        time.Time
		logTrackId        int
		start             time.Time
		end_time          time.Time
		elapsed           time.Duration
		logAttributeKey   string
		logAttributeValue string
		valid             int
	)

	for {
		start = time.Now()
		logStartId, logEndId, logCount, logTrackId, valid = 0, 0, 0, 0, 0 // Make sure they are reset for this next loop

		err = sessionsOut.QueryRow().Scan(&end_time)
		if err != nil {
			log.Println(err.Error())
		}

		logRows, err := logsOut.Query(end_time)
		if err != nil {
			log.Println(err.Error())
		}

		for logRows.Next() {
			err = logRows.Scan(&logId, &logName, &logTime)
			if err != nil {
				log.Println(err.Error())
			}

			logCount++

			if logName == "Results" {
				valid = 1
			}

			if logName == "SessionSetup" {
				logAttributeRows, err := logAttributesOut.Query(logId)
				if err != nil {
					log.Println(err.Error())
				}

				for logAttributeRows.Next() {
					err = logAttributeRows.Scan(&logAttributeKey, &logAttributeValue)
					if err != nil {
						log.Println(err.Error())
					}

					if logAttributeKey == "TrackId" {
						logTrackId, _ = strconv.Atoi(logAttributeValue)
						break
					}
				}
				logAttributeRows.Close()
			}

			if logName == "StateChanged" {
				logAttributeRows, err := logAttributesOut.Query(logId)
				if err != nil {
					log.Println(err.Error())
				}

				for logAttributeRows.Next() {
					err = logAttributeRows.Scan(&logAttributeKey, &logAttributeValue)
					if err != nil {
						log.Println(err.Error())
					}

					if logAttributeKey == "NewState" && logAttributeValue == "Loading" {
						logStartId, logStartTime = logId, logTime
						logEndId, logCount, logTrackId, valid = 0, 0, 0, 0 // Reset
					}

					if logAttributeKey == "NewState" && logAttributeValue == "Returning" {
						logEndId, logEndTime = logId, logTime
					}
				}
				logAttributeRows.Close()
			} else if logName == "SessionDestroyed" {
				logEndId, logEndTime = logId, logTime
			}

			if logStartId > 0 && logEndId > 0 {
				_, err = sessionsIns.Exec(nil, logStartId, logEndId, logStartTime, logEndTime, logTrackId, logCount, valid)
				if err != nil {
					log.Println(err.Error())
				}
				logStartId, logEndId, logCount, logTrackId, valid = 0, 0, 0, 0, 0
			}
		}

		logRows.Close()

		elapsed = time.Since(start)
		log.Printf("Hashing sessions took %s", elapsed)
		log.Println("Hasher idling for 1 hour")
		time.Sleep(time.Hour)
	}
}
