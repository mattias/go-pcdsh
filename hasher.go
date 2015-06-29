package main
import (
	"time"
	"log"
	"database/sql"
)

func hashSessions(configuration Configuration) {
	for {
		start := time.Now()

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

		var end_time time.Time

		err = sessionsOut.QueryRow().Scan(&end_time)
		if err != nil {
			log.Println(err.Error())
		}

		log.Println(end_time)

		logsOut, err := db.Prepare("SELECT id, time, refid, participantid FROM `logs` WHERE `time` > ? ORDER BY `time` DESC, `index` DESC")
		if err != nil {
			log.Println(err.Error())
		}
		defer logsOut.Close()

		logAttributesOut, err := db.Prepare("SELECT `key`, value FROM `log_attributes` WHERE log_id = ?")
		if err != nil {
			log.Println(err.Error())
		}
		defer logAttributesOut.Close()


		sessionsIns, err := db.Prepare("INSERT INTO sessions VALUES( ?, ?, ?, ?, ?, ?, ? )")
		if err != nil {
			log.Println(err.Error())
		}
		defer sessionsIns.Close()


		var (
			sessionId int64
			sessionTime time.Time
			sessionRefid int64
			sessionParticipantid int64
			logAttributeKey string
			logAttributeValue string
		)

		logRows, err := logsOut.Query(end_time)
		if err != nil {
			log.Println(err.Error())
		}

		for logRows.Next() {
			err = logRows.Scan(&sessionId, &sessionTime, &sessionRefid, &sessionParticipantid)
			if err != nil {
				log.Println(err.Error())
			}

			log.Printf("%d, %v, %d, %d", sessionId, sessionTime, sessionRefid, sessionParticipantid)

			logAttributeRows, err := logAttributesOut.Query(sessionId)
			if err != nil {
				log.Println(err.Error())
			}

			for logAttributeRows.Next() {
				err = logAttributeRows.Scan(&logAttributeKey, &logAttributeValue)
				if err != nil {
					log.Println(err.Error())
				}

				log.Printf("%s: %s", logAttributeKey, logAttributeValue)
			}
		}

		elapsed := time.Since(start)
		log.Printf("Hasing sessions took %s", elapsed)
		log.Println("Hasher idling for 1 hour")
		time.Sleep(time.Hour)
	}
}
