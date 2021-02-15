package index

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/input"
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

func (d *Dumb) DoesIndex(inputName string, columnName string) bool {
	return true
}

func (d *Dumb) Close() error {
	return nil
}
