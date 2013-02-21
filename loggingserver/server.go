package loggingserver

import (
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/wsxiaoys/colors"
)

type LoggingServer struct {
	ExcludePaths []string
	MockConfig
}

var defaultExcludePaths []string = []string{
	"/favicon.ico",
}

func (s *LoggingServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if i := sort.SearchStrings(s.ExcludePaths, req.URL.Path); i < len(s.ExcludePaths) && req.URL.Path == s.ExcludePaths[i] {
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

	log.Print(colors.Sprintf("@c%s @{!y}%s", req.Method, req.RequestURI))
	if data != "" {
		colors.Printf("  @bRequest@|%s\n", data)
	}

	// Find matcher
	matcher := s.findMatcher(req)
	if matcher != nil {
		matcher.Write(w, req)
		// TODO Better response logging
		response := matcher.Response

		colors.Printf("  @gResponse@|\n    %+v\n", response)
	} else {
		colors.Print("  @rUnmatched@|\n")

		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html><body>Unmatched request, please configure a mock request for path "+req.URL.Path+" and method "+req.Method+"</body></html>\n")
	}
}

func (s *LoggingServer) Init() {
	s.ExcludePaths = defaultExcludePaths
	sort.Sort(sort.StringSlice(s.ExcludePaths))
}

func (s *LoggingServer) findMatcher(req *http.Request) *MockMatcher {
	for _, matcher := range s.MockConfig {
		// log.Printf("Try matcher %v", i)
		if matcher.Matches(req) {
			return &matcher
		}
	}
	return nil
}
