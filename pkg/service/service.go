package service

import (
	"fmt"
	"github.com/sirupsen/logrus"
	configModule "rodb.io/pkg/config"
	"rodb.io/pkg/output"
)

type Service interface {
	Name() string
	Address() string
	Wait() error
	Close() error
}

type List = map[string]Service

func NewFromConfig(
	config configModule.Service,
	outputs map[string]output.Output,
) (Service, error) {
	switch config.(type) {
	case *configModule.HttpService:
		return NewHttp(config.(*configModule.HttpService), outputs)
	default:
		return nil, fmt.Errorf("Unknown service config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]configModule.Service,
	outputs map[string]output.Output,
	log *logrus.Logger,
) (List, error) {
	services := make(List)
	for serviceName, serviceConfig := range configs {
		service, err := NewFromConfig(serviceConfig, outputs)
		if err != nil {
			return nil, err
		}
		services[serviceName] = service

		log.Infof("Service '%v' is up and available here: %v\n", serviceName, service.Address())
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
