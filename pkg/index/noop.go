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
	inputIterator, end, err := input.IterateAll()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := end()
		if err != nil {
			noop.config.Logger.Errorf("Error while closing input iterator: %v", err)
		}
	}()

	return func() (*record.Position, error) {
		for {
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

		return nil, nil
	}, nil
}

func (noop *Noop) Close() error {
	return nil
}
