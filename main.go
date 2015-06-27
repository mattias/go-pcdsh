package main

import (
	"encoding/json"
	"os"
	"fmt"
	"github.com/bndr/gopencils"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"strconv"
)

type Configuration struct {
	BaseUrl    string
	Datasource string
}

type RespStruct struct {
	Result   string
	Response map[string]interface{}
}

func main() {
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

	api := gopencils.Api(configuration.BaseUrl)
	resp := new(RespStruct)
	logsOut, err := db.Prepare("SELECT `index` FROM `logs` ORDER BY `index` DESC LIMIT 1")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer logsOut.Close()

	var index int64

	err = logsOut.QueryRow().Scan(&index)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("The index number is: %d\n", index)
	_, err = api.Res("log/overview", resp).Get()

	firstLog := resp.Response["first"].(float64)
	logCount := resp.Response["count"].(float64) + firstLog
	count := strconv.Itoa(int(int64(logCount) - index) + 10)

	querystring := map[string]string{"offset": "-" + count, "count": count}

	_, err = api.Res("log/range", resp).Get(querystring)

	if err != nil {
		fmt.Println(err)
	} else {
		logsIns, err := db.Prepare("INSERT INTO logs VALUES( ?, ?, ?, ?, ?, ? )")
		if err != nil {
			fmt.Println(err.Error())
		}
		defer logsIns.Close()

		logAttributesIns, err := db.Prepare("INSERT INTO log_attributes VALUES( ?, ?, ?, ? )")
		if err != nil {
			fmt.Println(err.Error())
		}
		defer logAttributesIns.Close()

		for _, event := range resp.Response["events"].([]interface{}) {
			event, _ := event.(map[string]interface{})

			if int64(event["index"].(float64)) <= index {
				continue
			}

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

			result, err := logsIns.Exec(nil, event["index"].(float64), time.Unix(int64(event["time"].(float64)), 0), event["name"].(string), refid, participantid)
			if err != nil {
				fmt.Println(err.Error())
			}
			insertId, _ := result.LastInsertId()

			for key, value := range event["attributes"].(map[string]interface{}) {
				result, err = logAttributesIns.Exec(nil, insertId, key, value)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
		fmt.Println("New values inserted to database")
	}
	fmt.Println("//End")
}

func readConfiguration() (Configuration) {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}