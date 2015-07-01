package main
import (
	"time"
	"log"
	"database/sql"
)

func hashSessions(configuration Configuration) {
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

	sessionsIns, err := db.Prepare("INSERT INTO sessions VALUES( ?, ?, ?, ?, ?, ? )")
	if err != nil {
		log.Println(err.Error())
	}
	defer sessionsIns.Close()

	var (
		logId int64
		logName string
		logTime time.Time
		logCount int64
		logStartId int64
		logStartTime time.Time
		logEndId int64
		logEndTime time.Time
		start time.Time
		end_time time.Time
		elapsed time.Duration
	)

	for {
		start = time.Now()

		err = sessionsOut.QueryRow().Scan(&end_time)
		if err != nil {
			log.Println(err.Error())
		}

		log.Println(end_time)

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

			if logName == "SessionCreated" {
				logStartId, logStartTime  = logId, logTime
				logEndId, logCount = 0, 0 // Reset
			} else if logName == "SessionDestroyed" {
				logEndId, logEndTime = logId, logTime
			}

			if logStartId > 0 && logEndId > 0 {
				_, err = sessionsIns.Exec(nil, logStartId, logEndId, logStartTime, logEndTime, logCount)
				if err != nil {
					log.Println(err.Error())
				}
				logStartId, logEndId, logCount = 0, 0, 0
			}
		}

		elapsed = time.Since(start)
		log.Printf("Hashing sessions took %s", elapsed)
		log.Println("Hasher idling for 1 hour")
		time.Sleep(time.Hour)
	}
}
