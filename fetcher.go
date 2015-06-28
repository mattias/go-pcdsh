package main

import (
	"log"
	"github.com/bndr/gopencils"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"strconv"
)

func fetchNewData(configuration Configuration) {
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

		api := gopencils.Api(configuration.BaseUrl)
		resp := new(RespStruct)
		logsOut, err := db.Prepare("SELECT `index` FROM `logs` ORDER BY `index` DESC LIMIT 1")
		if err != nil {
			log.Println(err.Error())
		}
		defer logsOut.Close()

		var index int64

		err = logsOut.QueryRow().Scan(&index)
		if err != nil {
			log.Println(err.Error())
		}
		log.Printf("Last index number was: %d", index)
		_, err = api.Res("log/overview", resp).Get()

		firstLog := resp.Response["first"].(float64)
		logCount := resp.Response["count"].(float64) + firstLog
		count := strconv.Itoa(int(int64(logCount) - index) + 10)
		var entries int

		log.Printf("Fetching %s log entries from server", count)

		querystring := map[string]string{"offset": "-" + count, "count": count}

		_, err = api.Res("log/range", resp).Get(querystring)

		if err != nil {
			log.Println(err)
		} else {
			logsIns, err := db.Prepare("INSERT INTO logs VALUES( ?, ?, ?, ?, ?, ? )")
			if err != nil {
				log.Println(err.Error())
			}
			defer logsIns.Close()

			logAttributesIns, err := db.Prepare("INSERT INTO log_attributes VALUES( ?, ?, ?, ? )")
			if err != nil {
				log.Println(err.Error())
			}
			defer logAttributesIns.Close()

			for _, event := range resp.Response["events"].([]interface{}) {
				event, _ := event.(map[string]interface{})

				if int64(event["index"].(float64)) <= index {
					continue
				}

				entries++

				var refid float64
				switch event["refid"].(type) {
					case float64:
					refid = event["refid"].(float64)
					default:
					refid = 0
				}

				var participantid float64
				switch event["participantid"].(type) {
					case float64:
					participantid = event["participantid"].(float64)
					default:
					participantid = 0
				}

				_, err := logsIns.Exec(
					nil,
					event["index"].(float64),
					time.Unix(int64(event["time"].(float64)), 0),
					event["name"].(string),
					refid,
					participantid,
				)
				if err != nil {
					log.Println(err.Error())
				}

				for key, value := range event["attributes"].(map[string]interface{}) {
					_, err = logAttributesIns.Exec(nil, event["index"].(float64), key, value)
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}
		log.Printf("Added %d new log entries out of %s fetched ones", entries, count)
		elapsed := time.Since(start)
		log.Printf("Fetching data took %s", elapsed)
		log.Println("Idling for 1 minute")
		time.Sleep(time.Minute)
	}
}
