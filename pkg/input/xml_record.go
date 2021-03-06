package input

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	recordPackage "github.com/rodb-io/rodb/pkg/input/record"
	parserPackage "github.com/rodb-io/rodb/pkg/parser"
	"strconv"
	"strings"
)

type XmlRecord struct {
	config   *XmlConfig
	node     *xmlquery.Node
	parsers  parserPackage.List
	position recordPackage.Position
}

func NewXmlRecord(
	config *XmlConfig,
	node *xmlquery.Node,
	parsers parserPackage.List,
	position recordPackage.Position,
) (*XmlRecord, error) {
	return &XmlRecord{
		config:   config,
		node:     node,
		parsers:  parsers,
		position: position,
	}, nil
}

func (record *XmlRecord) All() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, property := range record.config.Properties {
		value, err := record.getAllValues(record.node, property)
		if err != nil {
			return nil, err
		}

		result[property.Name] = value
	}

	return result, nil
}

func (record *XmlRecord) nodeIteratorToArray(
	nodes []*xmlquery.Node,
	config *XmlPropertyConfig,
) ([]interface{}, error) {
	values := make([]interface{}, len(nodes))
	for i, node := range nodes {
		currentValue, err := record.getAllValues(node, config.Items)
		if err != nil {
			return nil, err
		}

		values[i] = currentValue
	}

	return values, nil
}

func (record *XmlRecord) nodeIteratorToObject(
	nodes []*xmlquery.Node,
	config *XmlPropertyConfig,
) (interface{}, error) {
	if len(nodes) == 0 {
		return nil, nil
	}
	if len(nodes) > 1 {
		return nil, record.xpathError(config, fmt.Sprintf("got multiple nodes, but the property has an object type"))
	}

	values := map[string]interface{}{}
	for _, property := range config.Properties {
		currentValue, err := record.getAllValues(nodes[0], property)
		if err != nil {
			return nil, err
		}

		values[property.Name] = currentValue
	}

	return values, nil
}

func (record *XmlRecord) getAllValues(
	node *xmlquery.Node,
	currentConfig *XmlPropertyConfig,
) (interface{}, error) {
	if currentConfig.Type == XmlInputPropertyTypeArray {
		nodes := xmlquery.QuerySelectorAll(node, currentConfig.CompiledXPath)
		return record.nodeIteratorToArray(nodes, currentConfig)
	} else if currentConfig.Type == XmlInputPropertyTypeObject {
		nodes := xmlquery.QuerySelectorAll(node, currentConfig.CompiledXPath)
		return record.nodeIteratorToObject(nodes, currentConfig)
	}

	result := currentConfig.CompiledXPath.Evaluate(
		xmlquery.CreateXPathNavigator(node),
	)

	if currentConfig.Type != XmlInputPropertyTypePrimitive {
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

	if nodeIterator, resultIsNodeIterator := result.(*xpath.NodeIterator); resultIsNodeIterator {
		return record.handleNodeIteratorValue(currentConfig, nodeIterator)
	}

	return nil, record.xpathError(currentConfig, fmt.Sprintf("returned an unexpected type: %#v", result))
}

// Converts a node iterator result into a string containing the raw XmlRecord,
// to facilitate debugging and setting-up the configuration
func (record *XmlRecord) handleNodeIteratorValue(config *XmlPropertyConfig, nodeIterator *xpath.NodeIterator) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	if _, parserIsString := parser.(*parserPackage.String); !parserIsString {
		return nil, record.xpathError(config, fmt.Sprintf("got XmlRecord node(s), but the property does not have the appropriate configuration or a string parser"))
	}

	result := ""
	for {
		if !nodeIterator.MoveNext() {
			break
		}

		nodeNavigator := nodeIterator.Current().(*xmlquery.NodeNavigator)
		if nodeNavigator == nil {
			continue
		}

		node := nodeNavigator.Current()
		if node == nil {
			continue
		}

		result = result + node.OutputXML(true)
	}

	return result, nil
}

func (record *XmlRecord) Get(path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("Cannot get the property '%v' because it's path is empty.", path)
	}

	pathArray := strings.Split(path, ".")

	for _, property := range record.config.Properties {
		if property.Name == pathArray[0] {
			return record.getSubValue(record.node, property, pathArray[1:])
		}
	}

	return nil, fmt.Errorf("The path '%v' does not exist in this record.", path)
}

func (record *XmlRecord) handleStringValue(config *XmlPropertyConfig, value string) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	return parser.Parse(value)
}

func (record *XmlRecord) handleNumericValue(config *XmlPropertyConfig, value float64) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	if _, isFloatParser := parser.(*parserPackage.Float); isFloatParser {
		return value, nil
	} else if _, isIntegerParser := parser.(*parserPackage.Integer); isIntegerParser {
		return int64(value), nil
	} else {
		return nil, record.xpathError(config, fmt.Sprintf("got a numeric value, but the property does not have a numeric parser"))
	}
}

func (record *XmlRecord) handleBoolValue(config *XmlPropertyConfig, value bool) (interface{}, error) {
	parser, parserExists := record.parsers[config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser '%v' was not found.", config.Parser)
	}

	if _, isBooleanParser := parser.(*parserPackage.Boolean); isBooleanParser {
		return value, nil
	} else {
		return nil, record.xpathError(config, fmt.Sprintf("got a boolean value, but the property does not have a boolean parser"))
	}
}

func (record *XmlRecord) getSubArrayValue(
	nodes []*xmlquery.Node,
	config *XmlPropertyConfig,
	path []string,
) (interface{}, error) {
	if len(path) == 0 {
		// Getting the whole sub-array
		return record.nodeIteratorToArray(nodes, config)
	}

	requestedIndex, err := strconv.Atoi(path[0])
	if err != nil {
		return nil, fmt.Errorf("Cannot get sub-path '%v' because the requested key is '%v', but the value is an array", path, path[0])
	}

	if requestedIndex >= len(nodes) {
		// Not having the required index is a common case that
		// should not trigger an error, but get a nil value
		return nil, nil
	}

	return record.getSubValue(nodes[requestedIndex], config.Items, path[1:])
}

func (record *XmlRecord) getSubObjectValue(
	nodes []*xmlquery.Node,
	config *XmlPropertyConfig,
	path []string,
) (interface{}, error) {
	if len(path) == 0 {
		// Getting the whole sub-object
		return record.nodeIteratorToObject(nodes, config)
	}

	if len(nodes) == 0 {
		return nil, nil
	}
	if len(nodes) > 1 {
		return nil, record.xpathError(config, fmt.Sprintf("got multiple nodes, but the property has an object type"))
	}

	for _, property := range config.Properties {
		if property.Name == path[0] {
			return record.getSubValue(nodes[0], property, path[1:])
		}
	}

	// Not having some properties is a common case that
	// should not trigger an error, but get a nil value
	return nil, nil
}

func (record *XmlRecord) getSubValue(
	node *xmlquery.Node,
	currentConfig *XmlPropertyConfig,
	path []string,
) (interface{}, error) {
	if currentConfig.Type == XmlInputPropertyTypeArray {
		nodes := xmlquery.QuerySelectorAll(node, currentConfig.CompiledXPath)
		return record.getSubArrayValue(nodes, currentConfig, path)
	} else if currentConfig.Type == XmlInputPropertyTypeObject {
		nodes := xmlquery.QuerySelectorAll(node, currentConfig.CompiledXPath)
		return record.getSubObjectValue(nodes, currentConfig, path)
	}

	result := currentConfig.CompiledXPath.Evaluate(
		xmlquery.CreateXPathNavigator(node),
	)

	// From here, we only have to handle the primitive types returned by the xpath

	if currentConfig.Type != XmlInputPropertyTypePrimitive {
		return nil, record.xpathError(currentConfig, fmt.Sprintf("got a primitive value, but the property does not have a primitive parser"))
	}

	// Handling the string case first, because it can get parsed to non-primitive values
	if stringResult, resultIsString := result.(string); resultIsString {
		value, err := record.handleStringValue(currentConfig, stringResult)
		if err != nil {
			return nil, err
		}

		return recordPackage.GetSubValue(value, path)
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
	if nodeIterator, resultIsNodeIterator := result.(*xpath.NodeIterator); resultIsNodeIterator {
		return record.handleNodeIteratorValue(currentConfig, nodeIterator)
	}

	return nil, record.xpathError(currentConfig, fmt.Sprintf("returned an unexpected type: %#v", result))
}

func (record *XmlRecord) xpathError(config *XmlPropertyConfig, message string) error {
	return fmt.Errorf(
		"xpath '%v' for property '%v': %v",
		config.CompiledXPath.String(),
		config.Name,
		message,
	)
}

func (record *XmlRecord) Position() recordPackage.Position {
	return record.position
}
