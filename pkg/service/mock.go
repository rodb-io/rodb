package service

import ()

type Mock struct {
	routes []*Route
}

func NewMock() *Mock {
	return &Mock{
		routes: make([]*Route, 0),
	}
}

func (service *Mock) AddRoute(route *Route) {
	service.routes = append(service.routes, route)
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
