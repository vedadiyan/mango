package generators

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/vedadiyan/mango/internal/codegen"
)

func TestRead(t *testing.T) {

	x := strcase.ToSnake("FirstName")

	_ = x
	pc, err := codegen.Parse("./schema_test.go")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	schemas, err := codegen.GetSchemas(pc)
	for _, i := range schemas {
		value, err := ToGoTypeModel(i.TypedParam, []string{i.Name}, false)
		if err != nil {
			t.FailNow()
		}
		GoGenRender(value)
		json, err := json.MarshalIndent(value, "", "\t")
		if err != nil {
			t.FailNow()
		}
		os.WriteFile("schema.json", json, os.ModePerm)
		fmt.Printf("%s\n", string(json))
		_ = value
	}

}

type OneOfVariation1 struct {
	CustomerName string
	Email        string
}

type OneOfVariation2 struct {
	CompanyName string
	TaxId       string
}

type UnionType struct {
	OneOfVariation1
	OneOfVariation2
}

type EitherOf struct {
}
