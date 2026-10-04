package mango

import (
	"fmt"
	"testing"

	"github.com/iancoleman/strcase"
)

func TestRead(t *testing.T) {

	x := strcase.ToSnake("FirstName")

	_ = x
	pc, err := Parse("./schema_test.go")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	schemas, err := GetSchemas(pc)
	for _, i := range schemas.([]TypedParam) {
		value, err := GetGoTypes(i, []string{"Root"})
		if err != nil {
			t.FailNow()
		}
		fmt.Printf("%v", value)
		_ = value
	}
}
