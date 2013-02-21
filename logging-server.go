package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"./loggingserver"
)

var port *int = flag.Int("port", 8080, "Listen on this port")
var mock *string = flag.String("mock", "", "Mock response configuration (JSON)")
var verbose *bool = flag.Bool("v", false, "Verbose mode (debug matching)")

func main() {
	flag.Parse()

	address := fmt.Sprintf(":%d", *port)

	loggingserver.DebugMatcher = *verbose

	server := new(loggingserver.LoggingServer)

	if *mock != "" {
		if buf, err := ioutil.ReadFile(*mock); err != nil {
			log.Fatalf("Could not open mock config %s: %v", *mock, err)
		} else {
			if err := json.Unmarshal(buf, &server.MockConfig); err != nil {
				log.Fatalf("Could not parse mock config %s: %v", *mock, err)
			}
		}
	}

	server.Init()

	log.Printf("Excluding %v", server.ExcludePaths)

	http.Handle("/", server)
	log.Printf("Listening on %s", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Could not listen on %s: %v", address, err)
	}
}
