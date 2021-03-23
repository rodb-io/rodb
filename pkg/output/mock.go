package output

import ()

type Mock struct {
}

func NewMock() *Mock {
	mock := &Mock{}

	return mock
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) Close() error {
	return nil
}
