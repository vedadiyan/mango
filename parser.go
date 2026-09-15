package mango

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
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

func (pc *ParserContext) ParseBasicLint(bl *ast.BasicLit) (string, error) {
	switch bl.Kind {
	case token.STRING, token.INT, token.FLOAT, token.IMAG, token.CHAR:
		{
			return strings.TrimRight(strings.TrimLeft(bl.Value, "\""), "\""), nil
		}
	default:
		{
			return "", fmt.Errorf("`%s` is not statically typed", bl.Value)
		}
	}
}

func (pc *ParserContext) ParseVar(vr *types.Var, ident *ast.Ident, origin ast.Expr) (string, error) {
	if !vr.IsField() {
		return "", fmt.Errorf("`%s` is not statically typed", ident.String())
	}

	if origin == nil {
		return "", fmt.Errorf("`%s` cannot parse field without having reference to the struct", ident.String())
	}

	data, ok := pc.Info.Types[origin]
	if !ok {
		return "", fmt.Errorf("`%s` is not linked to a known type", ident.String())
	}

	strct, ok := data.Type.Underlying().(*types.Struct)
	if !ok {
		return "", fmt.Errorf("`%s` expected struct but found `%T`", ident.String(), data.Type.Underlying())
	}

	for i := range strct.NumFields() {
		if strct.Field(i).Name() != vr.Name() {
			continue
		}
		tag := strct.Tag(i)
		if len(tag) == 0 {
			break
		}
		nativeTag := reflect.StructTag(tag)
		val, ok := nativeTag.Lookup("json")
		if !ok {
			break
		}
		return strings.Split(val, ",")[0], nil
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
			return pc.ParseVar(t, ident, origin)
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

func (pc *ParserContext) ParseCompositLitAsMap(current *ast.CompositeLit) (map[string]any, error) {
	out := make(map[string]any)
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
			switch typ.Type.(type) {
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
