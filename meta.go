package mango

import (
	"fmt"
	"maps"
	"strings"
)

type (
	Types map[string]map[string]string
)

const (
	EmptyString = ""
)

func GetGoTypes(in TypedParam, parents []string) (Types, error) {
	schema, err := in.GetSchema()
	if err != nil {
		return nil, err
	}

	if schema == nil {
		return nil, nil
	}

	properties, err := schema.GetProperties()
	if err != nil {
		return nil, err
	}

	if properties == nil {
		return nil, nil
	}

	out := make(Types)
	current := lastOrDefault(parents)
	if out[current] == nil {
		out[current] = make(map[string]string)
	}

	for key, value := range properties {
		if value.IsArray() {
			res, err := GetGoArrayType(key, value, parents)
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
			continue
		}
		if value.IsObject() {
			res, err := GetGoTypes(value, append(parents, key))
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
		}
		typ, err := value.GetType()
		if err != nil {
			return nil, err
		}
		goType, err := GetGoType(typ, key)
		if err != nil {
			return nil, err
		}
		out[current][key] = goType
	}

	return out, nil
}

func GetGoArrayType(key string, value TypedParam, parents []string) (Types, error) {
	out := make(Types)
	current := lastOrDefault(parents)
	if out[current] == nil {
		out[current] = make(map[string]string)
	}
	innerSchema, err := value.GetSchema()
	if err != nil {
		return nil, err
	}
	items, err := innerSchema.GetItems()
	if err != nil {
		return nil, err
	}
	if len(items) > 1 {
		out[current][key] = "[]any"
		return out, nil
	}

	if !items[0].IsScalar() {
		res, err := GetGoTypes(items[0], append(parents, key))
		if err != nil {
			return nil, err
		}
		copyGoTypes(out, res)
	}

	typ, err := items[0].GetType()
	if err != nil {
		return nil, err
	}
	finalType, err := GetGoType(typ, key)
	if err != nil {
		return nil, err
	}
	out[current][key] = fmt.Sprintf("[]%s", finalType)
	return out, nil
}

func GetGoType(in []string, name string) (string, error) {
	if len(in) > 2 {
		return "any", nil
	}

	optional := EmptyString
	typ := EmptyString
	for _, i := range in {
		switch i {
		case "nil":
			{
				optional = "*"
			}
		case "object":
			{
				typ = name
			}
		default:
			{
				typ = i
			}
		}
	}

	if strings.TrimSpace(typ) == EmptyString {
		return EmptyString, fmt.Errorf("unexpectd type")
	}

	return fmt.Sprintf("%s%s", optional, typ), nil
}

func copyGoTypes(dest Types, src Types) {
	for key, value := range src {
		if _, ok := dest[key]; !ok {
			dest[key] = value
			continue
		}
		maps.Copy(dest[key], value)
	}
}

func lastOrDefault[T any](in []T) T {
	if len(in) == 0 {
		var zero T
		return zero
	}
	return in[len(in)-1]
}
