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
	"rods/pkg/output"
	"rods/pkg/record"
	"rods/pkg/util"
	"sync"
)

type Http struct {
	config    *config.HttpService
	listener  net.Listener
	server    *http.Server
	waitGroup *sync.WaitGroup
	routes    []*httpRoute
	lastError error
}

type httpRoute struct {
	config     config.HttpServiceRoute
	path       *regexp.Regexp
	parameters []string
	output     output.Output
}

func NewHttp(
	config *config.HttpService,
	outputs map[string]output.Output,
) (*Http, error) {
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return nil, err
	}

	service := &Http{
		config:    config,
		waitGroup: &sync.WaitGroup{},
		listener:  listener,
		lastError: nil,
		server: &http.Server{
			ErrorLog: goLog.New(config.Logger.WriterLevel(logrus.ErrorLevel), "", 0),
		},
	}

	service.routes = make([]*httpRoute, 0, len(config.Routes))
	for _, route := range config.Routes {
		output, outputExists := outputs[route.Output]
		if !outputExists {
			return nil, fmt.Errorf("Output '%v' not found in outputs list.", route.Output)
		}

		routePath, parameters, err := service.createPathRegexp(route, output)
		if err != nil {
			return nil, fmt.Errorf("Cannot build regexp from route path '%v': %w", route.Path, err)
		}

		for _, paramName := range parameters {
			if !output.HasParameter(paramName) {
				return nil, fmt.Errorf("Output '%v' does not have a parameter called '%v'.", route.Output, paramName)
			}
		}

		service.routes = append(service.routes, &httpRoute{
			config:     route,
			path:       routePath,
			parameters: parameters,
			output:     output,
		})
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

func (service *Http) Address() string {
	return "http://" + util.GetAddress(service.listener.Addr())
}

// Returns a regular expression to match a string, and the list of param names
// (matching the sub-expressions of the regexp)
func (service *Http) createPathRegexp(
	routeConfig config.HttpServiceRoute,
	output output.Output,
) (*regexp.Regexp, []string, error) {
	paramRegexp, err := regexp.Compile("{([^}]+)}")
	if err != nil {
		return nil, nil, err
	}

	paramMatches := paramRegexp.FindAllStringSubmatch(routeConfig.Path, -1)
	params := make([]string, len(paramMatches))
	for i, paramMatch := range paramMatches {
		params[i] = paramMatch[1]
	}

	parts := paramRegexp.Split(routeConfig.Path, -1)
	path := parts[0]
	for partIndex := 1; partIndex < len(parts); partIndex++ {
		paramIndex := partIndex - 1
		paramName := params[paramIndex]

		parser, err := output.GetParameterParser(paramName)
		if err != nil {
			return nil, nil, err
		}

		paramPattern := parser.GetRegexpPattern()
		path = path + "(" + paramPattern + ")" + parts[partIndex]
	}

	regexp, err := regexp.Compile("^" + path + "$")
	if err != nil {
		return nil, nil, err
	}

	return regexp, params, nil
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

		params := service.getParams(route, request.URL)
		err = route.output.Handle(
			params,
			payload,
			func(err error) error {
				status := http.StatusInternalServerError
				if errors.Is(err, record.RecordNotFoundError) {
					status = http.StatusNotFound
				}

				return service.sendErrorResponse(response, status, err)
			},
			func() io.Writer {
				response.Header().Set("Content-Type", route.output.ResponseType()+"; charset=UTF-8")
				response.WriteHeader(http.StatusOK)
				return io.Writer(response)
			},
		)
		if err != nil {
			service.config.Logger.Errorf("Unhandled error while handling the route '%v': %v", route.config.Path, err)
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

func (service *Http) getMatchingRoute(request *http.Request) *httpRoute {
	for _, route := range service.routes {
		expectedPayloadType := route.output.ExpectedPayloadType()
		isValidGet := (request.Method == http.MethodGet && expectedPayloadType == nil)
		isValidPost := request.Method == http.MethodPost &&
			expectedPayloadType != nil &&
			request.Header.Get("Content-Type") == *expectedPayloadType
		if (isValidGet || isValidPost) && route.path.MatchString(request.URL.Path) {
			return route
		}
	}

	return nil
}

func (service *Http) getParams(route *httpRoute, url *url.URL) map[string]string {
	// Getting params from the query string
	params := make(map[string]string)
	for k, v := range url.Query() {
		params[k] = v[0]
	}

	// Adding params from the path's regex
	outputMatches := route.path.FindStringSubmatch(url.Path)
	for i, paramName := range route.parameters {
		params[paramName] = outputMatches[i+1]
	}

	return params
}

func (service *Http) getPayload(route *httpRoute, body io.Reader) ([]byte, error) {
	if route.output.ExpectedPayloadType() != nil {
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
