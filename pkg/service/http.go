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
	outputs   []output.Output
	lastError error
}

func NewHttp(
	config *config.HttpService,
	outputs map[string]output.Output,
) (*Http, error) {
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return nil, err
	}

	boundOutputs := make([]output.Output, 0, len(config.Routes))
	for _, route := range config.Routes {
		output, outputExists := outputs[route.Output]
		if !outputExists {
			return nil, fmt.Errorf("Output '%v' not found in outputs list.", route.Output)
		}

		boundOutputs = append(boundOutputs, output)
	}

	service := &Http{
		config:    config,
		waitGroup: &sync.WaitGroup{},
		outputs:   boundOutputs,
		listener:  listener,
		lastError: nil,
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

func (service *Http) Address() string {
	return "http://" + util.GetAddress(service.listener.Addr())
}

func (service *Http) getHandlerFunc() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		output := service.getMatchingOutput(request)
		if output == nil {
			errToSend := errors.New("No matching output was found")
			err2 := service.sendErrorResponse(response, http.StatusNotFound, errToSend)
			if err2 != nil {
				service.config.Logger.Errorf("Error '%+v' while sending the error '%+v'", errToSend, err2)
			}
			return
		}

		payload, err := service.getPayload(output, request.Body)
		if err != nil {
			err2 := service.sendErrorResponse(response, http.StatusInternalServerError, err)
			if err2 != nil {
				service.config.Logger.Errorf("Error '%+v' while sending the error '%+v'", err, err2)
			}
			return
		}

		params := service.getParams(output.Endpoint(), request.URL)
		err = output.Handle(
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
				response.Header().Set("Content-Type", output.ResponseType()+"; charset=UTF-8")
				response.WriteHeader(http.StatusOK)
				return io.Writer(response)
			},
		)
		if err != nil {
			service.config.Logger.Errorf("Unhandled error while handling the output '%v': %v", output.Endpoint(), err)
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

func (service *Http) getMatchingOutput(request *http.Request) output.Output {
	for _, output := range service.outputs {
		expectedPayloadType := output.ExpectedPayloadType()
		isValidGet := (request.Method == http.MethodGet && expectedPayloadType == nil)
		isValidPost := request.Method == http.MethodPost &&
			expectedPayloadType != nil &&
			request.Header.Get("Content-Type") == *expectedPayloadType
		if (isValidGet || isValidPost) && output.Endpoint().MatchString(request.URL.Path) {
			return output
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
	outputMatches := endpoint.FindStringSubmatch(url.Path)
	for i := 1; i < len(outputMatches) && i < len(endpointSubexps); i++ {
		params[endpointSubexps[i]] = outputMatches[i]
	}

	return params
}

func (service *Http) getPayload(output output.Output, body io.Reader) ([]byte, error) {
	if output.ExpectedPayloadType() != nil {
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
