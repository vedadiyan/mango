package codegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

const (
	schemaType = "github.com/vedadiyan/mango/static.Schema"
)

type (
	Schema struct {
		Package    string
		Name       string
		TypedParam TypedParam
	}
)

func GetSchemas(ParserContext *ParserContext) ([]Schema, error) {
	fns, err := ParserContext.ExtractFuncs()
	if err != nil {
		return nil, err
	}
	out := make([]Schema, 0)
	for _, fn := range fns {
		ident, ok := ParserContext.Info.Defs[fn.Name]
		if !ok {
			continue
		}
		if !ident.Exported() {
			continue
		}
		res := fn.Type.Results
		if res.NumFields() == 0 || res.NumFields() > 2 {
			continue
		}

		lastStatement := fn.Body.List[len(fn.Body.List)-1]
		returnStatement, ok := lastStatement.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		definition := returnStatement.Results[0]
		name := strings.TrimPrefix(fn.Name.String(), "Define")

		if len(returnStatement.Results) == 2 {
			val, ok := returnStatement.Results[0].(*ast.BasicLit)
			if !ok {
				return nil, fmt.Errorf("expected `BasicLit` but found `%T`", returnStatement.Results[0])
			}
			if val.Kind != token.STRING {
				return nil, fmt.Errorf("expected `string` but found `%s`", val.Kind)
			}
			name = val.Value
			definition = returnStatement.Results[1]
		}

		compositLit, ok := definition.(*ast.CompositeLit)
		if !ok {
			continue
		}
		parsedValue, err := ParserContext.ParseExpr(compositLit, nil)
		if err != nil {
			return nil, err
		}
		mapperValue, ok := parsedValue.(TypedParam)
		if !ok {
			return nil, fmt.Errorf("expected `map[string]any` but found %T", parsedValue)
		}
		if mapperValue.TypeName != schemaType {
			return nil, fmt.Errorf("expected `%s` but found %s", schemaType, mapperValue.TypeName)
		}
		out = append(out, Schema{ident.Pkg().Name(), name, mapperValue})
	}

	return out, nil
}
