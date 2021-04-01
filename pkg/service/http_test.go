package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"rods/pkg/config"
	outputModule "rods/pkg/output"
	"rods/pkg/parser"
	"rods/pkg/record"
	"strings"
	"testing"
)

func TestHttp(t *testing.T) {
	config := &config.HttpService{
		Listen:     ":0", // Auto-assign port
		ErrorsType: "application/json",
		Logger:     logrus.NewEntry(logrus.StandardLogger()),
		Routes: []config.HttpServiceRoute{
			{
				Path:   "/foo",
				Output: "mock",
			},
		},
	}
	parser := parser.NewMock()
	output := outputModule.NewMock(parser)
	outputs := outputModule.List{
		"mock": output,
	}
	server, err := NewHttp(config, outputs)
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	defer server.Close()

	t.Run("normal", func(t *testing.T) {
		output.MockOutput = func(params map[string]string) ([]byte, error) {
			return []byte("Hello " + params["name"] + "!"), nil
		}

		response, err := http.Get(server.Address() + "/foo?name=Universe")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := http.StatusOK, response.StatusCode; got != expect {
			t.Errorf("Expected status %+v, got '%+v'", expect, got)
		}

		body, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if got, expect := string(body), "Hello Universe!"; got != expect {
			t.Errorf("Expected body '%+v', got '%+v'", expect, got)
		}
		if got, expect := response.Header.Get("Content-Type"), "text/plain"; !strings.HasPrefix(got, expect) {
			t.Errorf("Expected Content-Type starting with '%+v', got '%+v'", expect, got)
		}
	})
	t.Run("404", func(t *testing.T) {
		output.MockOutput = func(params map[string]string) ([]byte, error) {
			return nil, record.RecordNotFoundError
		}

		response, err := http.Get(server.Address() + "/foo")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if expect, got := http.StatusNotFound, response.StatusCode; got != expect {
			t.Errorf("Expected status '%+v', got '%+v'", expect, got)
		}
		if got, expect := response.Header.Get("Content-Type"), "application/json"; !strings.HasPrefix(got, expect) {
			t.Errorf("Expected Content-Type starting with '%+v', got '%+v'", expect, got)
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		jsonBody := map[string]string{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		errorValue, errorValueExists := jsonBody["error"]
		if !errorValueExists {
			t.Errorf("Expected to have an 'error' key, got '%+v'", errorValue)
		}
	})
}

func TestHttpOutputList(t *testing.T) {
	config := &config.HttpService{
		Listen:     ":0", // Auto-assign port
		ErrorsType: "application/json",
		Logger:     logrus.NewEntry(logrus.StandardLogger()),
		Routes: []config.HttpServiceRoute{
			{
				Output: "foo",
			},
			{
				Output: "baz",
			},
		},
	}

	parser := parser.NewMock()
	outputFoo := outputModule.NewMock(parser)
	outputBar := outputModule.NewMock(parser)
	outputBaz := outputModule.NewMock(parser)

	server, err := NewHttp(config, outputModule.List{
		"foo": outputFoo,
		"bar": outputBar,
		"baz": outputBaz,
	})
	defer server.Close()
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	if expect, got := 2, len(server.routes); got != expect {
		t.Errorf("Expected the server to hande %v routes, got %v", expect, got)
	}
	if expect, got := outputFoo, server.routes[0].output; got != expect {
		t.Errorf("Expected the first route to be '%+v' routes, got '%+v'", expect, got)
	}
	if expect, got := outputBaz, server.routes[1].output; got != expect {
		t.Errorf("Expected the first route to be '%+v' routes, got '%+v'", expect, got)
	}
}

func TestHttpGetMatchingRoute(t *testing.T) {
	payloadType := "application/json"
	parser := parser.NewMock()

	getFooOutput := outputModule.NewMock(parser)
	getFooOutput.MockPayloadType = nil

	postBarOutput := outputModule.NewMock(parser)
	postBarOutput.MockPayloadType = &payloadType

	getBarOutput := outputModule.NewMock(parser)
	getBarOutput.MockPayloadType = nil

	server := &Http{
		routes: []*httpRoute{
			{
				path:   regexp.MustCompile("^/foo$"),
				output: getFooOutput,
			},
			{
				path:   regexp.MustCompile("^/bar$"),
				output: postBarOutput,
			},
			{
				path:   regexp.MustCompile("^/bar$"),
				output: getBarOutput,
			},
		},
	}

	requestUrl, err := url.Parse("/bar")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	t.Run("get", func(t *testing.T) {
		expect := getBarOutput
		got := server.getMatchingRoute(&http.Request{
			Method: "GET",
			URL:    requestUrl,
		}).output
		if got != expect {
			t.Errorf("Expected to get route '%+v', got '%+v'", expect, got)
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
			t.Errorf("Expected to get route '%+v', got '%+v'", expect, got)
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
			t.Errorf("Expected to get route '%+v', got '%+v'", expect, got)
		}
	})
}

func TestHttpGetParams(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		url, err := url.Parse("/foo/42/bar?id=wrong&foo=bar&baz=")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		route := &httpRoute{
			path:       regexp.MustCompile("/foo/(?P<id>[0-9]+)/bar"),
			parameters: []string{"id"},
		}

		server := &Http{}
		params := server.getParams(route, url)

		if got := params["id"]; got != "42" {
			t.Errorf("Expected param 'id' to be '42', got '%+v'", got)
		}
		if got := params["foo"]; got != "bar" {
			t.Errorf("Expected param 'foo' to be 'bar', got '%+v'", got)
		}
		if got := params["baz"]; got != "" {
			t.Errorf("Expected param 'baz' to be '', got '%+v'", got)
		}
	})
}

func TestHttpGetPayload(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		payloadType := "text/plain"
		parser := parser.NewMock()
		output := outputModule.NewMock(parser)
		output.MockPayloadType = &payloadType

		server := &Http{}

		data := "Hello World!"
		body := strings.NewReader(data)

		payload, err := server.getPayload(&httpRoute{output: output}, body)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if string(payload) != data {
			t.Errorf("Unexpected error: '%+v'", err)
		}
	})
	t.Run("no expected payload", func(t *testing.T) {
		parser := parser.NewMock()
		output := outputModule.NewMock(parser)
		output.MockPayloadType = nil

		server := &Http{}

		payload, err := server.getPayload(&httpRoute{output: output}, nil)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if len(payload) > 0 {
			t.Errorf("Unexpected payload: '%+v'. Expected nil", err)
		}
	})
}
