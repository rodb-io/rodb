package input

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/rodb-io/rodb/pkg/input/record"
	"github.com/rodb-io/rodb/pkg/parser"
	"time"
)

type Input interface {
	Name() string
	Get(position record.Position) (record.Record, error)
	Size() (int64, error)
	ModTime() (time.Time, error)

	// Iterates all the records in the input, ordered
	// from the smallest to the biggest position
	// The second returned parameter is a callback that
	// must be used to close the relevant resources
	IterateAll() (record.Iterator, func() error, error)

	Close() error
}

type Config interface {
	Validate(parsers map[string]parser.Config, log *logrus.Entry) error
	GetName() string
	ShouldDieOnInputChange() bool
}

type List = map[string]Input

func NewFromConfig(
	config Config,
	parsers parser.List,
) (Input, error) {
	switch config.(type) {
	case *CsvConfig:
		return NewCsv(config.(*CsvConfig), parsers)
	case *XmlConfig:
		return NewXml(config.(*XmlConfig), parsers)
	case *JsonConfig:
		return NewJson(config.(*JsonConfig))
	default:
		return nil, fmt.Errorf("Unknown input config type: %#v", config)
	}
}

func NewFromConfigs(
	configs map[string]Config,
	parsers parser.List,
) (List, error) {
	inputs := make(List)
	for inputName, inputConfig := range configs {
		input, err := NewFromConfig(inputConfig, parsers)
		if err != nil {
			return nil, err
		}
		inputs[inputName] = input
	}

	return inputs, nil
}

func Close(inputs List) error {
	for _, input := range inputs {
		if err := input.Close(); err != nil {
			return err
		}
	}

	return nil
}
