package service

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
)

type Http struct {
}

func NewHttp(
	config *config.HttpService,
	log *logrus.Logger,
) (*Http, error) {
	return &Http{
	}, nil
}

func (http *Http) Close() error {
	return nil
}
