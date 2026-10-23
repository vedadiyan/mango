package tools

import (
	"reflect"
	"strings"
)

func FieldName[T any]() string {
	v := reflect.TypeFor[T]()

	if v.Kind() != reflect.Struct || v.NumField() != 1 {
		return "-"
	}

	field := v.Field(0)

	if value, ok := field.Tag.Lookup("bson"); ok {
		if name, _, _ := strings.Cut(value, ","); name != "" {
			return name
		}
	}

	return field.Name
}
