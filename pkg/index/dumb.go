package index

import (
	"github.com/sirupsen/logrus"
)

// A dumb index is able to search into any data,
// but very inefficiently. It does not index anything.
type Dumb struct {
	logger *logrus.Logger
}

func NewDumb(
	log *logrus.Logger,
) (*Dumb, error) {
	return &Dumb{
		logger: log,
	}, nil
}

func (d *Dumb) DoesIndex(inputName string, columnName string) bool {
	return true
}

func (d *Dumb) Close() error {
	return nil
}
