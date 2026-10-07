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
		}
		typ, err := value.GetType()
		if err != nil {
			return nil, err
		}
		goType, err := GetGoType(typ, key)
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
	items, err := innerSchema.GetItems()
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
	finalType, err := GetGoType(typ, key)
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
	types := make([]string, 0)
	for i, spec := range specs {
		typ, err := spec.GetType()
		if err != nil {
			return nil, err
		}
		combinatorName := fmt.Sprintf("(%s)%sVariation%d", combinatorType, key, i)
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
			combinatorName = fmt.Sprintf("(%s)%s", combinatorType, MakePascalCase(*objectName))
		}
		if !spec.IsScalar() {
			res, err := GetGoTypes(spec, append(parents, key, combinatorName))
			if err != nil {
				return nil, err
			}
			copyGoTypes(out, res)
		} else {
			finalType, err := GetGoType(typ, key)
			if err != nil {
				return nil, err
			}
			out[current][combinatorName] = finalType
		}

		types = append(types, typ...)
	}
	finalType, err := GetGoType(types, key)
	if err != nil {
		return nil, err
	}
	out[current][key] = finalType
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
		return "any", nil
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

func flaten[T any](in []T) string {
	out := bytes.NewBufferString("")

	for i, value := range in {
		if i > 0 {
			out.WriteRune('.')
		}
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
