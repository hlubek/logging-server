package main

import (
	"flag"
	"log"
	"http"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"json"
	"regexp"

	"github.com/wsxiaoys/colors"
)

var port *int = flag.Int("port", 8080, "Listen on this port")
var mock *string = flag.String("mock", "", "Mock response configuration (JSON)")
var mockConfig *MockConfig

type MockConfig []MockMatcher
type MockMatcher struct {
	Method   string
	Path     string
	Response MockResponse
}
type MockResponse struct {
	ContentType string
	Body        interface{}
}

func (m *MockMatcher) Matches(req *http.Request) bool {
	if m.Method == req.Method {
		if m.Path == req.URL.Path {
			return true
		}
		if matches, _ := regexp.MatchString(m.Path, req.URL.Path); matches {
			return true
		}
	}
	return false
}

func (m *MockMatcher) Write(w http.ResponseWriter, req *http.Request) {
	var contentType string
	if m.Response.ContentType != "" {
		contentType = m.Response.ContentType
	} else {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	buf, _ := json.Marshal(m.Response.Body)
	w.Write(buf)
}

var excludePaths []string = []string{
	"/favicon.ico",
}

func LogRequest(w http.ResponseWriter, req *http.Request) {
	if i := sort.SearchStrings(excludePaths, req.URL.Path); i < len(excludePaths) && req.URL.Path == excludePaths[i] {
		return
	}

	req.ParseForm()

	// Log request
	var data string
	for key, _ := range req.Form {
		data += "\n    "
		if value := req.Form.Get(key); value != "" {
			data += colors.Sprintf("@.%s@|@w = @|%s", key, value)
		} else {
			data += colors.Sprintf("%s", key)
		}
	}

	log.Print(colors.Sprintf("@c%s @{!y}%s", req.Method, req.URL.RawPath))
	if data != "" {
		colors.Printf("  @bRequest@|%s\n", data)
	}

	// Find matcher
	matcher := findMatcher(req)
	if matcher != nil {
		matcher.Write(w, req)
		// TODO Better response logging
		response := matcher.Response

		colors.Printf("  @gResponse@|\n    %+v\n", response)
	} else {
		colors.Print("  @rUnmatched@|\n")

		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html><body>Unmatched request, please configure a mock request for path " + req.URL.RawPath + " and method " + req.Method + "</body></html>\n")
	}
}

func findMatcher(req *http.Request) *MockMatcher {
	for _, matcher := range *mockConfig {
		if matcher.Matches(req) {
			return &matcher
		}
	}
	return nil
}

func main() {
	flag.Parse()

	address := fmt.Sprintf(":%d", *port)

	mockConfig = new(MockConfig)

	if *mock != "" {
		if buf, err := ioutil.ReadFile(*mock); err != nil {
			log.Fatalf("Could not open mock config %s: %v", *mock, err)
		} else {
			if err := json.Unmarshal(buf, mockConfig); err != nil {
				log.Fatalf("Could not parse mock config %s: %v", *mock, err)
			}
		}
	}

	sort.Sort(sort.StringSlice(excludePaths))

	log.Printf("Excluding %v", excludePaths)

	http.HandleFunc("/", LogRequest)
	log.Printf("Listening on %s", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Could not listen on %s: %v", address, err)
	}
}
