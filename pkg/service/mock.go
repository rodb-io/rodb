package service

import ()

type Mock struct {
	Routes []*Route
}

func NewMock() *Mock {
	return &Mock{
		Routes: make([]*Route, 0),
	}
}

func (mock *Mock) Name() string {
	return "mock"
}

func (service *Mock) AddRoute(route *Route) {
	service.Routes = append(service.Routes, route)
}

func (service *Mock) DeleteRoute(route *Route) {
	routes := service.Routes
	for i, v := range routes {
		if v == route {
			service.Routes = append(routes[:i], routes[i+1:]...)
			return
		}
	}
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
