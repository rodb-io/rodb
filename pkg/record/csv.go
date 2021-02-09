package record

import (
	"rods/pkg/config"
)

type CsvRecord struct{
	config *config.CsvInputConfig
	data []string
}

func (record *CsvRecord) GetString(field string) (string, bool) {
	index, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return "", false
	}

	if index >= len(record.data) {
		return "", false
	}

	return record.data[index], true
}

// TODO replace exists boolean return value with proper error
// TODO implement other available types
// TODO implement properly the record for csv
// TODO implement the record interface
// TODO use the record interface in the csv input return and test
