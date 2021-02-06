package config

import (
	"errors"
	"reflect"
	"github.com/sirupsen/logrus"
)

type validable interface{
	validate(log *logrus.Logger) error
}

func getAllNonNilFields(config interface{}) []validable {
	reflectConfig := reflect.ValueOf(config).Elem()
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

func checkDuplicateEndpointsPerService(outputConfigs map[string]OutputConfig) error {
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
			return errors.New("Duplicate endpoint '" + endpoint + "' in service '" + service + "'")
		}

		serviceEndpoints[endpoint] = nil
	}

	return nil
}

func isCsvInputColumnTypeValid(typeToCheck string) bool {
	types := []string {"string", "integer", "float", "boolean"}
	for _, definedType := range types {
		if definedType == typeToCheck {
			return true
		}
	}

	return false
}
