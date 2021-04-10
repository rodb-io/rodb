package record

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"rods/pkg/config"
	parserModule "rods/pkg/parser"
)

type Xml struct {
	config        *config.XmlInput
	columnParsers []parserModule.Parser
	node          *xmlquery.Node
	nodeNavigator *xmlquery.NodeNavigator
	position      Position
}

func NewXml(
	config *config.XmlInput,
	columnParsers []parserModule.Parser,
	node *xmlquery.Node,
	position Position,
) (*Xml, error) {
	return &Xml{
		config:        config,
		columnParsers: columnParsers,
		node:          node,
		nodeNavigator: xmlquery.CreateXPathNavigator(node),
		position:      position,
	}, nil
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

func (record *Xml) Get(field string) (interface{}, error) {
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
