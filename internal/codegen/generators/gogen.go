package generators

import (
	"bytes"
	_ "embed"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/vedadiyan/mango/internal/codegen"
	"github.com/vedadiyan/mango/static"
)

type (
	GoBasicType  string
	GoStructType string
	GoArrayType  string

	Types map[string]map[string][2]string

	CombinatorType string

	GoTypeOptions struct {
		optional string
	}

	GoTypeOption func(*GoTypeOptions)
	GoTagOption  func(*[]string)

	Field struct {
		Name string
		Type string
		Tags string
	}

	GoGenModel struct {
		PackageName string
		ImportList  []string
		Fields      Types
	}
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

var (
	//go:embed gogen.go.tmpl
	goGenTemplate string
	opaqueTypes   = []string{"time.Time", "bson.ObjectID"}
	systemTypes   = []string{
		"bool",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"float32",
		"float64",
		"complex64",
		"complex128",
		"string",
		"byte",
		"rune",
		"time.Time",
		"bson.ObjectID",
	}
)

func GoGenRender(types Types) {
	tmpl := template.New("test")
	pattern := regexp.MustCompile(`([\*\[\]]*)`)
	tmpl.Funcs(template.FuncMap{
		"hasPrefix": func(a, b string) bool {
			return strings.HasPrefix(a, b)
		},
		"whichType": func(in string) string {
			values := SplitInTwo(pattern, in)
			if !slices.Contains(systemTypes, values[1]) {
				return fmt.Sprintf("%sT", values[0])
			}
			return in
		},
		"trimPrefix": func(in string, prefix string) string {
			return strings.TrimPrefix(in, prefix)
		},
	})
	tmpl, err := tmpl.Parse(goGenTemplate)
	if err != nil {
		panic(err)
	}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, &GoGenModel{"test", []string{"github.com/vedadiyan/mango/tools", "go.mongodb.org/mongo-driver/v2/bson"}, types}); err != nil {
		panic(err)
	}
	os.WriteFile("test.go", buffer.Bytes(), os.ModePerm)
}

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

		if value.IsArray() {
			res, err := ToGoArrayType(1, explicitKey, value, parents)
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

		explicitRequired, err := value.GetRequiredOrFalse()
		if err != nil {
			return nil, err
		}
		if slices.Contains(requiredList, explicitKey) {
			explicitRequired = true
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

func ToGoArrayType(dim int, explicitKey string, value codegen.TypedParam, parents []string) (Types, error) {
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
		res, err := ToGoArrayType(dim+1, identKey, items[0], parents)
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

	goTags := GetTags(explicitKey, WithOmitEmpty(!explicitRequired))

	out[parentIdentKey][identKey] = [2]string{fmt.Sprintf("%s%s", strings.Repeat("[]", dim), goType), goTags}
	return out, nil
}

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
			res, err := ToGoArrayType(ZeroDimension, innerIdentKey, spec, currentIdentSlice)
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
		switch static.BasicType(i) {
		case static.TypeNil:
			{
				options.optional = "*"
			}
		case "object":
			{
				typ = name
			}
		case static.TypeBinary:
			{
				typ = "[]byte"
			}
		case static.TypeBool:
			{
				typ = "bool"
			}
		case static.TypeDecimal, static.TypeDouble:
			{
				typ = "float64"
			}
		case static.TypeInt:
			{
				typ = "int32"
			}
		case static.TypeLong:
			{
				typ = "int64"
			}
		case static.TypeString, static.TypeJavaScript, static.TypeRegex:
			{
				typ = "string"
			}
		case static.TypeTimeStamp:
			{
				typ = "time.Time"
			}
		case static.TypeObjectId:
			{
				typ = "bson.ObjectID"
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

func WithOmitEmpty(val bool) GoTagOption {
	return func(s *[]string) {
		if val {
			*s = append(*s, "omitempty")
		}
	}
}

func WithOpaque(goTypeName string) GoTagOption {
	return func(s *[]string) {
		if slices.Contains(opaqueTypes, strings.TrimLeft(goTypeName, "*")) {
			*s = append(*s, "opaque")
		}
	}
}

func GetTags(explicitName string, opts ...GoTagOption) string {
	tags := make([]string, 0)
	tags = append(tags, explicitName)

	for _, opt := range opts {
		opt(&tags)
	}

	return fmt.Sprintf("`bson:\"%s\"`", strings.Join(tags, ","))
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

func flaten(in []string) string {
	out := bytes.NewBufferString("")

	for _, value := range in {
		out.WriteString(value)
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
