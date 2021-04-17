package record

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"rodb.io/pkg/config"
	parserModule "rodb.io/pkg/parser"
)

type Xml struct {
	config          *config.XmlInput
	propertyParsers []parserModule.Parser
	node            *xmlquery.Node
	nodeNavigator   *xmlquery.NodeNavigator
	position        Position
}

func NewXml(
	config *config.XmlInput,
	propertyParsers []parserModule.Parser,
	node *xmlquery.Node,
	position Position,
) (*Xml, error) {
	return &Xml{
		config:          config,
		propertyParsers: propertyParsers,
		node:            node,
		nodeNavigator:   xmlquery.CreateXPathNavigator(node),
		position:        position,
	}, nil
}

func (record *Xml) All() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, property := range record.config.Properties {
		value, err := record.Get(property.Name)
		if err != nil {
			return nil, err
		}

		result[property.Name] = value
	}

	return result, nil
}

func (record *Xml) Get(field string) (interface{}, error) {
	fieldIndex, exists := record.config.PropertyIndexByName[field]
	if !exists {
		return nil, fmt.Errorf("The property '%v' does not exist.", field)
	}

	if fieldIndex >= len(record.config.Properties) {
		return nil, nil
	}

	parser := record.propertyParsers[fieldIndex]
	propertyConfig := record.config.Properties[fieldIndex]

	result := propertyConfig.CompiledXPath.Evaluate(record.nodeNavigator)
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
				"The xpath '%v' for property '%v' returned a numeric value, but the property does not have a numeric type.",
				propertyConfig.XPath,
				propertyConfig.Name,
			)
		}
	case bool:
		if _, isBooleanParser := parser.(*parserModule.Boolean); isBooleanParser {
			return result, nil
		} else {
			return nil, fmt.Errorf(
				"The xpath '%v' for property '%v' returned a boolean value, but the property does not have a boolean type.",
				propertyConfig.XPath,
				propertyConfig.Name,
			)
		}
	default:
		return nil, fmt.Errorf(
			"The xpath '%v' for property '%v' did not return a primitive type.",
			propertyConfig.XPath,
			propertyConfig.Name,
		)
	}
}

func (record *Xml) Position() Position {
	return record.position
}
