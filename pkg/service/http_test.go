package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"github.com/rodb-io/rodb/pkg/input/record"
	outputPackage "github.com/rodb-io/rodb/pkg/output"
	"github.com/rodb-io/rodb/pkg/parser"
	"strings"
	"testing"
)

func TestHttp(t *testing.T) {
	config := &HttpConfig{
		Http: &HttpHttpConfig{
			Listen: ":0", // Auto-assign port
		},
		ErrorsType: "application/json",
		Logger:     logrus.NewEntry(logrus.StandardLogger()),
		Routes: []*HttpRouteConfig{
			{
				Path:   "/foo",
				Output: "mock",
			},
		},
	}
	parser := parser.NewMock()
	output := outputPackage.NewMock(parser)
	outputs := outputPackage.List{
		"mock": output,
	}
	server, err := NewHttp(config, outputs)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}
	defer server.Close()

	t.Run("normal", func(t *testing.T) {
		output.MockOutput = func(params map[string]string) ([]byte, error) {
			return []byte("Hello " + params["name"] + "!"), nil
		}

		response, err := http.Get(server.Address() + "/foo?name=Universe")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := http.StatusOK, response.StatusCode; got != expect {
			t.Fatalf("Expected status %+v, got '%+v'", expect, got)
		}

		body, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if got, expect := string(body), "Hello Universe!"; got != expect {
			t.Fatalf("Expected body '%+v', got '%+v'", expect, got)
		}
		if got, expect := response.Header.Get("Content-Type"), "text/plain"; !strings.HasPrefix(got, expect) {
			t.Fatalf("Expected Content-Type starting with '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("404", func(t *testing.T) {
		output.MockOutput = func(params map[string]string) ([]byte, error) {
			return nil, record.RecordNotFoundError
		}

		response, err := http.Get(server.Address() + "/foo")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := http.StatusNotFound, response.StatusCode; got != expect {
			t.Fatalf("Expected status '%+v', got '%+v'", expect, got)
		}
		if got, expect := response.Header.Get("Content-Type"), "application/json"; !strings.HasPrefix(got, expect) {
			t.Fatalf("Expected Content-Type starting with '%+v', got '%+v'", expect, got)
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		jsonBody := map[string]string{}
		if err := json.Unmarshal(body, &jsonBody); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		errorValue, errorValueExists := jsonBody["error"]
		if !errorValueExists {
			t.Fatalf("Expected to have an 'error' key, got '%+v'", errorValue)
		}
	})
}

func TestHttpOutputList(t *testing.T) {
	config := &HttpConfig{
		Http: &HttpHttpConfig{
			Listen: ":0", // Auto-assign port
		},
		ErrorsType: "application/json",
		Logger:     logrus.NewEntry(logrus.StandardLogger()),
		Routes: []*HttpRouteConfig{
			{
				Output: "foo",
			},
			{
				Output: "baz",
			},
		},
	}

	parser := parser.NewMock()
	outputFoo := outputPackage.NewMock(parser)
	outputBar := outputPackage.NewMock(parser)
	outputBaz := outputPackage.NewMock(parser)

	server, err := NewHttp(config, outputPackage.List{
		"foo": outputFoo,
		"bar": outputBar,
		"baz": outputBaz,
	})
	defer server.Close()
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	if expect, got := 2, len(server.routes); got != expect {
		t.Fatalf("Expected the server to hande %v routes, got %v", expect, got)
	}
	if expect, got := outputFoo, server.routes[0].output; got != expect {
		t.Fatalf("Expected the first route to be '%+v' routes, got '%+v'", expect, got)
	}
	if expect, got := outputBaz, server.routes[1].output; got != expect {
		t.Fatalf("Expected the first route to be '%+v' routes, got '%+v'", expect, got)
	}
}

func TestHttpGetMatchingRoute(t *testing.T) {
	payloadType := "application/json"
	parser := parser.NewMock()

	getFooOutput := outputPackage.NewMock(parser)
	getFooOutput.MockPayloadType = nil
	getFooRegexp, err := regexp.Compile("^/foo$")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	postBarOutput := outputPackage.NewMock(parser)
	postBarOutput.MockPayloadType = &payloadType
	postBarRegexp, err := regexp.Compile("^/bar$")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	getBarOutput := outputPackage.NewMock(parser)
	getBarOutput.MockPayloadType = nil
	getBarRegexp, err := regexp.Compile("^/bar$")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	server := &Http{
		routes: []*httpRoute{
			{
				path:   getFooRegexp,
				output: getFooOutput,
			},
			{
				path:   postBarRegexp,
				output: postBarOutput,
			},
			{
				path:   getBarRegexp,
				output: getBarOutput,
			},
		},
	}

	requestUrl, err := url.Parse("/bar")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}

	t.Run("get", func(t *testing.T) {
		expect := getBarOutput
		got := server.getMatchingRoute(&http.Request{
			Method: "GET",
			URL:    requestUrl,
		}).output
		if got != expect {
			t.Fatalf("Expected to get route '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("post", func(t *testing.T) {
		expect := postBarOutput
		requestHeader := http.Header(map[string][]string{})
		requestHeader.Set("Content-Type", payloadType)
		got := server.getMatchingRoute(&http.Request{
			Method: "POST",
			URL:    requestUrl,
			Header: requestHeader,
		}).output
		if got != expect {
			t.Fatalf("Expected to get route '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("wrong", func(t *testing.T) {
		var expect *httpRoute = nil
		requestHeader := http.Header(map[string][]string{})
		requestHeader.Set("Content-Type", "application/xml")
		got := server.getMatchingRoute(&http.Request{
			Method: "POST",
			URL:    requestUrl,
			Header: requestHeader,
		})
		if got != expect {
			t.Fatalf("Expected to get route '%+v', got '%+v'", expect, got)
		}
	})
}

func TestHttpGetParams(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		url, err := url.Parse("/foo/42/bar?id=wrong&foo=bar&baz=")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		regexp, err := regexp.Compile("/foo/(?P<id>[0-9]+)/bar")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		route := &httpRoute{
			path:       regexp,
			parameters: []string{"id"},
		}

		server := &Http{}
		params := server.getParams(route, url)

		if got := params["id"]; got != "42" {
			t.Fatalf("Expected param 'id' to be '42', got '%+v'", got)
		}
		if got := params["foo"]; got != "bar" {
			t.Fatalf("Expected param 'foo' to be 'bar', got '%+v'", got)
		}
		if got := params["baz"]; got != "" {
			t.Fatalf("Expected param 'baz' to be '', got '%+v'", got)
		}
	})
}

func TestHttpGetPayload(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		payloadType := "text/plain"
		parser := parser.NewMock()
		output := outputPackage.NewMock(parser)
		output.MockPayloadType = &payloadType

		server := &Http{}

		data := "Hello World!"
		body := strings.NewReader(data)

		payload, err := server.getPayload(&httpRoute{output: output}, body)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if string(payload) != data {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
	})
	t.Run("no expected payload", func(t *testing.T) {
		parser := parser.NewMock()
		output := outputPackage.NewMock(parser)
		output.MockPayloadType = nil

		server := &Http{}

		payload, err := server.getPayload(&httpRoute{output: output}, nil)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		if len(payload) > 0 {
			t.Fatalf("Unexpected payload: '%+v'. Expected nil", err)
		}
	})
}

func TestHttpCreatePathRegexp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		parser := parser.NewMock()
		output := outputPackage.NewMock(parser)
		server := &Http{}

		regexp, params, err := server.createPathRegexp(HttpRouteConfig{
			Path: "/foo/{foo_id}/bar-{bar}-id",
		}, output)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 2, len(params); got != expect {
			t.Fatalf("Expected to get %v parameters, got %+v", expect, got)
		}
		if expect, got := "foo_id", params[0]; got != expect {
			t.Fatalf("Expected to get '%v' as a first parameter, got '%+v'", expect, got)
		}
		if expect, got := "bar", params[1]; got != expect {
			t.Fatalf("Expected to get '%+v' as a second parameter, got '%+v'", expect, got)
		}

		testPath := "/foo/42/bar-test-id"
		if !regexp.MatchString(testPath) {
			t.Fatalf("Expected the regexp to match the path '%+v'", testPath)
		}
		matches := regexp.FindStringSubmatch(testPath)
		if expect, got := 3, len(matches); got != expect {
			t.Fatalf("Expected to get %v match results, got %+v", expect, got)
		}
		if expect, got := "42", matches[1]; got != expect {
			t.Fatalf("Expected to get '%v' as a second match, got '%+v'", expect, got)
		}
		if expect, got := "test", matches[2]; got != expect {
			t.Fatalf("Expected to get '%+v' as a third match, got '%+v'", expect, got)
		}
	})
	t.Run("no params", func(t *testing.T) {
		parser := parser.NewMock()
		output := outputPackage.NewMock(parser)
		server := &Http{}

		regexp, params, err := server.createPathRegexp(HttpRouteConfig{
			Path: "/foo",
		}, output)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if expect, got := 0, len(params); got != expect {
			t.Fatalf("Expected to get %v parameters, got %+v", expect, got)
		}

		testPath := "/foo"
		if !regexp.MatchString(testPath) {
			t.Fatalf("Expected the regexp to match the path '%+v'", testPath)
		}
	})
}
