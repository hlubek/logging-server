package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/howeyc/fsnotify"

	"github.com/chlu/logging-server/loggingserver"
)

var (
	port        *int    = flag.Int("port", 8080, "Listen on this port")
	bindAddress *string = flag.String("address", "127.0.0.1", "Bind to this address")
	watch       *bool   = flag.Bool("watch", false, "Watch config file for changes")
	debug       *bool   = flag.Bool("debug", false, "Debug mode (debug matching)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options] [mock configuration]\n\nArguments:\n  mock configuration: A JSON file with rules for mock responses\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	address := fmt.Sprintf("%s:%d", *bindAddress, *port)
	loggingserver.DebugMatcher = *debug

	server := new(loggingserver.LoggingServer)

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if flag.NArg() > 0 {
		mockConfiguration := flag.Arg(0)
		log.Printf("Using mock configuration \"%s\"", mockConfiguration)

		if *watch {
			if err := readAndWatchConfig(mockConfiguration, server); err != nil {
				log.Fatalf("Could not read mock config \"%s\": %v", mockConfiguration, err)
			}
			log.Printf("Watching for changes on \"%s\"", mockConfiguration)
		} else {
			if err := readConfig(mockConfiguration, server); err != nil {
				log.Fatalf("Could not read mock config \"%s\": %v", mockConfiguration, err)
			}
		}

	}

	server.Init()

	log.Printf("Excluding paths %v", server.ExcludePaths)

	http.Handle("/", server)

	log.Printf("Listening on %s", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Could not listen on %s: %v", address, err)
	}
}

func readAndWatchConfig(filePath string, server *loggingserver.LoggingServer) error {
	err := readConfig(filePath, server)
	if err != nil {
		return err
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	err = w.Watch(filePath)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case e := <-w.Event:
				if e.IsModify() {
					log.Printf("Reloading config file \"%s\"", filePath)
					if err := readConfig(filePath, server); err != nil {
						log.Printf("ERROR: Could not reload mock config \"%s\": %v", filePath, err)
					}
				}
			case err := <-w.Error:
				log.Println("error:", err)
			}
		}
	}()
	return nil
}

func readConfig(filePath string, server *loggingserver.LoggingServer) error {
	if buf, err := ioutil.ReadFile(filePath); err != nil {
		return err
	} else {
		if err := json.Unmarshal(buf, &server.MockConfig); err != nil {
			return err
		}
	}
	return nil
}
