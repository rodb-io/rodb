package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"rods/pkg/types"
)

type JsonObjectOutput struct {
	Service    string                   `yaml:"service"`
	Input      string                   `yaml:"input"`
	Endpoint   string                   `yaml:"endpoint"`
	Index      string                   `yaml:"index"`
	Parameters []JsonObjectOutputParams `yaml:"params"`
}

type JsonObjectOutputParams struct {
	Column string `yaml:"column"`
	Type   string `yaml:"type"`
	typeDefinition types.Type
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

	return nil
}

func (config *JsonObjectOutputParams) validate(log *logrus.Logger) error {
	// The existence of the column value will be validated at runtime
	if config.Column == "" {
		return errors.New("jsonObject.parameters[].column is empty")
	}

	typeDefinition, err := types.NewFromString(config.Type)
	if err != nil {
		return fmt.Errorf("jsonObject.parameters[].type: '%w'", err)
	}
	config.typeDefinition = typeDefinition

	return nil
}

func (config *JsonObjectOutputParams) TypeDefinition() types.Type {
	return config.typeDefinition
}
