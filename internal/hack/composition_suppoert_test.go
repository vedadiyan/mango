package hack

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	IDField struct {
		Id bson.ObjectID `bson:",opaque"`
	}
	TestStruct struct {
		IDField
		Name     string
		UserName string
		Time     time.Time `bson:",opaque"`
	}
)

func (x *TestStruct) Test() {
	x.Name = "ok"
}

func (x *TestStruct) Test2() {
	x.UserName = "ok"
}

func TestAutoInliner(t *testing.T) {
	typ, _ := AutoInliner(reflect.TypeFor[TestStruct]())
	sample := TestStruct{
		Name:     "Pouya",
		UserName: "VPouya",
		Id:       bson.NewObjectID(),
		Time:     time.Now(),
	}

	value := convert(&sample, typ).Interface()

	out, err := bson.MarshalExtJSON(value, true, true)
	if err != nil {
		t.FailNow()
	}

	back := reflect.New(typ).Interface()

	if err := bson.UnmarshalExtJSON(out, true, back); err != nil {
		t.FailNow()
	}

	back2 := convert(back, reflect.TypeFor[TestStruct]()).Interface()

	_ = back2
	fmt.Println(string(out))

	if !reflect.DeepEqual(sample, value) {
		t.FailNow()
	}
}

type TestRegistryName struct {
	Name string `bson:"name"`
}

type TestRegistryType struct {
	Type string `bson:"type"`
}

func TestCodec(t *testing.T) {
	r := bson.NewRegistry()
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://192.168.169.190:27017").SetRegistry(r))
	if err != nil {
		t.FailNow()
	}

	db := client.Database("test")
	col := db.Collection("test")
	query, err := col.Find(context.TODO(), bson.M{})
	if err != nil {
		t.FailNow()
	}

	type output struct {
		TestRegistryName
		TestRegistryType
		Recursive struct {
			Recursive **output
		}
	}

	var val output

	if err := RegisterCodec[output](r); err != nil {
		t.FailNow()
	}

	query.Next(context.TODO())
	if err := query.Decode(&val); err != nil {
		t.FailNow()
	}
	c := 10
	_ = c
}
