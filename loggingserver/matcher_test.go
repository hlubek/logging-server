package loggingserver

import (
	"testing"

	"net/http"
)

type mockScenario struct {
	matcher       MockMatcher
	requestMethod string
	requestUrl    string
	expected      bool
}

var testScenarios = map[string]mockScenario{
	"exact path matched": mockScenario{
		MockMatcher{Method: "GET", Path: "/foo/bar"},
		"GET",
		"/foo/bar",
		true,
	},
	"exact path not matched with suffix": mockScenario{
		MockMatcher{Method: "GET", Path: "/foo/bar"},
		"GET",
		"/foo/bar/baz",
		false,
	},
	"exact path not matched with prefix": mockScenario{
		MockMatcher{Method: "GET", Path: "/foo/bar"},
		"GET",
		"/test/foo/bar",
		false,
	},
	"exact regexp path matched": mockScenario{
		MockMatcher{Method: "GET", Path: "^/foo$"},
		"GET",
		"/foo",
		true,
	},
	"regexp path matched with suffix": mockScenario{
		MockMatcher{Method: "GET", Path: "^/foo"},
		"GET",
		"/foo/bar/baz",
		true,
	},
	"regexp path matched with pattern": mockScenario{
		MockMatcher{Method: "GET", Path: "^/foo/(bar|baz)$"},
		"GET",
		"/foo/bar",
		true,
	},
	"regexp path not matched with wrong pattern": mockScenario{
		MockMatcher{Method: "GET", Path: "^/foo/(bar|baz)$"},
		"GET",
		"/foo/bad",
		false,
	},
	"exact query string matched": mockScenario{
		MockMatcher{Method: "GET", Path: "/", QueryString: "foo=bar"},
		"GET",
		"/?foo=bar",
		true,
	},
	"exact query string not matched": mockScenario{
		MockMatcher{Method: "GET", Path: "/", QueryString: "foo=bar"},
		"GET",
		"/?foo=wrong",
		false,
	},
}

func TestMatcherWithPath(t *testing.T) {
	for title, scenario := range testScenarios {
		r, _ := http.NewRequest(scenario.requestMethod, scenario.requestUrl, nil)

		if x := scenario.matcher.Matches(r); x != scenario.expected {
			t.Errorf("Scenario '%s' was expected to return '%v' for match, did return '%v'", title, scenario.expected, x)
		}
	}
}
