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
	"rodb.io/pkg/config"
	"rodb.io/pkg/output"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/util"
	"sync"
)

type Http struct {
	config         *config.HttpService
	httpListener   net.Listener
	httpsListener  net.Listener
	httpServer     *http.Server
	httpsServer    *http.Server
	waitGroup      *sync.WaitGroup
	routes         []*httpRoute
	lastHttpError  error
	lastHttpsError error
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
	service := &Http{
		config:         config,
		waitGroup:      &sync.WaitGroup{},
		lastHttpError:  nil,
		lastHttpsError: nil,
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

	var err error
	if config.Http != nil {
		service.httpListener, service.httpServer, err = service.createServer(config.Http.Listen)
		if err != nil {
			return nil, err
		}

		service.waitGroup.Add(1)
		go func() {
			defer service.waitGroup.Done()
			service.lastHttpError = service.httpServer.Serve(service.httpListener)
		}()
	}
	if config.Https != nil {
		service.httpsListener, service.httpsServer, err = service.createServer(config.Https.Listen)
		if err != nil {
			return nil, err
		}

		service.waitGroup.Add(1)
		go func() {
			defer service.waitGroup.Done()
			service.lastHttpsError = service.httpsServer.ServeTLS(
				service.httpsListener,
				config.Https.CertificatePath,
				config.Https.PrivateKeyPath,
			)
		}()
	}

	return service, nil
}

func (service *Http) createServer(listen string) (net.Listener, *http.Server, error) {
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, nil, err
	}

	server := &http.Server{
		ErrorLog: goLog.New(service.config.Logger.WriterLevel(logrus.ErrorLevel), "", 0),
		Handler:  service.getHandlerFunc(),
	}

	return listener, server, nil
}

func (service *Http) Name() string {
	return service.config.Name
}

func (service *Http) Address() string {
	if service.config.Https != nil {
		return "https://" + util.GetAddress(service.httpsListener.Addr())
	} else {
		return "http://" + util.GetAddress(service.httpListener.Addr())
	}
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

		if !parser.Primitive() {
			return nil, nil, fmt.Errorf("Cannot use the parser '%v' as route parameter because it's not a primitive.", parser.Name())
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
		response.Header().Set("X-Powered-By", "RODB (rodb.io)")

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
		sendError := func(err error) error {
			status := http.StatusInternalServerError
			if errors.Is(err, record.RecordNotFoundError) {
				status = http.StatusNotFound
			}

			return service.sendErrorResponse(response, status, err)
		}
		sendSuccess := func() io.Writer {
			response.Header().Set("Content-Type", route.output.ResponseType()+"; charset=UTF-8")
			response.WriteHeader(http.StatusOK)
			return io.Writer(response)
		}
		if err := route.output.Handle(params, payload, sendError, sendSuccess); err != nil {
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
	if service.config.Http != nil && service.lastHttpError != http.ErrServerClosed {
		return service.lastHttpError
	}
	if service.config.Https != nil && service.lastHttpsError != http.ErrServerClosed {
		return service.lastHttpsError
	}

	return nil
}

func (service *Http) Close() error {
	if service.config.Http != nil {
		if err := service.httpServer.Shutdown(context.Background()); err != nil {
			return err
		}
	}

	if service.config.Https != nil {
		if err := service.httpsServer.Shutdown(context.Background()); err != nil {
			return err
		}
	}

	return service.Wait()
}
