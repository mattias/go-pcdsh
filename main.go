package main

import (
	"encoding/json"
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
	api := gopencils.Api(configuration.BaseUrl)
	resp := new(RespStruct)
	querystring := map[string]string{"offset": "-100", "count": "100"}

	_, err := api.Res("log/range", resp).Get(querystring)

	if err != nil {
		fmt.Println(err)
	} else {
		db, err := sql.Open("mysql", configuration.Datasource)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer db.Close()

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