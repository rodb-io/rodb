package config

import (
	"github.com/sirupsen/logrus"
)

type StringParser struct {
}

func (config *StringParser) validate(log *logrus.Logger) error {
	return nil
}
