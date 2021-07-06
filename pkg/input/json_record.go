package input

import (
	"fmt"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"strconv"
	"strings"
)

type JsonRecord struct {
	config        *JsonConfig
	columnParsers []parser.Parser
	data          map[string]interface{}
	position      record.Position
}

func NewJsonRecord(
	config *JsonConfig,
	data map[string]interface{},
	position record.Position,
) *JsonRecord {
	return &JsonRecord{
		config:   config,
		data:     data,
		position: position,
	}
}

func (record *JsonRecord) All() (map[string]interface{}, error) {
	return record.data, nil
}

func (record *JsonRecord) Get(path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("Cannot get the property '%v' because it's path is empty.", path)
	}

	pathArray := strings.Split(path, ".")

	return record.getSubValue(record.data, pathArray)
}

func (record *JsonRecord) getSubValue(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}

	if dataMap, dataIsMap := data.(map[string]interface{}); dataIsMap {
		property, propertyExists := dataMap[path[0]]
		if !propertyExists {
			// Property does not exist: return nil (we must not output
			// an error in this case, even if there is a sub-path)
			return nil, nil
		}

		return record.getSubValue(property, path[1:])
	} else if dataArray, dataIsArray := data.([]interface{}); dataIsArray {
		indexInPath, err := strconv.Atoi(path[0])
		if err != nil {
			return nil, fmt.Errorf("Cannot get path '%v' because the value is an array, but the index is non-numeric: %w", path, err)
		}

		// Index does not exist: return nil (we must not output
		// an error in this case, even if there is a sub-path)
		if indexInPath >= len(dataArray) {
			return nil, nil
		}

		return record.getSubValue(dataArray[indexInPath], path[1:])
	} else {
		return nil, fmt.Errorf("Cannot get path '%v' because the value is primitive", path)
	}
}

func (record *JsonRecord) Position() record.Position {
	return record.position
}
