package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"rods/pkg/config"
)

type Service interface {
	AddRoute(route *Route)
	Address() string
	Wait() error
	Close() error
}

type List = map[string]Service

type Route struct {
	Endpoint            *regexp.Regexp
	ExpectedPayloadType *string
	ResponseType        string
	Handler             func(params map[string]string, payload []byte) ([]byte, error)
}

func NewFromConfig(
	config config.Service,
	log *logrus.Logger,
) (Service, error) {
	if config.Http != nil {
		return NewHttp(config.Http, log)
	}

	return nil, errors.New("Failed to initialize source")
}

func NewFromConfigs(
	configs map[string]config.Service,
	log *logrus.Logger,
) (List, error) {
	services := make(List)
	for serviceName, serviceConfig := range configs {
		service, err := NewFromConfig(serviceConfig, log)
		if err != nil {
			return nil, err
		}
		services[serviceName] = service
	}

	return services, nil
}

func Wait(services List) error {
	for serviceName, service := range services {
		err := service.Wait()
		if err != nil {
			return fmt.Errorf("%v service: %w", serviceName, err)
		}
	}

	return nil
}

func Close(services List) error {
	for _, service := range services {
		err := service.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
