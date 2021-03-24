package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	goLog "log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"rods/pkg/config"
	"rods/pkg/util"
	"sync"
)

type Http struct {
	config     *config.HttpService
	listener   net.Listener
	server     *http.Server
	waitGroup  *sync.WaitGroup
	routes     []*Route
	routesLock *sync.Mutex
	lastError  error
}

func NewHttp(
	config *config.HttpService,
) (*Http, error) {
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return nil, err
	}

	service := &Http{
		config:     config,
		waitGroup:  &sync.WaitGroup{},
		routes:     make([]*Route, 0),
		routesLock: &sync.Mutex{},
		listener:   listener,
		lastError:  nil,
		server: &http.Server{
			ErrorLog: goLog.New(config.Logger.WriterLevel(logrus.ErrorLevel), "", 0),
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

func (service *Http) Name() string {
	return service.config.Name
}

func (service *Http) AddRoute(route *Route) {
	service.routesLock.Lock()
	defer service.routesLock.Unlock()

	service.routes = append(service.routes, route)
}

func (service *Http) DeleteRoute(route *Route) {
	service.routesLock.Lock()
	defer service.routesLock.Unlock()

	routes := service.routes
	for i, v := range routes {
		if v == route {
			service.routes = append(routes[:i], routes[i+1:]...)
			return
		}
	}
}

func (service *Http) Address() string {
	return "http://" + util.GetAddress(service.listener.Addr())
}

func (service *Http) getHandlerFunc() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		route := service.getMatchingRoute(request)
		if route == nil {
			errToSend := errors.New("No matching route was found")
			err2 := service.sendErrorResponse(response, http.StatusNotFound, errToSend)
			if err2 != nil {
				service.config.Logger.Errorf("Error '%+v' while sending the error '%+v'", errToSend, err2)
			}
			return
		}

		payload, err := service.getPayload(route, request.Body)
		if err != nil {
			err2 := service.sendErrorResponse(response, http.StatusInternalServerError, err)
			if err2 != nil {
				service.config.Logger.Errorf("Error '%+v' while sending the error '%+v'", err, err2)
			}
			return
		}

		params := service.getParams(route.Endpoint, request.URL)
		err = route.Handler(
			params,
			payload,
			func(err error) error {
				status := http.StatusInternalServerError
				if errors.Is(err, RecordNotFoundError) {
					status = http.StatusNotFound
				}

				return service.sendErrorResponse(response, status, err)
			},
			func() io.Writer {
				response.Header().Set("Content-Type", route.ResponseType+"; charset=UTF-8")
				response.WriteHeader(http.StatusOK)
				return io.Writer(response)
			},
		)
		if err != nil {
			service.config.Logger.Errorf("Unhandled error while handling the route '%v': %v", route.Endpoint, err)
		}

		return
	}
}

func (service *Http) sendErrorResponse(
	response http.ResponseWriter,
	status int,
	err error,
) error {
	var data []byte
	var outputType string = service.config.ErrorsType
	switch outputType {
	case "application/json":
		data, err = json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		if err != nil {
			return err
		}
	default:
		response.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		response.WriteHeader(status)
		_, err = response.Write([]byte(err.Error()))
		if err != nil {
			return err
		}

		return fmt.Errorf("ErrorResponse type '%v' is not supported by the HTTP service", service.config.ErrorsType)
	}

	response.Header().Set("Content-Type", outputType+"; charset=UTF-8")
	response.WriteHeader(status)
	_, err = response.Write(data)
	if err != nil {
		return err
	}

	return nil
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
