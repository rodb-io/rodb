package index

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rods/pkg/input"
	"rods/pkg/record"
)

// A dumb index is able to search into any data,
// but very inefficiently. It does not index anything.
type Dumb struct {
	inputs input.List
	logger *logrus.Logger
}

func NewDumb(
	inputs input.List,
	log *logrus.Logger,
) (*Dumb, error) {
	return &Dumb{
		inputs: inputs,
		logger: log,
	}, nil
}

func (d *Dumb) Prepare() error {
	return nil
}

func (d *Dumb) GetRecordsByColumn(inputName string, columnName string, limit uint) ([]record.Record, error) {
	input, ok := d.inputs[inputName]
	if !ok {
		return nil, fmt.Errorf("There is no input named '%v'", inputName)
	}

	records := make([]record.Record, 0)
	for result := range input.IterateAll() {
		if result.Error != nil {
			return nil, result.Error
		}

		value, err := result.Record.Get(columnName)
		if err != nil {
			return nil, err
		}

		if value == columnName {
			records = append(records, result.Record)
		}

		if uint(len(records)) >= limit {
			break
		}
	}

	return records, nil
}

func (d *Dumb) Close() error {
	return nil
}
