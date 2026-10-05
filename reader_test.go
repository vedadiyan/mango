package mango

import (
	"fmt"
	"reflect"
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
	for _, i := range schemas {
		value, err := GetGoTypes(i.TypedParam, []string{i.Name})
		if err != nil {
			t.FailNow()
		}
		fmt.Printf("%v", value)
		_ = value
	}

	t1 := AutoInliner(reflect.TypeFor[X]())
	test := X{}
	test.Fields.Name = "Pouya"
	test.Fields.Username = "VPouya"

	vallll := Convert(&test, t1).Interface()

	_ = vallll

	codec := NewAutoInlinerCodec[X]()

	outtt, err := codec.Encode(&test)

	xxxxx, err := codec.Decode(outtt)

	_ = xxxxx
	fmt.Println(string(outtt))
}

type Y struct {
	Name string `bson:"Name"`
}

type Z struct {
	Username string `bson:"Username"`
}

type Metadata interface {
	getMetadata(Metadata)
}

type FieldsMetadata struct {
	_x bool `Values`
}

func (x *FieldsMetadata) getMetadata(Metadata) {}

type X struct {
	Fields struct {
		FieldsMetadata
		Y
		Z
	}
}
