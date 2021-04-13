package index

import (
	"reflect"
	"rodb.io/pkg/config"
	"rodb.io/pkg/input"
	"rodb.io/pkg/record"
)

// A noop index is able to search into any data,
// but very inefficiently. It does not index anything.
type Noop struct {
	config *config.NoopIndex
	inputs input.List
}

func NewNoop(
	config *config.NoopIndex,
	inputs input.List,
) *Noop {
	return &Noop{
		config: config,
		inputs: inputs,
	}
}

func (noop *Noop) Name() string {
	return noop.config.Name
}

func (noop *Noop) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	inputIterator := input.IterateAll()
	return func() (*record.Position, error) {
		for result := range inputIterator {
			if result.Error != nil {
				return nil, result.Error
			}

			matches := true
			for columnName, filter := range filters {
				value, err := result.Record.Get(columnName)
				if err != nil {
					return nil, err
				}

				if value == nil {
					if filter == nil {
						continue
					} else {
						matches = false
						break
					}
				}

				value = reflect.ValueOf(value).Interface()
				if value != filter {
					matches = false
					break
				}
			}

			if matches {
				position := result.Record.Position()
				return &position, nil
			}
		}

		return nil, nil
	}, nil
}

func (noop *Noop) Close() error {
	return nil
}
