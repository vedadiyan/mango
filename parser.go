package mango

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

type (
	ParserContext struct {
		Ast         *ast.File
		Info        *types.Info
		PackageName string
	}
	KeyValue struct {
		Key   string
		Value any
	}
	TypedParam struct {
		TypeName string
		Value    any
	}
	RawSchema  map[string]any
	Properties map[string]TypedParam
	Items      []TypedParam
)

const (
	ScalarType = "github.com/vedadiyan/mango/static.Scalar"
	ObjectType = "github.com/vedadiyan/mango/static.Composite"
	ArrayType  = "github.com/vedadiyan/mango/static.Array"
)

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

func (pc *ParserContext) ExtractFuncs() ([]*ast.FuncDecl, error) {
	out := make([]*ast.FuncDecl, 0)
	for _, decl := range pc.Ast.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			out = append(out, fn)
		}
	}
	return out, nil
}

func (pc *ParserContext) ParseBasicLint(bl *ast.BasicLit) (any, error) {
	switch bl.Kind {
	case token.STRING, token.IMAG, token.CHAR:
		{
			return strings.TrimRight(strings.TrimLeft(bl.Value, "\""), "\""), nil
		}
	case token.INT:
		{
			return strconv.Atoi(bl.Value)
		}
	case token.FLOAT:
		{
			return strconv.ParseFloat(bl.Value, 64)
		}
	default:
		{
			return "", fmt.Errorf("`%s` is not statically typed", bl.Value)
		}
	}
}

func (pc *ParserContext) ParseVar(vr *types.Var, ident *ast.Ident) (string, error) {
	if !vr.IsField() {
		return "", fmt.Errorf("`%s` is not statically typed", ident.String())
	}

	return vr.Name(), nil
}

func (pc *ParserContext) ParseIdent(ident *ast.Ident, origin ast.Expr) (any, error) {
	data, ok := pc.Info.Uses[ident]
	if !ok {
		return "", fmt.Errorf("`%s` is not present", ident.String())
	}
	switch t := data.(type) {
	case *types.Const:
		{
			if t.Val().Kind() == constant.Int {
				return strconv.Atoi(t.Val().ExactString())
			}
			if t.Val().Kind() == constant.Float {
				return strconv.ParseFloat(t.Val().ExactString(), 64)
			}
			if t.Val().Kind() == constant.Bool {
				return t.Val().ExactString() == "true", nil
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

func (pc *ParserContext) ParseExpr(current ast.Expr, origin ast.Expr) (any, error) {
	switch t := current.(type) {
	case *ast.BasicLit:
		{
			return pc.ParseBasicLint(t)
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

func (r RawSchema) GetTitle() (*string, error) {
	return r.LookupCast[string]("Title")
}

func (r RawSchema) GetDescription() (*string, error) {
	return r.LookupCast[string]("Description")
}

func (r TypedParam) GetSchema() (*RawSchema, error) {
	return Cast[RawSchema](r.Value)
}

func (r TypedParam) GetRequired() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("Required")
}

func (r TypedParam) GetMinLen() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("MinLen")
}

func (r TypedParam) GetMaxLen() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("MaxLen")
}

func (r TypedParam) GetMin() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("Min")
}

func (r TypedParam) GetMax() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("Max")
}

func (r TypedParam) GetExclusiveMin() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("ExclusiveMin")
}

func (r TypedParam) GetExclusiveMax() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("ExclusiveMax")
}

func (r TypedParam) GetMultipleOf() (*float64, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[float64]("MultipleOf")
}

func (r TypedParam) GetPattern() (*string, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[string]("Pattern")
}

func (r TypedParam) GetMinItems() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("MinItems")
}

func (r TypedParam) GetMaxItems() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("MaxItems")
}

func (r TypedParam) GetUniqueItems() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("UniqueItems")
}

func (r TypedParam) GetMinProperties() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("MinProperties")
}

func (r TypedParam) GetMaxProperties() (*int, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[int]("MaxProperties")
}

func (r TypedParam) GetAdditionalItems() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("AdditionalItems")
}

func (r TypedParam) GetAdditionalProperties() (*bool, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	return innerSchema.LookupCast[bool]("AdditionalProperties")
}

func (r TypedParam) GetEnum() ([]any, error) {
	innerSchema, err := Cast[RawSchema](r.Value)
	if err != nil {
		return nil, err
	}
	enum, err := innerSchema.LookupCast[[]any]("Enum")
	if err != nil {
		return nil, err
	}
	if enum == nil {
		return nil, nil
	}
	return *enum, nil
}

func (r TypedParam) GetType() ([]string, error) {
	out := make([]string, 0)
	switch r.TypeName {
	case ScalarType:
		{
			break
		}
	case ObjectType:
		{
			return []string{"object"}, nil
		}
	case ArrayType:
		{
			return []string{"array"}, nil
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

	types, err := Cast[[]any](innerType.Value)
	if err != nil {
		return nil, err
	}
	for _, typ := range *types {
		str, err := Cast[string](typ)
		if err != nil {
			return nil, err
		}
		out = append(out, *str)
	}

	return out, nil
}

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

func (r RawSchema) GetItems() (Items, error) {
	typedParam, err := r.LookupCast[TypedParam]("Items")
	if err != nil {
		return nil, err
	}
	if typedParam == nil {
		return nil, nil
	}
	rawSchema, err := Cast[[]any](typedParam.Value)
	if err != nil {
		return nil, err
	}
	items := make(Items, 0)
	for _, value := range *rawSchema {
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

func Cast[T any](v any) (*T, error) {
	val, ok := v.(T)
	if !ok {
		var zero T
		return nil, fmt.Errorf("cannot convert %T to %T", v, zero)
	}
	return &val, nil
}

func (r RawSchema) LookupCast[T any](key string) (*T, error) {
	return LookupCast[T](key, r)
}

func LookupCast[T any, R any](key string, mapper map[string]R) (*T, error) {
	val, ok := mapper[key]
	if !ok {
		return nil, nil
	}
	return Cast[T](val)
}

func (r RawSchema) Lookup(key string) any {
	return r[key]
}
