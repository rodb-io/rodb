package record

import (
	"fmt"
	"rods/pkg/config"
	"rods/pkg/util"
	"strconv"
	"strings"
)

type Csv struct {
	config   *config.CsvInput
	data     []string
	position Position
}

func NewCsv(
	config *config.CsvInput,
	data []string,
	position Position,
) *Csv {
	return &Csv{
		config:   config,
		data:     data,
		position: position,
	}
}

func (record *Csv) getField(field string) (*string, int, error) {
	index, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, index, fmt.Errorf("The column '%v' does not exist.", field)
	}

	if index >= len(record.data) {
		return nil, index, nil
	}

	return util.PString(record.data[index]), index, nil
}

func (record *Csv) GetString(field string) (*string, error) {
	value, fieldIndex, err := record.getField(field)

	columnConfig := record.config.Columns[fieldIndex]
	if columnConfig.Type != config.String {
		return nil, fmt.Errorf("The column '%v' is not a string", field)
	}

	return value, err
}

func (record *Csv) GetInteger(field string) (*int, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	if columnConfig.Type != config.Integer {
		return nil, fmt.Errorf("The column '%v' is not an integer", field)
	}

	cleanedValue := util.RemoveCharacters(*value, columnConfig.IgnoreCharacters)
	intValue, err := strconv.Atoi(cleanedValue)
	if err != nil {
		return nil, err
	}

	return util.PInt(intValue), nil
}

func (record *Csv) GetFloat(field string) (*float64, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	if columnConfig.Type != config.Float {
		return nil, fmt.Errorf("The column '%v' is not a float", field)
	}

	cleanedValue := util.RemoveCharacters(*value, columnConfig.IgnoreCharacters)
	if columnConfig.DecimalSeparator != "." {
		cleanedValue = strings.ReplaceAll(cleanedValue, columnConfig.DecimalSeparator, ".")
	}

	floatValue, err := strconv.ParseFloat(cleanedValue, 64)
	if err != nil {
		return nil, err
	}

	return util.PFloat(floatValue), nil
}

func (record *Csv) GetBoolean(field string) (*bool, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	if columnConfig.Type != config.Boolean {
		return nil, fmt.Errorf("The column '%v' is not a boolean", field)
	}

	if util.IsInArray(*value, columnConfig.TrueValues) {
		return util.PBool(true), nil
	}
	if util.IsInArray(*value, columnConfig.FalseValues) {
		return util.PBool(false), nil
	}

	return nil, fmt.Errorf("The value '%v' was found but is neither declared in trueValues or falseValues.", *value)
}

func (record *Csv) Get(field string) (interface{}, error) {
	fieldIndex, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, fmt.Errorf("The column '%v' does not exist.", field)
	}

	columnConfig := record.config.Columns[fieldIndex]
	if columnConfig.Type == config.String {
		return record.GetString(field)
	}
	if columnConfig.Type == config.Integer {
		return record.GetInteger(field)
	}
	if columnConfig.Type == config.Float {
		return record.GetFloat(field)
	}
	if columnConfig.Type == config.Boolean {
		return record.GetBoolean(field)
	}

	return nil, fmt.Errorf("Unknown type '%v' in CsvRecord.Get", columnConfig.Type)
}

func (record *Csv) Position() Position {
	return record.position
}
