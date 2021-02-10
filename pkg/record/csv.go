package record

import (
	"errors"
	"strconv"
	"strings"
	"rods/pkg/config"
	"rods/pkg/utils"
)

type CsvRecord struct{
	config *config.CsvInputConfig
	data []string
}

func (record *CsvRecord) getField(field string) (string, int, error) {
	index, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return "", index, errors.New("The column '" + field + "' does not exist.")
	}

	if index >= len(record.data) {
		return "", index, errors.New("The column '" + field + "' was not found in this record.")
	}

	return record.data[index], index, nil
}

func (record *CsvRecord) GetString(field string) (string, error) {
	value, _, err := record.getField(field)
	return value, err
}

func (record *CsvRecord) GetInteger(field string) (int, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil {
		return 0, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	value = utils.RemoveCharacters(value, columnConfig.IgnoreCharacters)

	return strconv.Atoi(value)
}

func (record *CsvRecord) GetFloat(field string) (float64, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil {
		return 0, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	value = utils.RemoveCharacters(value, columnConfig.IgnoreCharacters)

	if columnConfig.DecimalSeparator != "." {
		value = strings.ReplaceAll(value, columnConfig.DecimalSeparator, ".")
	}

	return strconv.ParseFloat(value, 64)
}

func (record *CsvRecord) GetBoolean(field string) (bool, error) {
	value, fieldIndex, err := record.getField(field)
	if err != nil {
		return false, err
	}

	columnConfig := record.config.Columns[fieldIndex]
	if utils.IsInArray(value, columnConfig.TrueValues) {
		return true, nil
	}
	if utils.IsInArray(value, columnConfig.FalseValues) {
		return false, nil
	}

	return false, errors.New("The value '" + value + "' was found but is neither declared in trueValues or falseValues.")
}
