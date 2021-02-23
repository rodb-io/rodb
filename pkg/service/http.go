package service

import (
	"context"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"rods/pkg/config"
	"strconv"
	"sync"
)

type Http struct {
	server    *http.Server
	waitGroup *sync.WaitGroup
	routes    []Route
}

func NewHttp(
	config *config.HttpService,
	waitGroup *sync.WaitGroup,
	log *logrus.Logger,
) (*Http, error) {
	service := &Http{
		waitGroup: waitGroup,
		routes:    make([]Route, 0),
		server: &http.Server{
			Addr: ":" + strconv.Itoa(int(config.Port)),
		},
	}

	service.server.Handler = service.getHandlerFunc()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		err := service.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Http service: %v", err)
		}
	}()

	return service, nil
}

func (service *Http) AddEndpoint(route Route) {
	service.routes = append(service.routes, route)
}

func (service *Http) getHandlerFunc() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		route := service.getMatchingRoute(request)
		if route == nil {
			http.NotFound(response, request)
			return
		}

		payload, err := service.getPayload(route, request)
		if err != nil {
			http.Error(response, err.Error(), 500)
			return
		}

		params := service.getParams(route, request.URL)
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
			return &route
		}
	}

	return nil
}

func (service *Http) getParams(route *Route, url *url.URL) map[string]string {
	// Getting params from the query string
	params := make(map[string]string)
	for k, v := range url.Query() {
		params[k] = v[0]
	}

	// Adding params from the path's regex
	endpointSubexps := route.Endpoint.SubexpNames()
	routeMatches := route.Endpoint.FindStringSubmatch(url.Path)
	for i := 1; i < len(routeMatches) && i < len(endpointSubexps); i++ {
		params[endpointSubexps[i]] = routeMatches[i]
	}

	return params
}

func (service *Http) getPayload(route *Route, request *http.Request) ([]byte, error) {
	if route.ExpectedPayloadType != nil {
		return ioutil.ReadAll(request.Body)
	}

	return make([]byte, 0), nil
}

func (service *Http) Close() error {
	err := service.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	service.waitGroup.Wait()
	return nil
}
