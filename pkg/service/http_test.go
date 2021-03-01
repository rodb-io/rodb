package service

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"rods/pkg/config"
	"rods/pkg/util"
	"strings"
	"testing"
)

func TestHttp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		config := &config.HttpService{Port: 0} // Auto-assign port
		server, err := NewHttp(config, logrus.StandardLogger())
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer server.Close()

		server.AddRoute(&Route{
			Endpoint:            regexp.MustCompile("/foo"),
			ExpectedPayloadType: nil,
			ResponseType:        "text/plain",
			Handler: func(
				params map[string]string,
				payload []byte,
			) ([]byte, error) {
				return []byte("Hello " + params["name"] + "!"), nil
			},
		})

		response, err := http.Get(server.Address() + "/foo?name=Universe")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
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
}

func TestHttpAddRoute(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		server := &Http{
			routes: make([]*Route, 0),
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
			routes: make([]*Route, 0),
		}
		server.routes = append(server.routes, route)

		server.DeleteRoute(route)
		if got, expect := len(server.routes), 0; got != expect {
			t.Errorf("Expected the server to contain '%v' routes, got '%+v'", expect, got)
		}
	})
}

func TestHttpGetMatchingRoute(t *testing.T) {
	getFooRoute := &Route{
		Endpoint:            regexp.MustCompile("/foo"),
		ExpectedPayloadType: nil,
	}
	postBarRoute := &Route{
		Endpoint:            regexp.MustCompile("/bar"),
		ExpectedPayloadType: util.PString("application/json"),
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
		requestHeader.Set("Content-Type", "application/json")
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
		route := &Route{
			ExpectedPayloadType: util.PString("text/plain"),
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
