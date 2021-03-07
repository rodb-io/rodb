package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"strings"
)

type JsonObjectOutput struct {
	Service    string                   `yaml:"service"`
	Input      string                   `yaml:"input"`
	Endpoint   string                   `yaml:"endpoint"`
	Index      string                   `yaml:"index"`
	Parameters []JsonObjectOutputParams `yaml:"parameters"`
}

type JsonObjectOutputParams struct {
	Column string `yaml:"column"`
	Parser string `yaml:"parser"`
}

func (config *JsonObjectOutput) validate(log *logrus.Logger) error {
	// The service will be validated at runtime
	// The index will be validated at runtime
	// The default index value "" matches the dumb index
	// The input will be validated at runtime

	if len(config.Parameters) == 0 {
		return errors.New("A jsonObject output must have at least one parameter")
	}

	for _, parameter := range config.Parameters {
		err := parameter.validate(log)
		if err != nil {
			return err
		}
	}

	if config.Endpoint == "" {
		return errors.New("You must specify a non-empty jsonObject endpoint")
	}

	if !strings.Contains(config.Endpoint, "?") {
		return errors.New("A jsonObject endpoint must specify the identifier's location with '?'. For example \"/product/?\".")
	}

	if strings.Count(config.Endpoint, "?") != len(config.Parameters) {
		return errors.New("A jsonObject must have the same number of parameters than occurences of '?' in the endpoint")
	}

	return nil
}

func (config *JsonObjectOutputParams) validate(log *logrus.Logger) error {
	// The existence of the column value will be validated at runtime
	// The parser will be validated at runtime

	if config.Column == "" {
		return errors.New("jsonObject.parameters[].column is empty")
	}

	if config.Parser == "" {
		log.Debug("jsonObjet.parameters[].parser not defined. Assuming 'string'")
		config.Parser = "string"
	}

	return nil
}
