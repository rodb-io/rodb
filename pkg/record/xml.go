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
	currentConfig *config.XmlInputProperty,
) ([]interface{}, error) {
	values := make([]interface{}, 0)
	for {
		nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
		if nodeNavigator == nil {
			break
		}

		currentValue, err := record.getAllValues(nodeNavigator, currentConfig.Items)
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
	currentConfig *config.XmlInputProperty,
) (interface{}, error) {
	nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
	if nodeNavigator == nil {
		return nil, nil
	}

	if nodeIterator.MoveNext() {
		return nil, record.xpathError(currentConfig, fmt.Sprintf("got multiple nodes, but the property has an object type"))
	}

	values := map[string]interface{}{}
	for _, property := range currentConfig.Properties {
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

	// First, handling the array and object cases
	if nodeIterator, isNodeIterator := result.(*xpath.NodeIterator); isNodeIterator {
		if currentConfig.Type == config.XmlInputPropertyTypeArray {
			return record.nodeIteratorToArray(nodeIterator, currentConfig)
		} else if currentConfig.Type == config.XmlInputPropertyTypeObject {
			return record.nodeIteratorToObject(nodeIterator, currentConfig)
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a node list, but the property is nor an array or an object"))
		}
	}

	// From here, we only have to handle the primitive types returned by the xpath

	if currentConfig.Type != config.XmlInputPropertyTypePrimitive {
		return nil, record.xpathError(currentConfig, fmt.Sprintf("got a primitive value, but the property does not have a primitive type"))
	}

	parser, parserExists := record.parsers[currentConfig.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", currentConfig.Parser)
	}

	if stringResult, resultIsString := result.(string); resultIsString {
		return parser.Parse(stringResult)
	}

	// Now we only have to handle the non-parseable primitive types

	if floatResult, resultIsFloat := result.(float64); resultIsFloat {
		if _, isFloatParser := parser.(*parserModule.Float); isFloatParser {
			return floatResult, nil
		} else if _, isIntegerParser := parser.(*parserModule.Integer); isIntegerParser {
			return int(floatResult), nil
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a numeric value, but the property does not have a numeric parser"))
		}
	}

	if boolResult, resultIsBool := result.(bool); resultIsBool {
		if _, isBooleanParser := parser.(*parserModule.Boolean); isBooleanParser {
			return boolResult, nil
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a boolean value, but the property does not have a boolean parser"))
		}
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

func (record *Xml) getSubValue(
	currentNode *xmlquery.NodeNavigator,
	currentConfig *config.XmlInputProperty,
	path []string,
) (interface{}, error) {
	result := currentConfig.CompiledXPath.Evaluate(currentNode)

	// First, handling the array and object cases
	if nodeIterator, isNodeIterator := result.(*xpath.NodeIterator); isNodeIterator {
		if currentConfig.Type == config.XmlInputPropertyTypeArray {
			if len(path) == 0 {
				// Getting the whole sub-array
				return record.nodeIteratorToArray(nodeIterator, currentConfig)
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
					return record.getSubValue(nodeNavigator, currentConfig.Items, path[1:])
				}

				if !nodeIterator.MoveNext() {
					break
				}
				currentIndex++
			}

			// Not having the required index is a common case that
			// should not trigger an error, but get a nil value
			return nil, nil
		} else if currentConfig.Type == config.XmlInputPropertyTypeObject {
			if len(path) == 0 {
				// Getting the whole sub-object
				return record.nodeIteratorToObject(nodeIterator, currentConfig)
			}

			nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
			if nodeNavigator == nil {
				return nil, nil
			}

			if nodeIterator.MoveNext() {
				return nil, record.xpathError(currentConfig, fmt.Sprintf("got multiple nodes, but the property has an object type"))
			}

			for _, property := range currentConfig.Properties {
				if property.Name == path[0] {
					return record.getSubValue(nodeNavigator, property, path[1:])
				}
			}

			// Not having some properties is a common case that
			// should not trigger an error, but get a nil value
			return nil, nil
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a node list, but the property is nor an array or an object"))
		}
	}

	// From here, we only have to handle the primitive types returned by the xpath

	if currentConfig.Type != config.XmlInputPropertyTypePrimitive {
		return nil, record.xpathError(currentConfig, fmt.Sprintf("got a primitive value, but the property does not have a primitive parser"))
	}

	parser, parserExists := record.parsers[currentConfig.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", currentConfig.Parser)
	}

	// Handling the string case first, because it can get parsed to non-primitive values
	if stringResult, resultIsString := result.(string); resultIsString {
		value, err := parser.Parse(stringResult)
		if err != nil {
			return nil, err
		}

		if len(path) == 0 {
			return value, nil
		} else if parser.Primitive() {
			return nil, fmt.Errorf("Cannot get sub-path '%v' because the value is a primitive.", path)
		} else {
			return getSubValue(value, path)
		}
	}

	// Now we only have to handle the non-parseable primitive types

	if len(path) > 0 {
		return nil, fmt.Errorf("Cannot get sub-path '%v' because the value is numeric.", path)
	}

	if floatResult, resultIsFloat := result.(float64); resultIsFloat {
		if _, isFloatParser := parser.(*parserModule.Float); isFloatParser {
			return floatResult, nil
		} else if _, isIntegerParser := parser.(*parserModule.Integer); isIntegerParser {
			return int(floatResult), nil
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a numeric value, but the property does not have a numeric parser"))
		}
	}

	if boolResult, resultIsBool := result.(bool); resultIsBool {
		if _, isBooleanParser := parser.(*parserModule.Boolean); isBooleanParser {
			return boolResult, nil
		} else {
			return nil, record.xpathError(currentConfig, fmt.Sprintf("got a boolean value, but the property does not have a boolean parser"))
		}
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
