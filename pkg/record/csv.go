package record

import (
	"errors"
	"rods/pkg/config"
	"rods/pkg/util"
	"strconv"
	"strings"
)

type Csv struct {
	config *config.CsvInput
	data   []string
	position Position
}

func NewCsv(
	config *config.CsvInput,
	data []string,
	position Position,
) *Csv {
	return &Csv{
		config: config,
		data:   data,
		position: position,
	}
}

func (record *Csv) getField(field string) (*string, int, error) {
	index, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, index, errors.New("The column '" + field + "' does not exist.")
	}

	if index >= len(record.data) {
		return nil, index, nil
	}

	return util.PString(record.data[index]), index, nil
}

func (record *Csv) GetString(field string) (*string, error) {
	value, _, err := record.getField(field)
	return value, err
}

func (record *Csv) GetInteger(field string) (*int, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
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
	if util.IsInArray(*value, columnConfig.TrueValues) {
		return util.PBool(true), nil
	}
	if util.IsInArray(*value, columnConfig.FalseValues) {
		return util.PBool(false), nil
	}

	return nil, errors.New("The value '" + *value + "' was found but is neither declared in trueValues or falseValues.")
}

func (record *Csv) Position() Position {
	return record.position
}
