package types

type String struct{
}

func NewString() *String {
	return &String{
	}
}

func (string *String) GetRegexpPattern() string {
	return ".*"
}

func (string *String) Parse(value string) (interface{}, error) {
	return value, nil
}
