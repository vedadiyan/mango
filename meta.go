package mango

import (
	"bytes"
	"fmt"
	"maps"
	"strings"

	"github.com/iancoleman/strcase"
)

type (
	GoBasicType  string
	GoStructType string
	GoArrayType  string

	Types map[string]any
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
	current := flaten(parents)

	for key, value := range properties {
		key = MakePascalCase(key)
		if value.IsArray() {
			res, err := GetGoArrayType(1, key, value, append(parents, key))
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
			continue
		}
		// if value.IsCombinator() {
		// 	res, err := GetGoCombinatorType(key, value, append(parents, key))
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	copyGoTypes(out, res)
		// 	continue
		// }
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
		out[fmt.Sprintf("%s.%s", current, key)] = goType
	}

	return out, nil
}

func GetGoArrayType(dim int, key string, value TypedParam, parents []string) (Types, error) {
	out := make(Types)
	current := flaten(parents)
	innerSchema, err := value.GetSchema()
	if err != nil {
		return nil, err
	}
	items, err := innerSchema.GetItems()
	if err != nil {
		return nil, err
	}
	if len(items) > 1 {
		out[current] = "[]any"
		return out, nil
	}

	if items[0].IsObject() {
		res, err := GetGoTypes(items[0], parents)
		if err != nil {
			return nil, err
		}
		copyGoTypes(out, res)
	}

	if items[0].IsArray() {
		res, err := GetGoArrayType(dim+1, key, items[0], parents)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	typ, err := items[0].GetType()
	if err != nil {
		return nil, err
	}
	finalType, err := GetGoType(typ, key)
	if err != nil {
		return nil, err
	}
	out[current] = fmt.Sprintf("%s%s", strings.Repeat("[]", dim), finalType)
	return out, nil
}

func GetGoCombinatorType(key string, value TypedParam, parents []string) (Types, error) {
	out := make(Types)
	current := flaten(parents)
	schema, err := value.GetSchema()
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, nil
	}
	specs, err := schema.GetSpecs()
	if err != nil {
		return nil, err
	}
	types := make([]string, 0)
	for i, spec := range specs {
		typ, err := spec.GetType()
		if err != nil {
			return nil, err
		}

		if !spec.IsScalar() {
			res, err := GetGoTypes(spec, append(parents, fmt.Sprintf("$%sVariation%d", key, i)))
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
		} else {
			finalType, err := GetGoType(typ, key)
			if err != nil {
				return nil, err
			}
			out[fmt.Sprintf("$%sVariation%d", key, i)] = finalType
		}

		types = append(types, typ...)
	}
	finalType, err := GetGoType(types, key)
	if err != nil {
		return nil, err
	}
	out[fmt.Sprintf("%s.%s", current, key)] = finalType
	return out, nil
}

func GetGoType(in []string, name string) (any, error) {
	if len(in) > 2 {
		return "any", nil
	}

	optional := EmptyString
	typ := EmptyString
	for _, i := range in {
		switch i {
		case "null":
			{
				optional = "*"
			}
		case "object":
			{
				typ = name
			}
		case "binData":
			{
				typ = "[]byte"
			}
		default:
			{
				typ = i
			}
		}
	}

	if strings.TrimSpace(typ) == EmptyString {
		return GoBasicType("any"), nil
	}

	return fmt.Sprintf("%s%s", optional, typ), nil
}

func copyGoTypes(dest Types, src Types) {
	maps.Copy(dest, src)
}

func flaten[T any](in []T) string {
	out := bytes.NewBufferString("")

	for i, value := range in {
		if i > 0 {
			out.WriteRune('.')
		}
		out.WriteString(fmt.Sprintf("%v", value))

	}

	return out.String()
}

func MakePascalCase(str string) string {
	return strcase.ToCamel(str)
}
