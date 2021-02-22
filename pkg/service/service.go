package service

import (
	"errors"
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"sync"
)

type Service interface {
	Close() error
}

type List = map[string]Service

func NewFromConfig(
	config config.Service,
	waitGroup *sync.WaitGroup,
	log *logrus.Logger,
) (Service, error) {
	if config.Http != nil {
		return NewHttp(config.Http, waitGroup, log)
	}

	return nil, errors.New("Failed to initialize source")
}

func NewFromConfigs(
	configs map[string]config.Service,
	waitGroup *sync.WaitGroup,
	log *logrus.Logger,
) (List, error) {
	services := make(List)
	for serviceName, serviceConfig := range configs {
		service, err := NewFromConfig(serviceConfig, waitGroup, log)
		if err != nil {
			return nil, err
		}
		services[serviceName] = service
	}

	return services, nil
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
