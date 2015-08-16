package main

import (
	"encoding/json"
	"os"
	"log"
	"fmt"
	"github.com/emicklei/go-restful"
	"net/http"
)

func main() {
	configuration := readConfiguration()
	go fetchNewData(configuration)
	go hashSessions(configuration)

	wsContainer := restful.NewContainer()
	s := SessionResource{}
	s.RegisterTo(wsContainer)
	l := LogResource{}
	l.RegisterTo(wsContainer)


	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)

	wsContainer.Filter(wsContainer.OPTIONSFilter)

	log.Printf("Start listening on localhost:" + configuration.Port)
	server := &http.Server{Addr: configuration.Port, Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
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
