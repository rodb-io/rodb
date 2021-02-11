package record

import (
	"errors"
	"rods/pkg/config"
	"rods/pkg/utils"
	"strconv"
	"strings"
)

type CsvRecord struct {
	config *config.CsvInputConfig
	data   []string
}

func NewCsvRecord(config *config.CsvInputConfig, data []string) *CsvRecord {
	return &CsvRecord{
		config: config,
		data:   data,
	}
}

func (record *CsvRecord) getField(field string) (*string, int, error) {
	index, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, index, errors.New("The column '" + field + "' does not exist.")
	}

	if index >= len(record.data) {
		return nil, index, nil
	}

	return utils.PString(record.data[index]), index, nil
}

func (record *CsvRecord) GetString(field string) (*string, error) {
	value, _, err := record.getField(field)
	return value, err
}

func (record *CsvRecord) GetInteger(field string) (*int, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	cleanedValue := utils.RemoveCharacters(*value, columnConfig.IgnoreCharacters)

	intValue, err := strconv.Atoi(cleanedValue)
	if err != nil {
		return nil, err
	}

	return utils.PInt(intValue), nil
}

func (record *CsvRecord) GetFloat(field string) (*float64, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	cleanedValue := utils.RemoveCharacters(*value, columnConfig.IgnoreCharacters)

	if columnConfig.DecimalSeparator != "." {
		cleanedValue = strings.ReplaceAll(cleanedValue, columnConfig.DecimalSeparator, ".")
	}

	floatValue, err := strconv.ParseFloat(cleanedValue, 64)
	if err != nil {
		return nil, err
	}

	return utils.PFloat(floatValue), nil
}

func (record *CsvRecord) GetBoolean(field string) (*bool, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil || value == nil {
		return nil, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	if utils.IsInArray(*value, columnConfig.TrueValues) {
		return utils.PBool(true), nil
	}
	if utils.IsInArray(*value, columnConfig.FalseValues) {
		return utils.PBool(false), nil
	}

	return nil, errors.New("The value '" + *value + "' was found but is neither declared in trueValues or falseValues.")
}
