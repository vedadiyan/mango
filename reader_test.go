package mango

import (
	"testing"
)

func TestRead(t *testing.T) {
	pc, err := Parse("./schema_test.go")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	schemas, err := GetSchemas(pc)
	for _, i := range schemas.([]RawSchema) {
		props, err := i.GetProperties()
		if err != nil {
			t.FailNow()
		}
		for _, value := range *props {
			schema, err := value.GetSchema()
			if err != nil {
				t.FailNow()
			}
			title, err := schema.GetTitle()
			if err != nil {
				t.FailNow()
			}
			_ = title
			typ, err := value.GetType()
			if err != nil {
				t.FailNow()
			}
			if typ[0] == "object" {
				xxx, err := schema.GetProperties()
				if err != nil {
					t.FailNow()
				}
				_ = xxx
			}
			_ = typ
			required, err := value.GetRequired()
			if err != nil {
				t.FailNow()
			}
			_ = required
		}
	}
}
