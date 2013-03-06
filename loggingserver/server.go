package loggingserver

import (
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"encoding/json"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
)

type LoggingServer struct {
	ExcludePaths []string
	MockConfig
	mu    sync.RWMutex // Guard the fields for refresh
	mtime time.Time    // Last modified of configuration
	quit  chan bool
}

var defaultExcludePaths []string = []string{
	"/favicon.ico",
}

func (s *LoggingServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if i := sort.SearchStrings(s.ExcludePaths, req.URL.Path); i < len(s.ExcludePaths) && req.URL.Path == s.ExcludePaths[i] {
		return
	}

	req.ParseForm()

	// Log request
	var data string
	for key, _ := range req.Form {
		data += "\n    "
		if value := req.Form.Get(key); value != "" {
			data += color.Sprintf("@.%s@|@w = @|%s", key, value)
		} else {
			data += color.Sprintf("%s", key)
		}
	}

	log.Print(color.Sprintf("@c%s @{!y}%s", req.Method, req.RequestURI))
	if data != "" {
		color.Printf("  @bRequest@|%s\n", data)
	}

	// Find matcher
	matcher := s.findMatcher(req)
	if matcher != nil {
		matcher.Write(w, req)
		// TODO Better response logging
		response := matcher.Response

		color.Printf("  @gResponse@|\n    %+v\n", response)
	} else {
		color.Print("  @rUnmatched@|\n")

		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html><body>Unmatched request, please configure a mock request for path "+req.URL.Path+" and method "+req.Method+"</body></html>\n")
	}
}

func (s *LoggingServer) findMatcher(req *http.Request) *MockMatcher {
	for _, matcher := range s.MockConfig {
		if matcher.Matches(req) {
			return &matcher
		}
	}
	return nil
}

func (s *LoggingServer) loadMockConfig(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	mtime := fi.ModTime()
	if mtime.Before(s.mtime) && s.MockConfig != nil {
		return nil // No change
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	config := MockConfig{}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return fmt.Errorf("parsing %s: %v", filePath, err)
	}
	s.mu.Lock()
	s.mtime = mtime
	s.MockConfig = config
	s.mu.Unlock()

	return nil
}

func (s *LoggingServer) refreshMockConfig(filePath string, poll time.Duration) {
	t := time.NewTicker(poll)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := s.loadMockConfig(filePath); err != nil {
				// Ignore errors
			}
		case <-s.quit:
			return
		}
	}
}

func NewLoggingServer(mockConfiguration string, poll time.Duration) (*LoggingServer, error) {
	s := &LoggingServer{
		ExcludePaths: defaultExcludePaths,
		quit:         make(chan bool),
	}
	sort.Sort(sort.StringSlice(s.ExcludePaths))
	if mockConfiguration != "" {
		if err := s.loadMockConfig(mockConfiguration); err != nil {
			return nil, err
		}
		if poll > 0 {
			go s.refreshMockConfig(mockConfiguration, poll)
		}
	} else {
		s.MockConfig = MockConfig{}
	}
	return s, nil
}

func (s *LoggingServer) Stop() {
	close(s.quit)
}
