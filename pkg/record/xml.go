package record

import (
	"bytes"
	"fmt"
	"github.com/antchfx/xmlquery"
	"rods/pkg/config"
	parserModule "rods/pkg/parser"
)

type Xml struct {
	config        *config.XmlInput
	columnParsers []parserModule.Parser
	data          []byte
	node          *xmlquery.Node
	nodeNavigator *xmlquery.NodeNavigator
	position      Position
}

func NewXml(
	config *config.XmlInput,
	columnParsers []parserModule.Parser,
	data []byte,
	position Position,
) *Xml {
	return &Xml{
		config:        config,
		columnParsers: columnParsers,
		data:          data,
		node:          nil, // Dynamically loaded
		nodeNavigator: nil, // Dynamically loaded
		position:      position,
	}
}

func (record *Xml) All() (map[string]interface{}, error) {
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

func (record *Xml) parseData() error {
	var err error
	reader := bytes.NewReader(record.data)
	record.node, err = xmlquery.Parse(reader)
	if err != nil {
		return err
	}

	record.nodeNavigator = xmlquery.CreateXPathNavigator(record.node)

	return nil
}

func (record *Xml) Get(field string) (interface{}, error) {
	// Initializing the document only when actually needed
	if record.node == nil {
		err := record.parseData()
		if err != nil {
			return nil, err
		}
	}

	fieldIndex, exists := record.config.ColumnIndexByName[field]
	if !exists {
		return nil, fmt.Errorf("The column '%v' does not exist.", field)
	}

	if fieldIndex >= len(record.config.Columns) {
		return nil, nil
	}

	parser := record.columnParsers[fieldIndex]
	columnConfig := record.config.Columns[fieldIndex]

	result := columnConfig.CompiledXPath.Evaluate(record.nodeNavigator)
	switch result.(type) {
	case string:
		value, err := parser.Parse(result.(string))
		if err != nil {
			return nil, err
		}

		return value, nil
	case float64:
		if _, isFloatParser := parser.(*parserModule.Float); isFloatParser {
			return result, nil
		} else if _, isIntegerParser := parser.(*parserModule.Integer); isIntegerParser {
			return int(result.(float64)), nil
		} else {
			return nil, fmt.Errorf(
				"The xpath '%v' for column '%v' returned a numeric value, but the column does not have a numeric type.",
				columnConfig.XPath,
				columnConfig.Name,
			)
		}
	case bool:
		if _, isBooleanParser := parser.(*parserModule.Boolean); isBooleanParser {
			return result, nil
		} else {
			return nil, fmt.Errorf(
				"The xpath '%v' for column '%v' returned a boolean value, but the column does not have a boolean type.",
				columnConfig.XPath,
				columnConfig.Name,
			)
		}
	default:
		return nil, fmt.Errorf(
			"The xpath '%v' for column '%v' did not return a primitive type.",
			columnConfig.XPath,
			columnConfig.Name,
		)
	}
}

func (record *Xml) Position() Position {
	return record.position
}
