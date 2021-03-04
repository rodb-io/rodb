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

func (record *Csv) Get(field string) (interface{}, error) {
	fieldIndex, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, fmt.Errorf("The column '%v' does not exist.", field)
	}

	if fieldIndex >= len(record.data) {
		return nil, nil
	}

	columnConfig := record.config.Columns[fieldIndex]
	value := record.data[fieldIndex]

	if columnConfig.Type == "string" {
		return util.PString(value), nil
	}
	if columnConfig.Type == "integer" {
		cleanedValue := util.RemoveCharacters(value, columnConfig.IgnoreCharacters)
		intValue, err := strconv.Atoi(cleanedValue)
		if err != nil {
			return nil, err
		}

		return util.PInt(intValue), nil
	}
	if columnConfig.Type == "float" {
		cleanedValue := util.RemoveCharacters(value, columnConfig.IgnoreCharacters)
		if columnConfig.DecimalSeparator != "." {
			cleanedValue = strings.ReplaceAll(cleanedValue, columnConfig.DecimalSeparator, ".")
		}

		floatValue, err := strconv.ParseFloat(cleanedValue, 64)
		if err != nil {
			return nil, err
		}

		return util.PFloat(floatValue), nil
	}
	if columnConfig.Type == "boolean" {
		if util.IsInArray(value, columnConfig.TrueValues) {
			return util.PBool(true), nil
		}
		if util.IsInArray(value, columnConfig.FalseValues) {
			return util.PBool(false), nil
		}

		return nil, fmt.Errorf("The value '%v' was found but is neither declared in trueValues or falseValues.", value)
	}

	return nil, fmt.Errorf("Unknown type '%v' in CsvRecord.Get", columnConfig.Type)
}

func (record *Csv) Position() Position {
	return record.position
}
