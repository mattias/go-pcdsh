package main

import (
	"encoding/json"
	"time"
	"os"
	"fmt"
	"github.com/bndr/gopencils"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
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
	// TODO: Make the below into a func and run in goroutine
	// TODO: Make the goroutine execute every few minutes or something
	api := gopencils.Api(configuration.BaseUrl)
	resp := new(RespStruct)
	querystring := map[string]string{"offset": "-100", "count": "100"}

	_, err := api.Res("log/range", resp).Get(querystring)

	if err != nil {
		fmt.Println(err)
	} else {
		// TODO: Rewrite into the proper logic, so far just testing
		db, err := sql.Open("mysql", configuration.Datasource)
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()

		// Open doesn't open a connection. Validate DSN data:
		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}

		logsIns, err := db.Prepare("INSERT INTO logs VALUES( ?, ?, ?, ?, ? )") // ? = placeholder
		if err != nil {
			fmt.Println(err.Error()) // proper error handling instead of panic in your app
		}
		defer logsIns.Close() // Close the statement when we leave main() / the program terminates

		// Prepare statement for inserting data
		logAttributesIns, err := db.Prepare("INSERT INTO log_attributes VALUES( ?, ?, ?, ? )") // ? = placeholder
		if err != nil {
			fmt.Println(err.Error()) // proper error handling instead of panic in your app
		}
		defer logAttributesIns.Close() // Close the statement when we leave main() / the program terminates

		// Insert square numbers for 0-24 in the database
		for i := 0; i < 25; i++ {
			result, err := logsIns.Exec(nil, time.Now(), (i * i), (i * i * i), (i * i * i * i)) // Insert tuples (i, i^2)
			if err != nil {
				fmt.Println(err.Error()) // proper error handling instead of panic in your app
			}
			insertId, _ := result.LastInsertId()
			fmt.Println(insertId)

			result, err = logAttributesIns.Exec(nil, i+1, (i * i), (i * i * i)) // Insert tuples (i, i^2)
			if err != nil {
				fmt.Println(err.Error()) // proper error handling instead of panic in your app
			}

			insertId, _ = result.LastInsertId()
			fmt.Println(insertId)
		}

		for _, event := range resp.Response["events"].([]interface{}) {
			fmt.Println(event) // TODO: save to database
		}
	}
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