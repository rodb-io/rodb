package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"rods/pkg/config"
	"strings"
	"sync"
	"testing"
)

func TestHttp(t *testing.T) {
	config := &config.HttpService{
		Listen:     ":0", // Auto-assign port
		ErrorsType: "application/json",
		Logger:     logrus.NewEntry(logrus.StandardLogger()),
	}
	server, err := NewHttp(config)
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	defer server.Close()

	route := &Route{
		Endpoint:            regexp.MustCompile("/foo"),
		ExpectedPayloadType: nil,
		ResponseType:        "text/plain",
		Handler: func(
			params map[string]string,
			payload []byte,
			sendError func(err error) error,
			sendSucces func() io.Writer,
		) error {
			_, err := sendSucces().Write([]byte("Hello " + params["name"] + "!"))
			return err
		},
	}

	server.AddRoute(route)

	t.Run("normal", func(t *testing.T) {
		route.Handler = func(
			params map[string]string,
			payload []byte,
			sendError func(err error) error,
			sendSucces func() io.Writer,
		) error {
			_, err := sendSucces().Write([]byte("Hello " + params["name"] + "!"))
			return err
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
		route.Handler = func(
			params map[string]string,
			payload []byte,
			sendError func(err error) error,
			sendSucces func() io.Writer,
		) error {
			return sendError(RecordNotFoundError)
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

func TestHttpAddRoute(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		server := &Http{
			routes:     make([]*Route, 0),
			routesLock: &sync.Mutex{},
		}

		route := &Route{ResponseType: "application/test"}
		server.AddRoute(route)

		if got, expect := len(server.routes), 1; got != expect {
			t.Errorf("Expected the server to contain '%v' routes, got '%+v'", expect, got)
		} else if server.routes[0] != route {
			t.Errorf("Expected the server routes to contain '%+v', got '%+v'", route, server.routes)
		}
	})
}

func TestHttpDeleteRoute(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		route := &Route{ResponseType: "application/test"}
		server := &Http{
			routes:     make([]*Route, 0),
			routesLock: &sync.Mutex{},
		}
		server.routes = append(server.routes, route)

		server.DeleteRoute(route)
		if got, expect := len(server.routes), 0; got != expect {
			t.Errorf("Expected the server to contain '%v' routes, got '%+v'", expect, got)
		}
	})
}

func TestHttpGetMatchingRoute(t *testing.T) {
	payloadType := "application/json"
	getFooRoute := &Route{
		Endpoint:            regexp.MustCompile("/foo"),
		ExpectedPayloadType: nil,
	}
	postBarRoute := &Route{
		Endpoint:            regexp.MustCompile("/bar"),
		ExpectedPayloadType: &payloadType,
	}
	getBarRoute := &Route{
		Endpoint:            regexp.MustCompile("/bar"),
		ExpectedPayloadType: nil,
	}
	server := &Http{
		routes: []*Route{
			getFooRoute,
			postBarRoute,
			getBarRoute,
		},
		routesLock: &sync.Mutex{},
	}

	requestUrl, err := url.Parse("/bar")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}

	t.Run("get", func(t *testing.T) {
		except := getBarRoute
		got := server.getMatchingRoute(&http.Request{
			Method: "GET",
			URL:    requestUrl,
		})
		if got != except {
			t.Errorf("Expected to get route '%+v', got '%+v'", except, got)
		}
	})
	t.Run("post", func(t *testing.T) {
		except := postBarRoute
		requestHeader := http.Header(map[string][]string{})
		requestHeader.Set("Content-Type", payloadType)
		got := server.getMatchingRoute(&http.Request{
			Method: "POST",
			URL:    requestUrl,
			Header: requestHeader,
		})
		if got != except {
			t.Errorf("Expected to get route '%+v', got '%+v'", except, got)
		}
	})
	t.Run("wrong", func(t *testing.T) {
		var except *Route = nil
		requestHeader := http.Header(map[string][]string{})
		requestHeader.Set("Content-Type", "application/xml")
		got := server.getMatchingRoute(&http.Request{
			Method: "POST",
			URL:    requestUrl,
			Header: requestHeader,
		})
		if got != except {
			t.Errorf("Expected to get route '%+v', got '%+v'", except, got)
		}
	})
}

func TestHttpGetParams(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		url, err := url.Parse("/foo/42/bar?id=wrong&foo=bar&baz=")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		endpoint := regexp.MustCompile("/foo/(?P<id>[0-9]+)/bar")

		server := &Http{}
		params := server.getParams(endpoint, url)

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
		server := &Http{}
		payloadType := "text/plain"
		route := &Route{
			ExpectedPayloadType: &payloadType,
		}

		data := "Hello World!"
		body := strings.NewReader(data)

		payload, err := server.getPayload(route, body)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if string(payload) != data {
			t.Errorf("Unexpected error: '%+v'", err)
		}
	})
	t.Run("no expected payload", func(t *testing.T) {
		server := &Http{}
		route := &Route{
			ExpectedPayloadType: nil,
		}

		payload, err := server.getPayload(route, nil)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		if len(payload) > 0 {
			t.Errorf("Unexpected payload: '%+v'. Expected nil", err)
		}
	})
}
