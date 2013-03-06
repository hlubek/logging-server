package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chlu/logging-server/loggingserver"
)

var (
	port        *int    = flag.Int("port", 8080, "Listen on this port")
	bindAddress *string = flag.String("address", "127.0.0.1", "Bind to this address")
	watch       *bool   = flag.Bool("watch", false, "Watch config file for changes")
	debug       *bool   = flag.Bool("debug", false, "Debug mode (debug matching)")
)

const (
	pollForChanges = time.Second * 3
)

func main() {
	var mockConfiguration string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options] [mock configuration]\n\nArguments:\n  mock configuration: A JSON file with rules for mock responses\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() > 0 {
		mockConfiguration = flag.Arg(0)
	}

	address := fmt.Sprintf("%s:%d", *bindAddress, *port)
	loggingserver.DebugMatcher = *debug

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	server, err := loggingserver.NewLoggingServer(mockConfiguration, pollForChanges)
	if err != nil {
		log.Fatalf("Could not initialize server with config \"%s\": %v", mockConfiguration, err)
	}

	log.Printf("Excluding paths %v", server.ExcludePaths)

	http.Handle("/", server)

	log.Printf("Listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Could not listen on %s: %v", address, err)
	}
}
