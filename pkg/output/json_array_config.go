package output

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	indexPackage "rodb.io/pkg/index"
	inputPackage "rodb.io/pkg/input"
	parameterPackage "rodb.io/pkg/output/parameter"
	relationshipPackage "rodb.io/pkg/output/relationship"
	parserPackage "rodb.io/pkg/parser"
)

type JsonArrayConfig struct {
	Name          string                                             `yaml:"name"`
	Type          string                                             `yaml:"type"`
	Input         string                                             `yaml:"input"`
	Limit         JsonArrayLimitConfig                               `yaml:"limit"`
	Offset        JsonArrayOffsetConfig                              `yaml:"offset"`
	Parameters    map[string]*parameterPackage.ParameterConfig       `yaml:"parameters"`
	Relationships map[string]*relationshipPackage.RelationshipConfig `yaml:"relationships"`
	Logger        *logrus.Entry
}

type JsonArrayLimitConfig struct {
	Default   uint   `yaml:"default"`
	Max       uint   `yaml:"max"`
	Parameter string `yaml:"parameter"`
}

type JsonArrayOffsetConfig struct {
	Parameter string `yaml:"parameter"`
}

func (config *JsonArrayConfig) GetName() string {
	return config.Name
}

func (config *JsonArrayConfig) Validate(
	inputs map[string]inputPackage.Config,
	indexes map[string]indexPackage.Config,
	parsers map[string]parserPackage.Config,
	log *logrus.Entry,
) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("jsonArray.name is required")
	}

	if config.Input == "" {
		return errors.New("jsonArray.input is empty. This field is required.")
	}
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("jsonObject.input: Input '%v' not found in inputs list.", config.Input)
	}

	if err := config.Limit.Validate(log); err != nil {
		return fmt.Errorf("jsonArray.limit.%v", err)
	}

	if err := config.Offset.Validate(log); err != nil {
		return fmt.Errorf("jsonArray.offset.%v", err)
	}

	for configParamName, configParam := range config.Parameters {
		logPrefix := fmt.Sprintf("jsonArray.parameters.%v.", configParamName)
		if err := configParam.Validate(indexes, parsers, log, logPrefix, input); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if configParamName == config.Limit.Parameter {
			return fmt.Errorf("jsonArray.parameters.%v: Parameter '%v' is already used for the limit", configParamName, configParamName)
		}
		if configParamName == config.Offset.Parameter {
			return fmt.Errorf("jsonArray.parameters.%v: Parameter '%v' is already used for the offset", configParamName, configParamName)
		}
	}

	for relationshipIndex, relationship := range config.Relationships {
		logPrefix := fmt.Sprintf("jsonArray.relationships.%v.", relationshipIndex)
		if err := relationship.Validate(indexes, inputs, log, logPrefix); err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}
	}

	return nil
}

func (config *JsonArrayLimitConfig) Validate(log *logrus.Entry) error {
	if config.Default == 0 {
		log.Debug("jsonArray.limit.default not set. Assuming '100'")
		config.Default = 100
	}

	if config.Max == 0 {
		log.Debug("jsonArray.limit.max not set. Assuming '1000'")
		config.Max = 1000
	}

	if config.Default > config.Max {
		return fmt.Errorf("default is higher than the max value of %v.", config.Max)
	}

	if config.Parameter == "" {
		log.Debug("jsonArray.limit.parameter not set. Assuming 'limit'")
		config.Parameter = "limit"
	}

	return nil
}

func (config *JsonArrayOffsetConfig) Validate(log *logrus.Entry) error {
	if config.Parameter == "" {
		log.Debug("jsonArray.offset.parameter not set. Assuming 'offset'")
		config.Parameter = "offset"
	}

	return nil
}
