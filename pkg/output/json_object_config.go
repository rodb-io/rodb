package output

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	parameterPackage "rodb.io/pkg/output/parameter"
	relationshipPackage "rodb.io/pkg/output/relationship"
	inputPackage "rodb.io/pkg/input"
	parserPackage "rodb.io/pkg/parser"
	indexPackage "rodb.io/pkg/index"
)

type JsonObjectConfig struct {
	Name          string                                 `yaml:"name"`
	Type          string                                 `yaml:"type"`
	Input         string                                 `yaml:"input"`
	Parameters    map[string]*parameterPackage.ParameterConfig    `yaml:"parameters"`
	Relationships map[string]*relationshipPackage.RelationshipConfig `yaml:"relationships"`
	Logger        *logrus.Entry
}

func (config *JsonObjectConfig) GetName() string {
	return config.Name
}

func (config *JsonObjectConfig) Validate(
	inputs map[string]inputPackage.Config,
	indexes map[string]indexPackage.Config,
	parsers map[string]parserPackage.Config,
	log *logrus.Entry,
) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("jsonObject.name is required")
	}

	if len(config.Parameters) == 0 {
		return errors.New("jsonObject.parameters is empty. As least one is required.")
	}

	if config.Input == "" {
		return errors.New("jsonObject.input is empty. This field is required.")
	}
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("jsonObject.input: Input '%v' not found in inputs list.", config.Input)
	}

	for parameterName, parameter := range config.Parameters {
		logPrefix := fmt.Sprintf("jsonObject.parameters.%v.", parameterName)
		if err := parameter.Validate(indexes, parsers, log, logPrefix, input); err != nil {
			return fmt.Errorf("jsonObject.parameters.%v.%w", parameterName, err)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		logPrefix := fmt.Sprintf("jsonObject.relationships.%v.", relationshipIndex)
		if err := relationship.Validate(indexes, inputs, log, logPrefix); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}
	}

	return nil
}
