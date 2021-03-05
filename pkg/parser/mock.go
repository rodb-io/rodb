package parser

type Mock struct{}

func NewMock() *Mock {
	return &Mock{}
}

func (string *Mock) GetRegexpPattern() string {
	return ".*"
}

func (string *Mock) Parse(value string) (interface{}, error) {
	return value, nil
}
