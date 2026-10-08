package generators

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/google/uuid"
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
	EmptyString   = ""
	ZeroDimension = 0

	CombinatorTypeInavlid CombinatorType = "Invalid"
	CombinatorTypeOneOf   CombinatorType = "OneOf"
	CombinatorTypeAnyOf   CombinatorType = "AnyOf"
	CombinatorTypeAllOf   CombinatorType = "AllOf"
	CombinatorTypeNot     CombinatorType = "Not"
)

func ToGoTypeModel(in codegen.TypedParam, parents []string, forceOptional bool) (Types, error) {
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

	oneOfs, err := in.GetOneOfs()
	if err != nil {
		return nil, err
	}

	requiredList, err := schema.GetRequiredStringArray()
	if err != nil {
		return nil, err
	}

	out := make(Types)
	current := flaten(parents)

	for key, value := range properties {
		pascalCaseKey := MakePascalCase(key)
		typeName := pascalCaseKey
		if value.IsArray() {
			res, err := ToGoArrayType(1, pascalCaseKey, value, parents)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			continue
		}
		if value.IsCombinator() {
			res, err := ToGoCombinatorType(pascalCaseKey, value, parents)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			continue
		}
		if value.IsObject() {
			res, err := ToGoTypeModel(value, append(parents, pascalCaseKey), forceOptional)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			typeName = flaten(append(parents, pascalCaseKey))
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

			typ = []string{flaten([]string{current, pascalCaseKey})}
			enumType := flaten([]string{"$", current, pascalCaseKey})
			if _, ok := out[enumType]; !ok {
				out[enumType] = make(map[string]string)
			}
			for _, value := range enum {
				val := fmt.Sprintf("%v", value)
				out[enumType][flaten([]string{current, pascalCaseKey, val})] = val
			}
		}

		required, err := value.GetRequiredOrFalse()
		if err != nil {
			return nil, err
		}
		if slices.Contains(requiredList, key) {
			required = true
		}
		goType, err := GetGoType(typ, typeName, WithRequired(required && !forceOptional))
		if err != nil {
			return nil, err
		}
		if _, ok := out[current]; !ok {
			out[current] = make(map[string]string)
		}
		out[current][pascalCaseKey] = goType
	}

	if oneOfs != nil {
		res, err := ToGoCombinatorType("", *oneOfs, parents)
		if err != nil {
			return nil, err
		}
		deepCopy(out, res)
	}

	return out, nil
}

func ToGoArrayType(dim int, key string, value codegen.TypedParam, parents []string) (Types, error) {
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
		res, err := ToGoTypeModel(items[0], append(parents, key), false)
		if err != nil {
			return nil, err
		}
		deepCopy(out, res)
	}

	if items[0].IsArray() {
		res, err := ToGoArrayType(dim+1, key, items[0], parents)
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

func ToGoCombinatorType(key string, value codegen.TypedParam, parents []string) (Types, error) {
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

	identKey := key
	isDynamicIdent := false
	if strings.TrimSpace(key) == EmptyString {
		identKey = uuid.NewString()
		isDynamicIdent = true
	}

	parentIdentKey := identKey
	if len(parents) > 0 {
		parentIdentKey = flaten(parents)
	}

	currentIdentSlice := append(parents, identKey)
	currentIdentKey := flaten(append(parents, identKey))

	out := make(Types)
	out[parentIdentKey] = make(map[string]string)
	out[currentIdentKey] = make(map[string]string)

	for _, spec := range specs {
		innerSchema, err := spec.GetSchema()
		if err != nil {
			return nil, nil
		}

		if innerSchema == nil {
			continue
		}

		explicitObjectName, err := innerSchema.GetObjectName()
		if err != nil {
			return nil, err
		}

		innerIdentKey := ""
		hasInnerIdentKey := false

		if explicitObjectName != nil {
			innerIdentKey = MakePascalCase(*explicitObjectName)
			hasInnerIdentKey = true
		}

		innerIdentSlice := append(currentIdentSlice, innerIdentKey)

		if spec.IsArray() {
			res, err := ToGoArrayType(ZeroDimension, innerIdentKey, spec, currentIdentSlice)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			continue
		}

		if spec.IsObject() {
			res, err := ToGoTypeModel(spec, innerIdentSlice, hasInnerIdentKey)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
		}

		if innerIdentKey == EmptyString {
			continue
		}

		explicitType, err := spec.GetType()
		if err != nil {
			return nil, err
		}
		explicitRequired, err := spec.GetRequiredOrFalse()
		if err != nil {
			return nil, err
		}
		goType, err := GetGoType(explicitType, flaten(innerIdentSlice), WithRequired(explicitRequired))
		if err != nil {
			return nil, err
		}
		out[currentIdentKey][innerIdentKey] = goType
	}

	if !isDynamicIdent {
		out[parentIdentKey][key] = currentIdentKey
		return out, nil
	}

	out[parentIdentKey] = out[currentIdentKey]
	delete(out, currentIdentKey)
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

func deepCopy(dest Types, src Types) {
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
