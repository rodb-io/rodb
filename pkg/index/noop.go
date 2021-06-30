package index

import (
	"fmt"
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
	inputIterator, end, err := input.IterateAll()
	if err != nil {
		return nil, err
	}

	closed := false

	return func() (*record.Position, error) {
		for {
			if closed {
				return nil, nil
			}

			record, err := inputIterator()
			if err != nil {
				return nil, err
			}
			if record == nil {
				break
			}

			matches := true
			for propertyName, filter := range filters {
				value, err := record.Get(propertyName)
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
				position := record.Position()
				return &position, nil
			}
		}

		closed = true
		if err := end(); err != nil {
			return nil, fmt.Errorf("Error while closing input iterator: %w", err)
		}

		return nil, nil
	}, nil
}

func (noop *Noop) Close() error {
	return nil
}
