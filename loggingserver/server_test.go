package loggingserver

import (
	"testing"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func TestNewLoggingServerWithNoConfiguration(t *testing.T) {
	s, err := NewLoggingServer("", 0)
	if err != nil {
		t.Fatalf("Error initializing server: %v", err)
		return
	}
	defer s.Stop()

	if s.MockConfig == nil {
		t.Errorf("Server config should not be nil")
	}
}

func TestNewLoggingServerWithConfigurationNoPoll(t *testing.T) {
	testConfig := createTempFile(t)
	defer os.Remove(testConfig)

	writeMockConfig(testConfig, t, MockConfig{
		{
			Method: "GET",
			Path:   "/",
		},
	})

	s, err := NewLoggingServer(testConfig, 0)
	if err != nil {
		t.Fatalf("Error initializing server: %v", err)
		return
	}
	defer s.Stop()

	if s.MockConfig == nil {
		t.Fatalf("Server config should not be nil")
	}
	if len(s.MockConfig) != 1 {
		t.Errorf("Mock config should have one entry, got %d", len(s.MockConfig))
	}
}

func TestNewLoggingServerWithConfigurationPoll(t *testing.T) {
	testConfig := createTempFile(t)
	defer os.Remove(testConfig)

	writeMockConfig(testConfig, t, MockConfig{})

	s, err := NewLoggingServer(testConfig, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Error initializing server: %v", err)
		return
	}
	defer s.Stop()

	if s.MockConfig == nil {
		t.Fatalf("Server config should not be nil")
	}
	if len(s.MockConfig) != 0 {
		t.Errorf("Mock config should have no entries, got %d", len(s.MockConfig))
	}

	writeMockConfig(testConfig, t, MockConfig{
		{
			Method: "GET",
			Path:   "/",
		},
	})

	time.Sleep(100 * time.Millisecond)

	if len(s.MockConfig) != 1 {
		t.Errorf("Mock config should have one entry after reload, got %d", len(s.MockConfig))
	}
}

func TestServerHandler(t *testing.T) {
	testConfig := createTempFile(t)
	defer os.Remove(testConfig)

	writeMockConfig(testConfig, t, MockConfig{
		MockMatcher{
			Method: "GET",
			Path:   "/foo",
		},
	})

	s, err := NewLoggingServer(testConfig, 0)
	if err != nil {
		t.Fatalf("Error initializing server: %v", err)
		return
	}
	defer s.Stop()

	var tests = []struct {
		method string
		path   string
		code   int
		body   string
	}{
		{"GET", "/foo", 200, ""},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		rw := httptest.NewRecorder()
		rw.Body = new(bytes.Buffer)
		s.ServeHTTP(rw, req)

		if g, w := rw.Code, test.code; g != w {
			t.Errorf("code = %d, want %d", g, w)
		}
		if g, w := rw.Body.String(), test.body; g != w {
			t.Errorf("body = %q, want %q", g, w)
		}
	}

}

func writeMockConfig(filePath string, t *testing.T, config MockConfig) {
	if data, err := json.Marshal(&config); err != nil {
		t.Fatalf("Could not encode config: %v", err)
	} else {
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			t.Fatalf("Could not write config: %v", err)
		}
	}
}

func createTempFile(t *testing.T) string {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	defer tmpFile.Close()
	return tmpFile.Name()
}
