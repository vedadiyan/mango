package generators

import (
	"bytes"
	"fmt"
	"maps"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/vedadiyan/mango/internal/codegen"
)

type (
	GoBasicType  string
	GoStructType string
	GoArrayType  string

	Types map[string]map[string]string

	CombinatorType string

	GoTypeOptions struct {
		optional string
	}

	GoTypeOption func(*GoTypeOptions)
)

const (
	EmptyString = ""

	CombinatorTypeInavlid CombinatorType = "Invalid"
	CombinatorTypeOneOf   CombinatorType = "OneOf"
	CombinatorTypeAnyOf   CombinatorType = "AnyOf"
	CombinatorTypeAllOf   CombinatorType = "AllOf"
	CombinatorTypeNot     CombinatorType = "Not"
)

func GetGoTypes(in codegen.TypedParam, parents []string) (Types, error) {
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
		typeName := key
		if value.IsArray() {
			res, err := GetGoArrayType(1, key, value, parents)
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
			continue
		}
		if value.IsCombinator() {
			res, err := GetGoCombinatorType(key, value, parents)
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
			typeName = flaten(append(parents, key))
		}

		typ, err := value.GetType()
		if err != nil {
			return nil, err
		}

		if typ == nil {
			enum, err := value.GetEnum()
			if err != nil {
				return nil, err
			}

			typ = []string{flaten([]string{current, key})}
			enumType := flaten([]string{"$", current, key})
			if _, ok := out[enumType]; !ok {
				out[enumType] = make(map[string]string)
			}
			for _, value := range enum {
				val := fmt.Sprintf("%v", value)
				out[enumType][flaten([]string{current, key, val})] = val
			}
		}

		required, err := value.GetRequiredOrFalse()
		if err != nil {
			return nil, err
		}
		goType, err := GetGoType(typ, typeName, WithRequired(required))
		if err != nil {
			return nil, err
		}
		if _, ok := out[current]; !ok {
			out[current] = make(map[string]string)
		}
		out[current][key] = goType
	}

	return out, nil
}

func GetGoArrayType(dim int, key string, value codegen.TypedParam, parents []string) (Types, error) {
	out := make(Types)
	current := flaten(parents)
	innerSchema, err := value.GetSchema()
	if err != nil {
		return nil, err
	}
	items, err := innerSchema.GetItemsAsArray()
	if err != nil {
		return nil, err
	}

	if _, ok := out[current]; !ok {
		out[current] = make(map[string]string)
	}

	if len(items) > 1 {
		out[current][key] = "[]any"
		return out, nil
	}

	if items[0].IsObject() {
		res, err := GetGoTypes(items[0], append(parents, key))
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

	finalType, err := GetGoType(typ, flaten(append(parents, key)))
	if err != nil {
		return nil, err
	}
	out[current][key] = fmt.Sprintf("%s%s", strings.Repeat("[]", dim), finalType)
	return out, nil
}

func GetGoCombinatorType(key string, value codegen.TypedParam, parents []string) (Types, error) {
	combinatorType, ok := GetCombinatorType(value.TypeName)
	if !ok {
		return nil, fmt.Errorf("invalid combinator")
	}

	out := make(Types)
	current := flaten(parents)
	schema, err := value.GetSchema()
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, nil
	}

	if _, ok := out[current]; !ok {
		out[current] = make(map[string]string)
	}

	specs, err := schema.GetSpecs()
	if err != nil {
		return nil, err
	}
	typeName := fmt.Sprintf("%s%s", combinatorType, key)
	combinatorKey := flaten(append(parents, typeName))
	out[combinatorKey] = make(map[string]string)

	for i, spec := range specs {
		typ, err := spec.GetType()
		if err != nil {
			return nil, err
		}
		combinatorName := fmt.Sprintf("%sVariation%d", key, i)
		innerSchema, err := spec.GetSchema()
		if err != nil {
			return nil, nil
		}
		if innerSchema == nil {
			continue
		}
		objectName, err := innerSchema.GetObjectName()
		if err != nil {
			return nil, err
		}
		if objectName != nil {
			combinatorName = MakePascalCase(*objectName)
		}
		flattenedKey := flaten(append(parents, typeName, combinatorName))

		if spec.IsArray() {
			res, err := GetGoArrayType(0, combinatorName, spec, append(parents, typeName))
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
			continue
		}
		if spec.IsObject() {
			res, err := GetGoTypes(spec, append(parents, typeName, combinatorName))
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
		}

		required, err := spec.GetRequiredOrFalse()
		if err != nil {
			return nil, err
		}

		finalType, err := GetGoType(typ, flattenedKey, WithRequired(required))
		if err != nil {
			return nil, err
		}
		out[combinatorKey][combinatorName] = finalType
	}
	out[current][key] = combinatorKey
	return out, nil
}

func WithRequired(required bool) GoTypeOption {
	return func(gto *GoTypeOptions) {
		if !required {
			gto.optional = "*"
		}
	}
}

func GetGoType(in []string, name string, opts ...GoTypeOption) (string, error) {
	if len(in) > 2 {
		return "any", nil
	}

	options := GoTypeOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	typ := EmptyString
	for _, i := range in {
		switch i {
		case "null":
			{
				options.optional = "*"
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
		return "any", nil
	}

	return fmt.Sprintf("%s%s", options.optional, typ), nil
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

func flaten[T any](in []T) string {
	out := bytes.NewBufferString("")

	for _, value := range in {
		fmt.Fprintf(out, "%v", value)

	}

	return out.String()
}

func MakePascalCase(str string) string {
	return strcase.ToCamel(str)
}

func GetCombinatorType(in string) (CombinatorType, bool) {
	_, str, _ := strings.CutLast(strings.TrimFunc(in, func(r rune) bool { return r == '[' || r == ']' }), ".")
	typ := CombinatorType(str)
	switch typ {
	case CombinatorTypeOneOf, CombinatorTypeAnyOf, CombinatorTypeAllOf, CombinatorTypeNot:
		{
			return typ, true
		}
	default:
		{
			return CombinatorType(""), false
		}
	}
}
