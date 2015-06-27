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
	// TODO: Check local database and last index
	// TODO: Check api overview and last index
	// TODO: Calculate offset/count
	// TODO: Do padding of 10 logs? How many could be generated while we do this?

	logsOut, err := db.Prepare("SELECT `index` FROM `logs` ORDER BY `index` DESC LIMIT 1")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer logsOut.Close()

	var index int64

	err = logsOut.QueryRow().Scan(&index)
	if err != nil {
		fmt.Println(err.Error()) // proper error handling instead of panic in your app
	}
	fmt.Printf("The index number is: %d\n", index)
	_, err = api.Res("log/overview", resp).Get()

	firstLog := resp.Response["first"].(float64)
	logCount := resp.Response["count"].(float64) + firstLog
	count := strconv.Itoa(int(int64(logCount) - index))

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
			// TODO: Check if index is already added
			// TODO: continue if already added
			event, _ := event.(map[string]interface{})

			if int64(event["index"].(float64)) <= index {
				continue
			}

			result, err := logsIns.Exec(nil, event["index"].(float64), time.Unix(int64(event["time"].(float64)), 0), event["name"].(string), event["refid"].(float64), event["participantid"].(float64)) // Insert tuples (i, i^2)
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