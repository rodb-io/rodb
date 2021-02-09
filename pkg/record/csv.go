package record

import (
	"rods/pkg/config"
)

type CsvRecord struct{
	config []config.CsvInputColumnConfig
	data []string
}

func (record *CsvRecord) ColumnsCount() int {
	return len(record.data)
}
// TODO fix test
// TODO implement properly the record interface
// TODO make the csv input recorn a record
