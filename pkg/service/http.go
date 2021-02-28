package service

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	goLog "log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"rods/pkg/config"
	"strconv"
	"sync"
)

type Http struct {
	listener  net.Listener
	server    *http.Server
	waitGroup *sync.WaitGroup
	routes    []*Route
	lastError error
}

func NewHttp(
	config *config.HttpService,
	log *logrus.Logger,
) (*Http, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(config.Port)))
	if err != nil {
		return nil, err
	}

	service := &Http{
		waitGroup: &sync.WaitGroup{},
		routes:    make([]*Route, 0),
		listener:  listener,
		lastError: nil,
		server: &http.Server{
			ErrorLog: goLog.New(log.WriterLevel(logrus.ErrorLevel), "", 0),
		},
	}

	service.server.Handler = service.getHandlerFunc()

	service.waitGroup.Add(1)
	go func() {
		defer service.waitGroup.Done()
		service.lastError = service.server.Serve(service.listener)
	}()

	return service, nil
}

func (service *Http) AddEndpoint(route *Route) {
	service.routes = append(service.routes, route)
}

func (service *Http) Address() string {
	return "http://" + service.listener.Addr().String()
}

func (service *Http) getHandlerFunc() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		route := service.getMatchingRoute(request)
		if route == nil {
			http.NotFound(response, request)
			return
		}

		payload, err := service.getPayload(route, request.Body)
		if err != nil {
			http.Error(response, err.Error(), 500)
			return
		}

		params := service.getParams(route.Endpoint, request.URL)
		data, err := route.Handler(params, payload)
		if err != nil {
			http.Error(response, err.Error(), 500)
			return
		}

		response.WriteHeader(200)
		response.Header().Set("Content-Type", route.ResponseType)
		response.Write(data)
		return
	}
}

func (service *Http) getMatchingRoute(request *http.Request) *Route {
	for _, route := range service.routes {
		isValidGet := (request.Method == http.MethodGet && route.ExpectedPayloadType == nil)
		isValidPost := request.Method == http.MethodPost &&
			route.ExpectedPayloadType != nil &&
			request.Header.Get("Content-Type") == *route.ExpectedPayloadType
		if (isValidGet || isValidPost) && route.Endpoint.MatchString(request.URL.Path) {
			return route
		}
	}

	return nil
}

func (service *Http) getParams(endpoint *regexp.Regexp, url *url.URL) map[string]string {
	// Getting params from the query string
	params := make(map[string]string)
	for k, v := range url.Query() {
		params[k] = v[0]
	}

	// Adding params from the path's regex
	endpointSubexps := endpoint.SubexpNames()
	routeMatches := endpoint.FindStringSubmatch(url.Path)
	for i := 1; i < len(routeMatches) && i < len(endpointSubexps); i++ {
		params[endpointSubexps[i]] = routeMatches[i]
	}

	return params
}

func (service *Http) getPayload(route *Route, body io.Reader) ([]byte, error) {
	if route.ExpectedPayloadType != nil {
		return ioutil.ReadAll(body)
	}

	return make([]byte, 0), nil
}

func (service *Http) Wait() error {
	service.waitGroup.Wait()
	if service.lastError != http.ErrServerClosed {
		return service.lastError
	}

	return nil
}

func (service *Http) Close() error {
	err := service.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	return service.Wait()
}
