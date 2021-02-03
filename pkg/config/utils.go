package config

import (
	"reflect"
)

type validable interface{
	validate() error
}

func getAllNonNilFields(config interface{}) []validable {
	reflectConfig := reflect.ValueOf(config)
    nonNilFields := make([]validable, 0)
	for fieldIndex := 0; fieldIndex < reflectConfig.NumField(); fieldIndex++ {
		field := reflectConfig.Field(fieldIndex).Interface().(*validable)
		if field != nil {
			nonNilFields = append(nonNilFields, *field)
		}
	}

	return nonNilFields
}
