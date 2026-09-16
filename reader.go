package mango

import (
	"fmt"
	"go/ast"
)

const (
	schemaType = "github.com/vedadiyan/mango/static.Schema"
)

func GetSchemas(ParserContext *ParserContext) (any, error) {
	fns, err := ParserContext.ExtractFuncs()
	if err != nil {
		return nil, err
	}
	out := make([]TypedParam, 0)
	for _, fn := range fns {
		ident, ok := ParserContext.Info.Defs[fn.Name]
		if !ok {
			continue
		}
		if !ident.Exported() {
			continue
		}
		res := fn.Type.Results
		if res.NumFields() != 1 {
			continue
		}

		lastStatement := fn.Body.List[len(fn.Body.List)-1]
		returnStatement, ok := lastStatement.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		compositLit, ok := returnStatement.Results[0].(*ast.CompositeLit)
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
		out = append(out, mapperValue)
	}

	return out, nil
}
