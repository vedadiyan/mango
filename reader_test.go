package mango

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRead(t *testing.T) {
	pc, err := Parse("./schema_test.go")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	schemas, err := GetSchemas(pc)
	for _, i := range schemas.([]TypedParam) {
		value, err := ToBsonSchema(i)
		if err != nil {
			t.FailNow()
		}
		out, err := json.MarshalIndent(value, "", "\t")
		if err != nil {
			t.FailNow()
		}
		fmt.Println(string(out))
		_ = value
	}
}
