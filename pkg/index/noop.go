package index

import (
	"reflect"
	"rods/pkg/input"
	"rods/pkg/record"
)

// A noop index is able to search into any data,
// but very inefficiently. It does not index anything.
type Noop struct {
	inputs input.List
}

func NewNoop(
	inputs input.List,
) *Noop {
	return &Noop{
		inputs: inputs,
	}
}

func (d *Noop) GetRecordPositions(input input.Input, filters map[string]interface{}, limit uint) (record.PositionList, error) {
	records := make(record.PositionList, 0)
	for result := range input.IterateAll() {
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
			records = append(records, result.Record.Position())
			if limit != 0 && len(records) >= int(limit) {
				break
			}
		}
	}

	return records, nil
}

func (d *Noop) Close() error {
	return nil
}
