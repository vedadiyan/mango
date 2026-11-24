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
	"go/token"
	"strings"
)

const (
	// schemaType is the fully qualified name of the type a schema
	// definition must return.
	schemaType = "github.com/vedadiyan/mango/static.Schema"
)

type (
	// Schema describes a schema definition found in source code.
	Schema struct {
		Package    string     // name of the package declaring the schema
		Name       string     // schema name
		TypedParam TypedParam // parsed schema literal
	}
)

// GetSchemas returns the schema definitions found in the parsed package.
//
// A schema definition is an exported function whose last statement returns
// a static.Schema composite literal, either alone or preceded by a string
// literal giving the schema's name:
//
//	func DefineUser() static.Schema { return static.Schema{...} }
//	func DefineUser() (string, static.Schema) { return "users", static.Schema{...} }
//
// Without an explicit name, the function name with any "Define" prefix
// removed is used.
func GetSchemas(ParserContext *ParserContext) ([]Schema, error) {
	fns, err := ParserContext.ExtractFuncs()
	if err != nil {
		return nil, err
	}
	out := make([]Schema, 0)
	for _, fn := range fns {
		// Only exported functions returning one or two values are candidates.
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

		// The definition is taken from the function's final return statement.
		lastStatement := fn.Body.List[len(fn.Body.List)-1]
		returnStatement, ok := lastStatement.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		definition := returnStatement.Results[0]
		name := strings.TrimPrefix(fn.Name.String(), "Define")

		// With two results, the first is the schema name.
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
