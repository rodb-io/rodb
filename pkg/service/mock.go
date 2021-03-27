package service

import ()

type Mock struct {
}

func NewMock() *Mock {
	return &Mock{
	}
}

func (mock *Mock) Name() string {
	return "mock"
}

func (service *Mock) Address() string {
	return ""
}

func (service *Mock) Wait() error {
	return nil
}

func (service *Mock) Close() error {
	return nil
}
