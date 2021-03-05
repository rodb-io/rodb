package record

import (
	"fmt"
	"rods/pkg/config"
	"rods/pkg/parser"
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
		value, err := record.Get(column.Name)
		if err != nil {
			return nil, err
		}

		result[column.Name] = value
	}

	return result, nil
}

func (record *Csv) Get(field string) (interface{}, error) {
	fieldIndex, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, fmt.Errorf("The column '%v' does not exist.", field)
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

func (record *Csv) Position() Position {
	return record.position
}
