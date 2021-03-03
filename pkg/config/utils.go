package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"rods/pkg/util"
)

type ColumnType string

const (
	String  ColumnType = "string"
	Integer ColumnType = "integer"
	Float   ColumnType = "float"
	Boolean ColumnType = "boolean"
)

type validable interface {
	validate(log *logrus.Logger) error
}

func getAllNonNilFields(config interface{}) []validable {
	reflectConfig := reflect.ValueOf(config)
	if reflectConfig.Kind() == reflect.Ptr && !reflectConfig.IsNil() {
		reflectConfig = reflectConfig.Elem()
	}

	nonNilFields := make([]validable, 0)
	for fieldIndex := 0; fieldIndex < reflectConfig.NumField(); fieldIndex++ {
		reflectFieldIndex := reflectConfig.Field(fieldIndex)
		if !reflectFieldIndex.IsNil() {
			field := reflectFieldIndex.Interface().(validable)
			nonNilFields = append(nonNilFields, field)
		}
	}

	return nonNilFields
}

func checkDuplicateEndpointsPerService(outputConfigs map[string]Output) error {
	endpointsPerService := make(map[string]map[string]interface{})
	for _, outputConfigContainer := range outputConfigs {
		outputConfig := reflect.ValueOf(getAllNonNilFields(outputConfigContainer)[0])

		service := outputConfig.Elem().FieldByName("Service").String()
		endpoint := outputConfig.Elem().FieldByName("Endpoint").String()
		if service == "" || endpoint == "" {
			continue
		}

		serviceEndpoints, serviceExists := endpointsPerService[service]
		if !serviceExists {
			serviceEndpoints = make(map[string]interface{})
			endpointsPerService[service] = serviceEndpoints
		}

		if _, endpointExists := serviceEndpoints[endpoint]; endpointExists {
			return fmt.Errorf("Duplicate endpoint '%v' in service '%v'", endpoint, service)
		}

		serviceEndpoints[endpoint] = nil
	}

	return nil
}

func isValidColumnType(columnType ColumnType) bool {
	return util.IsInArray(
		string(columnType),
		[]string{
			string(String),
			string(Integer),
			string(Float),
			string(Boolean),
		},
	)
}
