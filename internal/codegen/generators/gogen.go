// Copyright 2026 Pouya Vedadiyan. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generators

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/vedadiyan/mango/internal/codegen"
	"github.com/vedadiyan/mango/static"
)

type (
	// GoBasicType, GoStructType and GoArrayType name Go types in generated code.
	GoBasicType  string
	GoStructType string
	GoArrayType  string

	// Types is the model of the code to generate. It maps each type name to
	// its fields, and each field name to its Go type and struct tag.
	//
	// Keys starting with "$" describe enums: the map holds the enum's
	// constants and its underlying type under "$type".
	Types map[string]map[string][2]string

	// CombinatorType names a schema combinator.
	CombinatorType string

	// GoTypeOptions configures GetGoType.
	GoTypeOptions struct {
		optional string // "*" if the type is optional
	}

	// GoTypeOption configures GetGoType.
	GoTypeOption func(*GoTypeOptions)

	// GoTagOption adds options to a struct tag built by GetTags.
	GoTagOption func(*[]string)

	// Field describes a field of a generated struct.
	Field struct {
		Name string
		Type string
		Tags string
	}

	// GoGenModel is the data passed to the code generation template.
	GoGenModel struct {
		PackageName string
		ImportList  []string
		Fields      Types
	}

	// ConcreteType is an element type together with its array depth.
	ConcreteType struct {
		Type  string
		Depth int
	}
)

const (
	// EmptyString is the empty string.
	EmptyString = ""

	// Supported combinator types.
	CombinatorTypeInavlid CombinatorType = "Invalid"
	CombinatorTypeOneOf   CombinatorType = "OneOf"
	CombinatorTypeAnyOf   CombinatorType = "AnyOf"
	CombinatorTypeAllOf   CombinatorType = "AllOf"
	CombinatorTypeNot     CombinatorType = "Not"
)

// Go type names used in generated code.
const (
	TimeType       = "time.Time"
	ObjectIdType   = "bson.ObjectID"
	BsonAType      = "bson.A"
	BsonEType      = "bson.E"
	BsonDType      = "bson.D"
	BsonMType      = "bson.M"
	BoolType       = "bool"
	IntType        = "int"
	Int8Type       = "int8"
	Int16Type      = "int16"
	Int32Type      = "int32"
	Int64Type      = "int64"
	UintType       = "uint"
	Uint8Type      = "uint8"
	Uint16Type     = "uint16"
	Uint32Type     = "uint32"
	Uint64Type     = "uint64"
	Float32Type    = "float32"
	Float64Type    = "float64"
	Complex64Type  = "complex64"
	Complex128Type = "complex128"
	StringType     = "string"
	ByteType       = "byte"
	RuneType       = "rune"
	BinaryType     = "[]byte"
	AnyType        = "any"
)

var (
	//go:embed gogen.go.tmpl
	goGenTemplate string

	// opaqueTypes are encoded as-is and tagged "opaque" so the codec does
	// not rewrite them.
	opaqueTypes = []string{
		TimeType,
		ObjectIdType,
		BsonAType,
		BsonEType,
		BsonDType,
		BsonMType,
	}

	// systemTypes are built-in and well-known types, left unchanged by the
	// genericType template function.
	systemTypes = []string{
		BoolType,
		IntType,
		Int8Type,
		Int16Type,
		Int32Type,
		Int64Type,
		UintType,
		Uint8Type,
		Uint16Type,
		Uint32Type,
		Uint64Type,
		Float32Type,
		Float64Type,
		Complex64Type,
		Complex128Type,
		StringType,
		ByteType,
		RuneType,
		TimeType,
		ObjectIdType,
	}
)

// TemplateFunctions returns the functions available to the code generation
// template:
//
//   - hasPrefix, trimPrefix: wrappers for strings.HasPrefix and TrimPrefix.
//   - genericType: replaces a non-system element type with T, keeping any
//     pointer and slice prefix (for example, []*Foo becomes []*T).
//   - arraySymbol: returns the pointer and slice prefix of a type.
//   - dict: builds a map from alternating keys and values.
func TemplateFunctions() template.FuncMap {
	splitPattern := regexp.MustCompile(`([\*\[\]]*)`)
	arrayPattern := regexp.MustCompile(`^([\*\[\]]{2,})+`)
	return template.FuncMap{
		"hasPrefix": func(a, b string) bool {
			return strings.HasPrefix(a, b)
		},
		"genericType": func(in string) string {
			values := SplitInTwo(splitPattern, in)
			if !slices.Contains(systemTypes, values[1]) {
				return fmt.Sprintf("%sT", values[0])
			}
			return in
		},
		"trimPrefix": func(in string, prefix string) string {
			return strings.TrimPrefix(in, prefix)
		},
		"arraySymbol": func(in string) string {
			return arrayPattern.FindString(in)
		},
		"dict": func(values ...any) map[string]any {
			result := make(map[string]any)
			for i := 0; i < len(values); i += 2 {
				result[values[i].(string)] = values[i+1]
			}
			return result
		},
	}
}

// GoGenRender renders types as Go source in package packageName, writes it
// to fileName.autogen.go and formats it with gofmt run in dir. It panics on
// failure.
func GoGenRender(dir string, fileName string, packageName string, types Types) {
	tmpl := template.New(uuid.New().String())
	tmpl.Funcs(TemplateFunctions())
	tmpl, err := tmpl.Parse(goGenTemplate)
	if err != nil {
		panic(err)
	}
	var buffer bytes.Buffer
	imports := []string{
		"fmt",
		"github.com/vedadiyan/mango/tools",
		"go.mongodb.org/mongo-driver/v2/bson",
		"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore",
	}
	if err := tmpl.Execute(&buffer, &GoGenModel{packageName, imports, types}); err != nil {
		panic(err)
	}
	dest := fmt.Sprintf("%s.autogen.go", fileName)
	os.WriteFile(dest, buffer.Bytes(), os.ModePerm)
	if err := Exec("gofmt", dir, "-w", dest); err != nil {
		panic(err)
	}
}

// ToGoTypeModel builds the type model for the object schema in. parents is
// the path of enclosing type names, which prefixes generated type names. If
// forceOptional is true, all fields are optional.
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

	rootCombinators, err := in.GetRootCombinators()
	if err != nil {
		return nil, err
	}

	requiredList, err := schema.GetRequiredStringArray()
	if err != nil {
		return nil, err
	}

	out := make(Types)
	parentIdentKey := flaten(parents)
	out[parentIdentKey] = make(map[string][2]string)

	for explicitKey, value := range properties {

		explicitRequired, err := value.GetRequiredOrFalse()
		if err != nil {
			return nil, err
		}
		if slices.Contains(requiredList, explicitKey) {
			explicitRequired = true
		}

		if value.IsArray() {
			res, err := ToGoArrayType([]bool{false}, explicitKey, value, parents, WithOmitEmpty(!explicitRequired))
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			continue
		}
		if value.IsCombinator() {
			res, err := ToGoCombinatorType(explicitKey, value, parents)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			continue
		}

		identKey := MakePascalCase(explicitKey)
		currentIdentSlice := append(parents, identKey)
		typeName := identKey

		if value.IsObject() {
			res, err := ToGoTypeModel(value, currentIdentSlice, forceOptional)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			typeName = flaten(currentIdentSlice)
		}

		explicitType, err := value.GetType()
		if err != nil {
			return nil, err
		}

		constructedType := explicitType
		isEnum := false

		if explicitType == nil {
			explicitEnum, err := value.GetEnum()
			if err != nil {
				return nil, err
			}

			if explicitEnum == nil {
				return nil, fmt.Errorf("unspecified type")
			}
			isEnum = true
			constructedTypeName := flaten(currentIdentSlice)
			constructedType = []string{constructedTypeName}
			enumType := flaten([]string{"$", constructedTypeName})
			out[enumType] = make(map[string][2]string)

			typ := EmptyString
			for _, value := range explicitEnum {
				if typ == EmptyString {
					typ = fmt.Sprintf("%T", value)
				}
				switch t := value.(type) {
				case string:
					{
						enumValue := fmt.Sprintf("\"%s\"", t)
						out[enumType][flaten(append(currentIdentSlice, t))] = [2]string{enumValue}
					}
				default:
					{
						enumValue := fmt.Sprintf("%v", t)
						out[enumType][flaten(append(currentIdentSlice, enumValue))] = [2]string{enumValue}
					}
				}
			}
			out[enumType]["$type"] = [2]string{typ}
		}

		goType, err := GetGoType(constructedType, typeName, WithRequired(explicitRequired && !forceOptional))
		if err != nil {
			return nil, err
		}

		goTags := GetTags(explicitKey, WithOmitEmpty(!explicitRequired || forceOptional), WithOpaque(goType))

		if isEnum {
			goType = fmt.Sprintf("$%s", goType)
		}

		out[parentIdentKey][identKey] = [2]string{goType, goTags}
	}

	for _, rootCombinator := range rootCombinators {
		res, err := ToGoCombinatorType(EmptyString, rootCombinator, parents)
		if err != nil {
			return nil, err
		}
		deepCopy(out, res)
	}

	return out, nil
}

// ToGoArrayType builds the type model for the array field explicitKey
// described by value. dim records, for each enclosing array dimension,
// whether its elements may be null. opts configure the field's struct tag.
func ToGoArrayType(dim []bool, explicitKey string, value codegen.TypedParam, parents []string, opts ...GoTagOption) (Types, error) {
	if strings.TrimSpace(explicitKey) == EmptyString {
		return nil, fmt.Errorf("array definition requires a key")
	}

	innerSchema, err := value.GetSchema()
	if err != nil {
		return nil, err
	}

	if innerSchema == nil {
		return nil, nil
	}

	allowNull, err := innerSchema.GetAllowNull()
	if err != nil {
		return nil, err
	}

	items, err := innerSchema.GetItemsAsArray()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("expected at least one type but found zero")
	}

	explicitRequired, err := value.GetRequiredOrFalse()
	if err != nil {
		return nil, err
	}

	identKey := MakePascalCase(explicitKey)
	parentIdentKey := flaten(parents)
	currentIdentSlice := append(parents, identKey)

	out := make(Types)
	out[parentIdentKey] = make(map[string][2]string)

	if len(items) > 1 {
		out[parentIdentKey][identKey] = [2]string{"[]any"}
		return out, nil
	}

	currentType := items[0]

	if currentType.IsObject() {
		res, err := ToGoTypeModel(items[0], currentIdentSlice, false)
		if err != nil {
			return nil, err
		}
		deepCopy(out, res)
	}

	if currentType.IsArray() {
		res, err := ToGoArrayType(append(dim, allowNull), identKey, items[0], parents, WithOmitEmpty(!explicitRequired))
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	explicitType, err := currentType.GetType()
	if err != nil {
		return nil, err
	}

	goType, err := GetGoType(explicitType, flaten(currentIdentSlice))
	if err != nil {
		return nil, err
	}

	if allowNull {
		goType = fmt.Sprintf("*%s", goType)
	}

	goTags := GetTags(explicitKey, opts...)

	arrayDimensions := ""
	for _, value := range dim {
		pointer := ""
		if value {
			pointer = "*"
		}
		arrayDimensions = fmt.Sprintf("%s%s[]", arrayDimensions, pointer)
	}

	out[parentIdentKey][identKey] = [2]string{fmt.Sprintf("%s%s", arrayDimensions, goType), goTags}
	return out, nil
}

// ToGoCombinatorType builds the type model for a combinator. If explicitKey
// is empty, the combinator's fields are merged into the parent type;
// otherwise they form a new type referenced by field explicitKey.
func ToGoCombinatorType(explicitKey string, value codegen.TypedParam, parents []string) (Types, error) {
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

	identKey := MakePascalCase(explicitKey)
	isDynamicIdent := false
	if strings.TrimSpace(identKey) == EmptyString {
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
	out[parentIdentKey] = make(map[string][2]string)
	out[currentIdentKey] = make(map[string][2]string)

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
			res, err := ToGoArrayType([]bool{false}, innerIdentKey, spec, currentIdentSlice)
			if err != nil {
				return nil, err
			}
			deepCopy(out, res)
			continue
		}

		if spec.IsObject() {
			res, err := ToGoTypeModel(spec, innerIdentSlice, !hasInnerIdentKey)
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

		goTags := GetTags(*explicitObjectName, WithOmitEmpty(!explicitRequired), WithOpaque(goType))

		out[currentIdentKey][innerIdentKey] = [2]string{goType, goTags}
	}

	if !isDynamicIdent {
		goTags := GetTags(explicitKey)
		out[parentIdentKey][identKey] = [2]string{currentIdentKey, goTags}
		return out, nil
	}

	out[parentIdentKey] = out[currentIdentKey]
	delete(out, currentIdentKey)
	return out, nil
}

// WithRequired makes the type a pointer unless required is true.
func WithRequired(required bool) GoTypeOption {
	return func(gto *GoTypeOptions) {
		if !required {
			gto.optional = "*"
		}
	}
}

// GetGoType maps schema type names to a Go type. name is used for object
// types. A "null" type makes the result a pointer. More than two types
// yield any.
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
		switch static.BasicType(i) {
		case static.TypeNil:
			{
				options.optional = "*"
			}
		case static.TypeObject:
			{
				typ = name
			}
		case static.TypeBinary:
			{
				typ = BinaryType
			}
		case static.TypeBool:
			{
				typ = BoolType
			}
		case static.TypeDecimal, static.TypeDouble:
			{
				typ = Float64Type
			}
		case static.TypeInt:
			{
				typ = Int32Type
			}
		case static.TypeLong:
			{
				typ = Int64Type
			}
		case static.TypeString, static.TypeJavaScript, static.TypeRegex:
			{
				typ = StringType
			}
		case static.TypeTimeStamp:
			{
				typ = TimeType
			}
		case static.TypeObjectId:
			{
				typ = ObjectIdType
			}
		default:
			{
				typ = i
			}
		}
	}

	if strings.TrimSpace(typ) == EmptyString {
		return AnyType, nil
	}

	return fmt.Sprintf("%s%s", options.optional, typ), nil
}

// WithOmitEmpty adds the omitempty option if val is true.
func WithOmitEmpty(val bool) GoTagOption {
	return func(s *[]string) {
		if val {
			*s = append(*s, "omitempty")
		}
	}
}

// WithOpaque adds the opaque option if goTypeName is an opaque type.
func WithOpaque(goTypeName string) GoTagOption {
	return func(s *[]string) {
		if slices.Contains(opaqueTypes, strings.TrimLeft(goTypeName, "*")) {
			*s = append(*s, "opaque")
		}
	}
}

// GetTags returns a bson struct tag for the field explicitName.
func GetTags(explicitName string, opts ...GoTagOption) string {
	tags := make([]string, 0)
	tags = append(tags, explicitName)

	for _, opt := range opts {
		opt(&tags)
	}

	return fmt.Sprintf("`bson:\"%s\"`", strings.Join(tags, ","))
}

// deepCopy merges src into dest. Fields of a type present in both are
// merged, with src taking precedence.
func deepCopy(dest Types, src Types) {
	for key, value := range src {
		if _, ok := dest[key]; !ok {
			dest[key] = value
			continue
		}
		maps.Copy(dest[key], value)
	}
}

// flaten concatenates in into a single identifier.
func flaten(in []string) string {
	out := bytes.NewBufferString("")

	for _, value := range in {
		out.WriteString(value)
	}

	return out.String()
}

// MakePascalCase converts str to PascalCase.
func MakePascalCase(str string) string {
	return strcase.ToCamel(str)
}

// GetCombinatorType returns the combinator named by the fully qualified type
// name in, and reports whether in names a combinator.
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

// SplitInTwo splits s into its leading pointer and slice prefix (matched by
// re) and the remaining element type.
func SplitInTwo(re *regexp.Regexp, s string) []string {
	n := 2

	if len(s) == 0 {
		return []string{""}
	}

	matches := re.FindAllStringIndex(s, n)
	strings := make([]string, 0, len(matches))

	for i := range len(matches) - 1 {
		match := matches[i]
		if n > 0 && len(strings) >= n-1 {
			break
		}

		strings = append(strings, s[match[0]:match[1]])
	}

	finalMatch := matches[len(matches)-1]
	strings = append(strings, s[finalMatch[0]-1:])

	return strings
}

// Exec runs the command name with args in dir, with a one-minute timeout.
// On failure, the error includes the command's standard error.
func Exec(name string, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command execution timeout: %s", name)
		}
		return fmt.Errorf("command %q failed: %w\nstderr: %s", name, err, stderr.String())
	}

	return nil
}
