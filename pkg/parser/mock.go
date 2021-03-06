package parser

type Mock struct {
	prefix string
}

func NewMock() *Mock {
	return &Mock{}
}

func NewMockWithPrefix(prefix string) *Mock {
	return &Mock{
		prefix: prefix,
	}
}

func (mock *Mock) Name() string {
	return "mock"
}

func (mock *Mock) Primitive() bool {
	return true
}

func (mock *Mock) GetRegexpPattern() string {
	return ".*"
}

func (mock *Mock) Parse(value string) (interface{}, error) {
	return mock.prefix + value, nil
}
