package config

type CsvInputConfig struct{
	Source string
	IgnoreFirstRow string
	Columns []CsvInputColumnConfig
}

type CsvInputColumnConfig struct{
	Name string
	Type *string
	IgnoreCharacters *string
	DecimalSeparator *string
	TrueValues []string
	FalseValues []string
}

func (config *CsvInputConfig) validate() error {
	return nil
}
