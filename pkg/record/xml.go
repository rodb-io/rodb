package record

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"rodb.io/pkg/config"
	parserModule "rodb.io/pkg/parser"
	"strconv"
	"strings"
)

type Xml struct {
	config        *config.XmlInput
	node          *xmlquery.Node
	parsers       parserModule.List
	nodeNavigator *xmlquery.NodeNavigator
	position      Position
}

func NewXml(
	config *config.XmlInput,
	node *xmlquery.Node,
	parsers parserModule.List,
	position Position,
) (*Xml, error) {
	return &Xml{
		config:        config,
		node:          node,
		parsers:       parsers,
		nodeNavigator: xmlquery.CreateXPathNavigator(node),
		position:      position,
	}, nil
}

func (record *Xml) All() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, property := range record.config.Properties {
		value, err := record.getAllValues(record.nodeNavigator, property)
		if err != nil {
			return nil, err
		}

		result[property.Name] = value
	}

	return result, nil
}

func (record *Xml) nodeIteratorToArray(
	nodeIterator *xpath.NodeIterator,
	config *config.XmlInputProperty,
) ([]interface{}, error) {
	values := make([]interface{}, 0)
	for {
		nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
		if nodeNavigator == nil {
			break
		}

		currentValue, err := record.getAllValues(nodeNavigator, config.Items)
		if err != nil {
			return nil, err
		}

		values = append(values, currentValue)

		if !nodeIterator.MoveNext() {
			break
		}
	}

	return values, nil
}

func (record *Xml) nodeIteratorToObject(
	nodeIterator *xpath.NodeIterator,
	config *config.XmlInputProperty,
) (interface{}, error) {
	nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
	if nodeNavigator == nil {
		return nil, nil
	}

	if nodeIterator.MoveNext() {
		return nil, record.xpathError(config, fmt.Sprintf("got multiple nodes, but the property has an object type"))
	}

	values := map[string]interface{}{}
	for _, property := range config.Properties {
		currentValue, err := record.getAllValues(nodeNavigator, property)
		if err != nil {
			return nil, err
		}

		values[property.Name] = currentValue
	}

	return values, nil
}

func (record *Xml) getAllValues(
	currentNode *xmlquery.NodeNavigator,
	currentConfig *config.XmlInputProperty,
) (interface{}, error) {
	result := currentConfig.CompiledXPath.Evaluate(currentNode)

	if nodeIterator, isNodeIterator := result.(*xpath.NodeIterator); isNodeIterator {
		if currentConfig.Type == config.XmlInputPropertyTypeArray {
			return record.nodeIteratorToArray(nodeIterator, currentConfig)
		} else if currentConfig.Type == config.XmlInputPropertyTypeObject {
			return record.nodeIteratorToObject(nodeIterator, currentConfig)
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a node list, but the property is nor an array or an object"))
		}
	}

	if currentConfig.Type != config.XmlInputPropertyTypePrimitive {
		return nil, record.xpathError(currentConfig, fmt.Sprintf("got a primitive value, but the property does not have a primitive type"))
	}

	if stringResult, resultIsString := result.(string); resultIsString {
		return record.handleStringValue(currentConfig, stringResult)
	}
	if floatResult, resultIsFloat := result.(float64); resultIsFloat {
		return record.handleNumericValue(currentConfig, floatResult)
	}
	if boolResult, resultIsBool := result.(bool); resultIsBool {
		return record.handleBoolValue(currentConfig, boolResult)
	}

	return nil, record.xpathError(currentConfig, fmt.Sprintf("returned an unexpected type: %#v", result))
}

func (record *Xml) Get(path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("Cannot get the property '%v' because it's path is empty.", path)
	}

	pathArray := strings.Split(path, ".")

	for _, property := range record.config.Properties {
		if property.Name == pathArray[0] {
			return record.getSubValue(record.nodeNavigator, property, pathArray[1:])
		}
	}

	return nil, fmt.Errorf("The path '%v' does not exist in this record.", path)
}

func (record *Xml) handleStringValue(config *config.XmlInputProperty, value string) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	return parser.Parse(value)
}

func (record *Xml) handleNumericValue(config *config.XmlInputProperty, value float64) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	if _, isFloatParser := parser.(*parserModule.Float); isFloatParser {
		return value, nil
	} else if _, isIntegerParser := parser.(*parserModule.Integer); isIntegerParser {
		return int(value), nil
	} else {
		return nil, record.xpathError(config, fmt.Sprintf("got a numeric value, but the property does not have a numeric parser"))
	}
}

func (record *Xml) handleBoolValue(config *config.XmlInputProperty, value bool) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	if _, isBooleanParser := parser.(*parserModule.Boolean); isBooleanParser {
		return value, nil
	} else {
		return nil, record.xpathError(config, fmt.Sprintf("got a boolean value, but the property does not have a boolean parser"))
	}
}

func (record *Xml) getSubArrayValue(
	nodeIterator *xpath.NodeIterator,
	config *config.XmlInputProperty,
	path []string,
) (interface{}, error) {
	if len(path) == 0 {
		// Getting the whole sub-array
		return record.nodeIteratorToArray(nodeIterator, config)
	}

	requestedIndex, err := strconv.Atoi(path[0])
	if err != nil {
		return nil, fmt.Errorf("Cannot get sub-path '%v' because the requested key is '%v', but the value is an array", path, path[0])
	}

	currentIndex := 0
	for {
		nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
		if nodeNavigator == nil {
			break
		}

		if currentIndex == requestedIndex {
			return record.getSubValue(nodeNavigator, config.Items, path[1:])
		}

		if !nodeIterator.MoveNext() {
			break
		}
		currentIndex++
	}

	// Not having the required index is a common case that
	// should not trigger an error, but get a nil value
	return nil, nil
}

func (record *Xml) getSubObjectValue(
	nodeIterator *xpath.NodeIterator,
	config *config.XmlInputProperty,
	path []string,
) (interface{}, error) {
	if len(path) == 0 {
		// Getting the whole sub-object
		return record.nodeIteratorToObject(nodeIterator, config)
	}

	nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
	if nodeNavigator == nil {
		return nil, nil
	}

	if nodeIterator.MoveNext() {
		return nil, record.xpathError(config, fmt.Sprintf("got multiple nodes, but the property has an object type"))
	}

	for _, property := range config.Properties {
		if property.Name == path[0] {
			return record.getSubValue(nodeNavigator, property, path[1:])
		}
	}

	// Not having some properties is a common case that
	// should not trigger an error, but get a nil value
	return nil, nil
}

func (record *Xml) getSubValue(
	currentNode *xmlquery.NodeNavigator,
	currentConfig *config.XmlInputProperty,
	path []string,
) (interface{}, error) {
	result := currentConfig.CompiledXPath.Evaluate(currentNode)

	// First, handling the array and object cases
	if nodeIterator, isNodeIterator := result.(*xpath.NodeIterator); isNodeIterator {
		if currentConfig.Type == config.XmlInputPropertyTypeArray {
			return record.getSubArrayValue(nodeIterator, currentConfig, path)
		} else if currentConfig.Type == config.XmlInputPropertyTypeObject {
			return record.getSubObjectValue(nodeIterator, currentConfig, path)
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a node list, but the property is nor an array or an object"))
		}
	}

	// From here, we only have to handle the primitive types returned by the xpath

	if currentConfig.Type != config.XmlInputPropertyTypePrimitive {
		return nil, record.xpathError(currentConfig, fmt.Sprintf("got a primitive value, but the property does not have a primitive parser"))
	}

	// Handling the string case first, because it can get parsed to non-primitive values
	if stringResult, resultIsString := result.(string); resultIsString {
		value, err := record.handleStringValue(currentConfig, stringResult)
		if err != nil {
			return nil, err
		}

		return getSubValue(value, path)
	}

	// Now we only have to handle the non-parseable primitive types

	if len(path) > 0 {
		return nil, fmt.Errorf("Cannot get sub-path '%v' because the value is primitive.", path)
	}

	if floatResult, resultIsFloat := result.(float64); resultIsFloat {
		return record.handleNumericValue(currentConfig, floatResult)
	}
	if boolResult, resultIsBool := result.(bool); resultIsBool {
		return record.handleBoolValue(currentConfig, boolResult)
	}

	return nil, record.xpathError(currentConfig, fmt.Sprintf("returned an unexpected type: %#v", result))
}

func (record *Xml) xpathError(config *config.XmlInputProperty, message string) error {
	return fmt.Errorf(
		"xpath '%v' for property '%v': %v",
		config.CompiledXPath.String(),
		config.Name,
		message,
	)
}

func (record *Xml) Position() Position {
	return record.position
}
