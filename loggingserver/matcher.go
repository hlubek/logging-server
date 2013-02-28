package loggingserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

var DebugMatcher = false

type MockConfig []MockMatcher
type MockMatcher struct {
	id          int
	Method      string
	Path        string
	QueryString string
	Response    MockResponse
}
type MockResponse struct {
	StatusCode  int
	ContentType string
	Body        interface{}
	Defer       interface{}
}

func (m *MockMatcher) Matches(req *http.Request) bool {
	if m.Method != req.Method {
		if DebugMatcher {
			log.Printf("Method %s not matched", req.Method)
		}
		return false
	}

	if m.pathIsRegexp() {
		if matches, _ := regexp.MatchString(m.Path, req.URL.Path); !matches {
			if DebugMatcher {
				log.Printf("Path %s not matched with regexp %s", req.URL.Path, m.Path)
			}
			return false
		}
	} else {
		if m.Path != req.URL.Path {
			if DebugMatcher {
				log.Printf("Path %s not exactly matched with %s", req.URL.Path, m.Path)
			}
			return false
		}
	}

	if m.QueryString != "" {
		if m.queryStringIsRegexp() {
			if matches, _ := regexp.MatchString(m.QueryString, req.URL.RawQuery); !matches {
				if DebugMatcher {
					log.Printf("QueryString %s not matched with regexp %s", req.URL.RawQuery, m.QueryString)
				}
				return false
			}
		} else {
			if m.QueryString != req.URL.RawQuery {
				if DebugMatcher {
					log.Printf("QueryString %s not exactly matched with %s", req.URL.RawQuery, m.QueryString)
				}
				return false
			}
		}
	}

	return true
}

func (m *MockMatcher) Write(w http.ResponseWriter, req *http.Request) {
	// Defer processing of this request
	// TODO Parse duration after reading configuration
	var durValue string
	switch t := m.Response.Defer.(type) {
	case float64:
		durValue = fmt.Sprintf("%fs", t)
	case string:
		durValue = t
	}
	if durValue != "" {

		if dur, err := time.ParseDuration(durValue); err != nil {
			log.Printf("Error: Cannot parse duration \"%v\"", durValue)
		} else {
			<-time.After(dur)
		}
	}

	var contentType string
	if m.Response.ContentType != "" {
		contentType = m.Response.ContentType
	} else {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	if m.Response.StatusCode > 0 {
		w.WriteHeader(m.Response.StatusCode)
	}
	buf, _ := json.Marshal(m.Response.Body)
	w.Write(buf)
}

func (m *MockMatcher) queryStringIsRegexp() bool {
	return m.QueryString[0:1] == "^"
}

func (m *MockMatcher) pathIsRegexp() bool {
	return m.Path[0:1] == "^"
}
