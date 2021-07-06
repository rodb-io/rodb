package input

import (
	"fmt"
	"rodb.io/pkg/parser"
	recordPackage "rodb.io/pkg/input/record"
	"strings"
)

type CsvRecord struct {
	config        *CsvConfig
	columnParsers []parser.Parser
	data          []string
	position      recordPackage.Position
}

func NewCsvRecord(
	config *CsvConfig,
	columnParsers []parser.Parser,
	data []string,
	position recordPackage.Position,
) *CsvRecord {
	return &CsvRecord{
		config:        config,
		columnParsers: columnParsers,
		data:          data,
		position:      position,
	}
}

func (record *CsvRecord) All() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, column := range record.config.Columns {
		value, err := record.getColumn(column.Name)
		if err != nil {
			return nil, err
		}

		result[column.Name] = value
	}

	return result, nil
}

// Gets the value of the column (no dot-separated names, only the column content itself)
func (record *CsvRecord) getColumn(columnName string) (interface{}, error) {
	fieldIndex, exists := record.config.ColumnIndexByName[columnName]
	if !exists {
		return nil, fmt.Errorf("The column '%v' does not exist.", columnName)
	}

	if fieldIndex >= len(record.data) {
		return nil, nil
	}

	parser := record.columnParsers[fieldIndex]
	value, err := parser.Parse(record.data[fieldIndex])
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (record *CsvRecord) Get(path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("Cannot get the property '%v' because it's path is empty.", path)
	}

	pathArray := strings.Split(path, ".")

	value, err := record.getColumn(pathArray[0])
	if err != nil {
		return nil, err
	}

	return recordPackage.GetSubValue(value, pathArray[1:])
}

func (record *CsvRecord) Position() recordPackage.Position {
	return record.position
}
