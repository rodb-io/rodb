package record

import (
	"fmt"
	"rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"strconv"
	"strings"
)

type Csv struct {
	config        *config.CsvInput
	columnParsers []parser.Parser
	data          []string
	position      Position
}

func NewCsv(
	config *config.CsvInput,
	columnParsers []parser.Parser,
	data []string,
	position Position,
) *Csv {
	return &Csv{
		config:        config,
		columnParsers: columnParsers,
		data:          data,
		position:      position,
	}
}

func (record *Csv) All() (map[string]interface{}, error) {
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
func (record *Csv) getColumn(columnName string) (interface{}, error) {
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

func (record *Csv) getSubValue(value interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return value, nil
	}

	switch value.(type) {
	case map[string]interface{}:
		mapValue, mapValueExists := value.(map[string]interface{})[path[0]]
		if !mapValueExists {
			// Not having some properties is a common case that
			// should not trigger an error, but get a nil value
			return nil, nil
		}

		return record.getSubValue(mapValue, path[1:])
	case []interface{}:
		index, err := strconv.Atoi(path[0])
		if err != nil {
			return nil, fmt.Errorf("Cannot get sub-path '%v' because the requested key is '%v', but the value is an array '%#v'", path, path[0], value)
		}

		valueArray := value.([]interface{})
		if index >= len(valueArray) {
			// Not having the required index is a common case that
			// should not trigger an error, but get a nil value
			return nil, nil
		}

		return record.getSubValue(valueArray[index], path[1:])
	default:
		return nil, fmt.Errorf("Cannot get sub-path '%v' because the value is primitive or unknown: '%#v'", path, value)
	}
}

func (record *Csv) Get(path string) (interface{}, error) {
	if path == "" {
		// Avoid having an empty splitted array
		return record.getColumn(path)
	}

	pathArray := strings.Split(path, ".")

	value, err := record.getColumn(pathArray[0])
	if err != nil {
		return nil, err
	}

	return record.getSubValue(value, pathArray[1:])
}

func (record *Csv) Position() Position {
	return record.position
}
