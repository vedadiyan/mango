package mango

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/iancoleman/strcase"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

	registry := bson.NewRegistry()
	RegisterCodec[X](registry)

	client, _ := mongo.Connect(options.Client().ApplyURI("mongodb://192.168.100.100:27017").SetRegistry(registry))

	t1 := AutoInliner(reflect.TypeFor[X]())
	test := X{}
	test.Fields = &struct {
		FieldsMetadata
		Y
		Z
	}{}
	test.Fields.Name = "Pouya"
	test.Fields.Username = "VPouya"

	zzzzz, err := client.Database("abc").Collection("test").InsertOne(context.Background(), test)

	_ = zzzzz
	vallll := Convert(&test, t1).Interface()

	_ = vallll

	// outtt, err := codec.Encode(&test)

	// xxxxx, err := codec.Decode(outtt)

	// _ = xxxxx
	// fmt.Println()
	// fmt.Println(string(outtt))
}

type Y struct {
	Name string `bson:"Name"`
}

type Z struct {
	Username string `bson:"Username"`
}

type FieldsMetadata struct {
	Metadata `bson:"Values"`
}

type X struct {
	Fields *struct {
		FieldsMetadata
		Y
		Z
	}
}
