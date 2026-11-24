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

package codegen

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
)

type (
	// ParserContext holds a parsed and type-checked Go source file.
	ParserContext struct {
		Ast         *ast.File
		Info        *types.Info
		PackageName string
	}

	// KeyValue is a parsed key-value element of a composite literal.
	KeyValue struct {
		Key   string
		Value any
	}

	// TypedParam is a parsed composite literal and its fully qualified type
	// name. Value is a RawSchema for struct and map literals, or []any for
	// slice and array literals.
	TypedParam struct {
		TypeName string
		Value    any
	}

	// RawSchema maps the keys of a parsed struct or map literal to their
	// parsed values.
	RawSchema map[string]any

	// Properties maps property names to their schemas.
	Properties map[string]TypedParam

	// Items is a list of item schemas.
	Items []TypedParam
)

// Fully qualified names of the schema types recognized by the parser.
const (
	ScalarType    = "github.com/vedadiyan/mango/static.Scalar"
	CompositeType = "github.com/vedadiyan/mango/static.Composite"

	OneOf = "github.com/vedadiyan/mango/static.Combinator[github.com/vedadiyan/mango/static.OneOf]"
	AnyOf = "github.com/vedadiyan/mango/static.Combinator[github.com/vedadiyan/mango/static.AnyOf]"
	AllOf = "github.com/vedadiyan/mango/static.Combinator[github.com/vedadiyan/mango/static.AllOf]"
	Not   = "github.com/vedadiyan/mango/static.Combinator[github.com/vedadiyan/mango/static.Not]"
)

var (
	// arrayPattern matches a static.Array instantiation and captures its
	// element type.
	arrayPattern = regexp.MustCompile(`^(github\.com\/vedadiyan\/mango\/static\.Array\[)(.*)(\])`)
)

// Parse parses and type-checks the Go source file at filePath. Imports are
// resolved from source.
func Parse(filePath string) (*ParserContext, error) {
	fset := token.NewFileSet()
	parsedAst, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	info := types.Info{
		Defs:       make(map[*ast.Ident]types.Object),
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Instances:  make(map[*ast.Ident]types.Instance),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	config := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	if _, err := config.Check(parsedAst.Name.Name, fset, []*ast.File{parsedAst}, &info); err != nil {
		return nil, err
	}

	return &ParserContext{parsedAst, &info, parsedAst.Name.String()}, nil
}

// ExtractFuncs returns the top-level function and method declarations in
// the file. The returned error is always nil.
func (pc *ParserContext) ExtractFuncs() ([]*ast.FuncDecl, error) {
	out := make([]*ast.FuncDecl, 0)
	for _, decl := range pc.Ast.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			out = append(out, fn)
		}
	}
	return out, nil
}

// ParseBasicLit returns the value of a basic literal: a string for string
// and rune literals, an int64 for integer literals and a float64 for
// floating-point literals.
func (pc *ParserContext) ParseBasicLit(bl *ast.BasicLit) (any, error) {
	switch bl.Kind {
	case token.STRING, token.CHAR:
		{
			return strconv.Unquote(bl.Value)
		}
	case token.INT:
		{
			val := constant.MakeFromLiteral(bl.Value, bl.Kind, 0)
			num, ok := constant.Int64Val(val)
			if !ok {
				return "", fmt.Errorf("`%s` is not int", bl.Value)
			}
			return num, nil
		}
	case token.FLOAT, token.IMAG:
		{
			val := constant.MakeFromLiteral(bl.Value, bl.Kind, 0)
			num, ok := constant.Float64Val(val)
			if !ok {
				return "", fmt.Errorf("`%s` is not float", bl.Value)
			}
			return num, nil
		}
	default:
		{
			return "", fmt.Errorf("`%s` is not statically typed", bl.Value)
		}
	}
}

// ParseVar returns the name of the struct field vr. Other variables are
// rejected because their values are not known statically.
func (pc *ParserContext) ParseVar(vr *types.Var, ident *ast.Ident) (string, error) {
	if !vr.IsField() {
		return "", fmt.Errorf("`%s` is not statically typed", ident.String())
	}

	return vr.Name(), nil
}

// ParseIdent resolves ident to either a constant value or a struct field
// name. Constants yield an int64, float64, bool, complex128 or string.
// origin is currently unused.
func (pc *ParserContext) ParseIdent(ident *ast.Ident, origin ast.Expr) (any, error) {
	data, ok := pc.Info.Uses[ident]
	if !ok {
		return "", fmt.Errorf("`%s` is not present", ident.String())
	}
	switch t := data.(type) {
	case *types.Const:
		{
			if t.Val().Kind() == constant.Int {
				num, ok := constant.Int64Val(t.Val())
				if !ok {
					return "", fmt.Errorf("`%s` is not int", t.Val())
				}
				return num, nil
			}
			if t.Val().Kind() == constant.Float {
				num, ok := constant.Float64Val(t.Val())
				if !ok {
					return "", fmt.Errorf("`%s` is not int", t.Val())
				}
				return num, nil
			}
			if t.Val().Kind() == constant.Bool {
				return constant.BoolVal(t.Val()), nil
			}
			if t.Val().Kind() == constant.Complex {
				return strconv.ParseComplex(t.Val().ExactString(), 64)
			}
			return constant.StringVal(t.Val()), nil
		}
	case *types.Var:
		{
			return pc.ParseVar(t, ident)
		}
	default:
		{
			return "", fmt.Errorf("`%s` is not supported", ident.String())
		}
	}
}

// ParseKeyValue parses a key-value element of a composite literal. The key
// must resolve to a string, such as a struct field name.
func (pc *ParserContext) ParseKeyValue(kv *ast.KeyValueExpr, origin ast.Expr) (*KeyValue, error) {
	key, err := pc.ParseExpr(kv.Key, origin)
	if err != nil {
		return nil, err
	}
	keyStr, ok := key.(string)
	if !ok {
		return nil, fmt.Errorf("expected string but found %T", key)
	}

	value, err := pc.ParseExpr(kv.Value, origin)
	if err != nil {
		return nil, err
	}

	return &KeyValue{keyStr, value}, nil
}

// ParseCompositLit parses the elements of a slice or array literal.
func (pc *ParserContext) ParseCompositLit(current *ast.CompositeLit) ([]any, error) {
	out := make([]any, 0)
	for _, i := range current.Elts {
		val, err := pc.ParseExpr(i, current)
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return out, nil
}

// ParseCompositLitAsMap parses a struct or map literal whose elements are
// all key-value pairs.
func (pc *ParserContext) ParseCompositLitAsMap(current *ast.CompositeLit) (RawSchema, error) {
	out := make(RawSchema)
	for _, i := range current.Elts {
		val, err := pc.ParseExpr(i, current)
		if err != nil {
			return nil, err
		}
		kv, ok := val.(*KeyValue)
		if !ok {
			return nil, fmt.Errorf("expected KeyValue but found %T", val)
		}
		out[kv.Key] = kv.Value
	}
	return out, nil
}

// ParseExpr parses a statically known expression into a Go value. Composite
// literals become a TypedParam, and a leading & is ignored. origin is the
// enclosing composite literal, or nil at the top level.
func (pc *ParserContext) ParseExpr(current ast.Expr, origin ast.Expr) (any, error) {
	switch t := current.(type) {
	case *ast.BasicLit:
		{
			return pc.ParseBasicLit(t)
		}
	case *ast.Ident:
		{
			return pc.ParseIdent(t, origin)
		}
	case *ast.SelectorExpr:
		{
			return pc.ParseIdent(t.Sel, origin)
		}
	case *ast.KeyValueExpr:
		{
			return pc.ParseKeyValue(t, origin)
		}
	case *ast.CompositeLit:
		{
			typ, ok := pc.Info.Types[current]
			if !ok {
				return nil, fmt.Errorf("`%v` is not present", t)
			}
			switch typ.Type.Underlying().(type) {
			case *types.Slice, *types.Array:
				{
					out, err := pc.ParseCompositLit(t)
					if err != nil {
						return nil, err
					}
					return TypedParam{typ.Type.String(), out}, nil
				}
			default:
				{
					out, err := pc.ParseCompositLitAsMap(t)
					if err != nil {
						return nil, err
					}
					return TypedParam{typ.Type.String(), out}, nil
				}
			}
		}
	case *ast.UnaryExpr:
		{
			op := t.Op.String()
			if op != "&" && op != "*" {
				return nil, fmt.Errorf("`%v` unsupported statement", t)
			}
			return pc.ParseExpr(t.X, origin)
		}
	default:
		{
			return nil, fmt.Errorf("`%v` unsupported statement", t)
		}
	}
}

// GetTitle returns the schema's Title, or nil if it is not set.
func (r RawSchema) GetTitle() (*string, error) {
	return r.LookupCast[string]("Title")
}

// GetDescription returns the schema's Description, or nil if it is not set.
func (r RawSchema) GetDescription() (*string, error) {
	return r.LookupCast[string]("Description")
}

// GetSchema returns r's value as a RawSchema.
func (r TypedParam) GetSchema() (*RawSchema, error) {
	return Cast[RawSchema](r.Value)
}

// GetRequired returns the schema's Required flag, or nil if it is not set.
// For object schemas, where Required lists property names, use
// RawSchema.GetRequiredStringArray.
func (r TypedParam) GetRequired() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("Required")
}

// GetRequiredOrFalse is like GetRequired but returns false when Required is
// not set or is not a bool.
func (r TypedParam) GetRequiredOrFalse() (bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return false, err
	}
	value := innerSchema.Lookup("Required")
	switch t := value.(type) {
	case bool:
		{
			return t, nil
		}
	default:
		{
			return false, nil
		}
	}
}

// GetRequiredStringArray returns the property names listed in Required, or
// nil if Required is not set or is not a list.
func (r RawSchema) GetRequiredStringArray() ([]string, error) {
	required := r.Lookup("Required")
	if required == nil {
		return nil, nil
	}
	switch t := required.(type) {
	case TypedParam:
		{
			values, err := Cast[[]any](t.Value)
			if err != nil {
				return nil, err
			}
			out := make([]string, 0)
			for _, i := range *values {
				str, err := Cast[string](i)
				if err != nil {
					return nil, err
				}
				out = append(out, *str)
			}
			return out, nil
		}
	default:
		{
			return nil, nil
		}
	}
}

// GetMinLen returns the schema's MinLen, or nil if it is not set.
func (r TypedParam) GetMinLen() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("MinLen")
}

// GetMaxLen returns the schema's MaxLen, or nil if it is not set.
func (r TypedParam) GetMaxLen() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("MaxLen")
}

// GetMin returns the schema's Min, or nil if it is not set.
func (r TypedParam) GetMin() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("Min")
}

// GetMax returns the schema's Max, or nil if it is not set.
func (r TypedParam) GetMax() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("Max")
}

// GetExclusiveMin returns the schema's ExclusiveMin flag, or nil if it is not set.
func (r TypedParam) GetExclusiveMin() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("ExclusiveMin")
}

// GetExclusiveMax returns the schema's ExclusiveMax flag, or nil if it is not set.
func (r TypedParam) GetExclusiveMax() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("ExclusiveMax")
}

// GetMultipleOf returns the schema's MultipleOf, or nil if it is not set.
func (r TypedParam) GetMultipleOf() (*float64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[float64]("MultipleOf")
}

// GetPattern returns the schema's Pattern, or nil if it is not set.
func (r TypedParam) GetPattern() (*string, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[string]("Pattern")
}

// GetMinItems returns the schema's MinItems, or nil if it is not set.
func (r TypedParam) GetMinItems() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("MinItems")
}

// GetMaxItems returns the schema's MaxItems, or nil if it is not set.
func (r TypedParam) GetMaxItems() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("MaxItems")
}

// GetUniqueItems returns the schema's UniqueItems flag, or nil if it is not set.
func (r TypedParam) GetUniqueItems() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("UniqueItems")
}

// GetMinProperties returns the schema's MinProperties, or nil if it is not set.
func (r TypedParam) GetMinProperties() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("MinProperties")
}

// GetMaxProperties returns the schema's MaxProperties, or nil if it is not set.
func (r TypedParam) GetMaxProperties() (*int64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int64]("MaxProperties")
}

// GetAdditionalItems returns the schema's AdditionalItems flag, or nil if it is not set.
func (r TypedParam) GetAdditionalItems() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("AdditionalItems")
}

// GetAdditionalProperties returns the schema's AdditionalProperties flag, or nil if it is not set.
func (r TypedParam) GetAdditionalProperties() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("AdditionalProperties")
}

// GetEnum returns the values listed in the schema's Enum, or nil if it is not set.
func (r TypedParam) GetEnum() ([]any, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	enum, err := innerSchema.LookupCast[TypedParam]("Enum")
	if err != nil {
		return nil, err
	}
	if enum == nil {
		return nil, nil
	}
	values, err := Cast[[]any](enum.Value)
	if err != nil {
		return nil, err
	}
	if values == nil {
		return nil, nil
	}
	return *values, nil
}

// IsCombinator reports whether r is a OneOf, AnyOf, AllOf or Not combinator.
func (r TypedParam) IsCombinator() bool {
	switch r.TypeName {
	case OneOf, AnyOf, AllOf, Not:
		{
			return true
		}
	default:
		{
			return false
		}
	}
}

// IsArray reports whether r is a static.Array of scalars, composites or nested arrays.
func (r TypedParam) IsArray() bool {
	return isArray(r.TypeName)
}

// isArray reports whether typeName is a static.Array whose element type is
// Scalar, Composite or, recursively, another such array.
func isArray(typeName string) bool {
	subMatches := arrayPattern.FindStringSubmatch(typeName)
	if len(subMatches) != 4 {
		return false
	}

	nextToken := subMatches[2]

	switch nextToken {
	case ScalarType, CompositeType:
		{
			return true
		}
	default:
		{
			return isArray(nextToken)
		}
	}
}

// IsObject reports whether r is a composite (object) schema.
func (r TypedParam) IsObject() bool {
	switch r.TypeName {
	case CompositeType:
		{
			return true
		}
	default:
		{
			return false
		}
	}
}

// GetType returns the JSON Schema type names for r: "array" for arrays,
// "object" for composites and the combinator name for combinators. For
// scalars, it returns the values listed in the Type field, or nil if Type is
// not set.
func (r TypedParam) GetType() ([]string, error) {
	if r.IsArray() {
		return []string{"array"}, nil
	}

	out := make([]string, 0)
	switch r.TypeName {
	case ScalarType:
		{
			break
		}
	case CompositeType:
		{
			return []string{"object"}, nil
		}
	case OneOf:
		{
			return []string{"oneof"}, nil
		}
	case AnyOf:
		{
			return []string{"anyof"}, nil
		}
	case AllOf:
		{
			return []string{"allof"}, nil
		}
	case Not:
		{
			return []string{"not"}, nil
		}
	}

	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}

	innerType, err := innerSchema.LookupCast[TypedParam]("Type")
	if err != nil {
		return nil, err
	}
	if innerType == nil {
		return nil, nil
	}

	types, err := CopyCast[[]any](innerType.Value)
	if err != nil {
		return nil, err
	}
	for _, typ := range types {
		str, err := Cast[string](typ)
		if err != nil {
			return nil, err
		}
		out = append(out, *str)
	}

	return out, nil
}

// GetOneOfs returns the schema's OneOf combinator, or nil if it is not set.
func (r TypedParam) GetOneOfs() (*TypedParam, error) {
	rawSchema, err := r.GetSchema()
	if err != nil {
		return nil, err
	}
	if rawSchema == nil {
		return nil, nil
	}
	oneOfs, err := rawSchema.LookupCast[TypedParam]("OneOf")
	if err != nil {
		return nil, err
	}
	return oneOfs, nil
}

// GetAnyOfs returns the schema's AnyOf combinator, or nil if it is not set.
func (r TypedParam) GetAnyOfs() (*TypedParam, error) {
	rawSchema, err := r.GetSchema()
	if err != nil {
		return nil, err
	}
	if rawSchema == nil {
		return nil, nil
	}
	oneOfs, err := rawSchema.LookupCast[TypedParam]("AnyOf")
	if err != nil {
		return nil, err
	}
	return oneOfs, nil
}

// GetAllOfs returns the schema's AllOf combinator, or nil if it is not set.
func (r TypedParam) GetAllOfs() (*TypedParam, error) {
	rawSchema, err := r.GetSchema()
	if err != nil {
		return nil, err
	}
	if rawSchema == nil {
		return nil, nil
	}
	oneOfs, err := rawSchema.LookupCast[TypedParam]("AllOf")
	if err != nil {
		return nil, err
	}
	return oneOfs, nil
}

// GetNots returns the schema's Not combinator, or nil if it is not set.
func (r TypedParam) GetNots() (*TypedParam, error) {
	rawSchema, err := r.GetSchema()
	if err != nil {
		return nil, err
	}
	if rawSchema == nil {
		return nil, nil
	}
	oneOfs, err := rawSchema.LookupCast[TypedParam]("Not")
	if err != nil {
		return nil, err
	}
	return oneOfs, nil
}

// GetRootCombinators returns the combinators set on the schema, in the order
// OneOf, AnyOf, AllOf, Not.
func (r TypedParam) GetRootCombinators() ([]TypedParam, error) {
	out := make([]TypedParam, 0)

	if v, err := r.GetOneOfs(); err == nil {
		if v != nil {
			out = append(out, *v)
		}
	} else {
		return nil, err
	}

	if v, err := r.GetAnyOfs(); err == nil {
		if v != nil {
			out = append(out, *v)
		}
	} else {
		return nil, err
	}

	if v, err := r.GetAllOfs(); err == nil {
		if v != nil {
			out = append(out, *v)
		}
	} else {
		return nil, err
	}

	if v, err := r.GetNots(); err == nil {
		if v != nil {
			out = append(out, *v)
		}
	} else {
		return nil, err
	}

	return out, nil
}

// GetProperties returns the schema's Properties, or nil if it is not set.
func (r RawSchema) GetProperties() (Properties, error) {
	typedParam, err := r.LookupCast[TypedParam]("Properties")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}
	rawSchema, err := Cast[RawSchema](typedParam.Value)
	if err != nil {
		return nil, err
	}
	properties := make(Properties)
	for key, value := range *rawSchema {
		typedParam, err := Cast[TypedParam](value)
		if err != nil {
			return nil, err
		}
		properties[key] = *typedParam
	}

	return properties, nil
}

// GetConditions returns the schema's Conditions as plain maps, with nested
// schemas converted recursively. It returns nil if Conditions is not set.
func (r RawSchema) GetConditions() ([]map[string]any, error) {
	typedParam, err := r.LookupCast[TypedParam]("Conditions")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}
	items, err := CopyCast[[]any](typedParam.Value)
	if err != nil {
		return nil, err
	}
	return parseCondition(items)
}

// GetObjectName returns the schema's ObjectName, or nil if it is not set.
func (r RawSchema) GetObjectName() (*string, error) {
	return r.LookupCast[string]("ObjectName")
}

// parseCondition converts a list of condition literals to plain maps.
func parseCondition(rawSchema []any) ([]map[string]any, error) {
	items := make([]map[string]any, 0)
	for _, value := range rawSchema {
		typedParam, err := Cast[TypedParam](value)
		if err != nil {
			return nil, err
		}
		if typedParam == nil {
			continue
		}
		res, err := typedParam.getConditions()
		if err != nil {
			return nil, err
		}
		items = append(items, res...)
	}

	return items, nil
}

// getConditions converts r to a plain map, converting nested schemas recursively.
func (r RawSchema) getConditions() (map[string]any, error) {
	items := make(map[string]any, 0)
	for key, value := range r {
		switch t := value.(type) {
		case TypedParam:
			{
				res, err := t.getConditions()
				if err != nil {
					return nil, err
				}
				items[key] = res
			}
		case RawSchema:
			{
				res, err := t.getConditions()
				if err != nil {
					return nil, err
				}
				items[key] = res
			}
		default:
			{
				items[key] = t
			}
		}
	}

	return items, nil
}

// getConditions converts r to a list of plain maps. A schema yields one map;
// a list yields one or more maps per element.
func (r TypedParam) getConditions() ([]map[string]any, error) {
	switch t := r.Value.(type) {
	case RawSchema:
		{
			res, err := t.getConditions()
			if err != nil {
				return nil, err
			}
			return []map[string]any{res}, nil
		}
	case []any:
		{
			items := make([]map[string]any, 0)
			for _, value := range t {
				typedParam, err := Cast[TypedParam](value)
				if err != nil {
					return nil, err
				}
				if typedParam == nil {
					continue
				}
				res, err := typedParam.getConditions()
				if err != nil {
					return nil, err
				}
				items = append(items, res...)
			}

			return items, nil
		}
	default:
		{
			return nil, fmt.Errorf("unexpected case")
		}
	}

}

// GetItems returns the schema's Items: a TypedParam for a single item
// schema, or Items for a list of item schemas. It returns nil if Items is
// not set.
func (r RawSchema) GetItems() (any, error) {
	typedParam, err := r.LookupCast[TypedParam]("Items")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}

	switch t := typedParam.Value.(type) {
	case []any:
		{
			return r.getItemsArray(t)
		}
	default:
		{
			return *typedParam, nil
		}
	}
}

// GetAllowNull returns the schema's AllowNull flag, or false if it is not set.
func (r RawSchema) GetAllowNull() (bool, error) {
	value, err := r.LookupCast[bool]("AllowNull")
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}

	return *value, nil
}

// GetItemsAsArray is like GetItems but always returns a list, wrapping a
// single item schema in a one-element list.
func (r RawSchema) GetItemsAsArray() (Items, error) {
	typedParam, err := r.LookupCast[TypedParam]("Items")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}

	switch t := typedParam.Value.(type) {
	case []any:
		{
			return r.getItemsArray(t)
		}
	default:
		{
			return Items{*typedParam}, nil
		}
	}
}

// getItemsArray converts a list of item literals to Items.
func (r RawSchema) getItemsArray(rawSchema []any) (Items, error) {
	items := make(Items, 0)
	for _, value := range rawSchema {
		typedParam, err := Cast[TypedParam](value)
		if err != nil {
			return nil, err
		}
		if typedParam == nil {
			continue
		}
		items = append(items, *typedParam)
	}

	return items, nil
}

// GetSpecs returns the schema's Specs, or nil if it is not set.
func (r RawSchema) GetSpecs() (Items, error) {
	typedParam, err := r.LookupCast[TypedParam]("Specs")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}
	rawSchema, err := CopyCast[[]any](typedParam.Value)
	if err != nil {
		return nil, err
	}
	items := make(Items, 0)
	for _, value := range rawSchema {
		typedParam, err := Cast[TypedParam](value)
		if err != nil {
			return nil, err
		}
		if typedParam == nil {
			continue
		}
		items = append(items, *typedParam)
	}

	return items, nil
}

// GetDependencies returns the schema's Dependencies, mapping each property
// to the properties it requires. It returns nil if Dependencies is not set.
func (r RawSchema) GetDependencies() (map[string][]string, error) {
	typedParam, err := r.LookupCast[TypedParam]("Dependencies")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}
	rawSchema, err := Cast[RawSchema](typedParam.Value)
	if err != nil {
		return nil, err
	}
	dependencies := make(map[string][]string)
	for key, value := range *rawSchema {
		typedParam, err := Cast[TypedParam](value)
		if err != nil {
			return nil, err
		}
		if typedParam == nil {
			continue
		}
		rawValues, err := CopyCast[[]any](typedParam.Value)
		if err != nil {
			return nil, err
		}
		if rawValues == nil {
			continue
		}
		l := len(rawValues)
		values := make([]string, l)
		for i := range l {
			value, err := CopyCast[string](rawValues[i])
			if err != nil {
				return nil, err
			}
			values[i] = value
		}
		dependencies[key] = values
	}
	return dependencies, nil
}

// Cast returns a pointer to a copy of v asserted to type T, or an error if v
// is not a T.
func Cast[T any](v any) (*T, error) {
	val, ok := v.(T)
	if !ok {
		var zero T
		return nil, fmt.Errorf("cannot convert %T to %T", v, zero)
	}
	return &val, nil
}

// CopyCast is like Cast but returns the value instead of a pointer.
func CopyCast[T any](v any) (T, error) {
	val, ok := v.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("cannot convert %T to %T", v, zero)
	}
	return val, nil
}

// LookupCast is like the LookupCast function, using r as the map.
func (r RawSchema) LookupCast[T any](key string) (*T, error) {
	return LookupCast[T](key, r)
}

// LookupCast returns the value stored under key in mapper, asserted to type
// T. It returns nil and no error if key is not present.
func LookupCast[T any, R any](key string, mapper map[string]R) (*T, error) {
	val, ok := mapper[key]
	if !ok {
		return nil, nil
	}
	return Cast[T](val)
}

// Lookup returns the value stored under key, or nil if it is not present.
func (r RawSchema) Lookup(key string) any {
	return r[key]
}
