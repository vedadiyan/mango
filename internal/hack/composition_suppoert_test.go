package hack

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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
	typ := AutoInliner(reflect.TypeFor[TestStruct]())
	sample := TestStruct{
		Name:     "Pouya",
		UserName: "VPouya",
		Id:       bson.NewObjectID(),
		Time:     time.Now(),
	}

	value := Convert(&sample, typ).Interface()

	out, err := bson.MarshalExtJSON(value, true, true)
	if err != nil {
		t.FailNow()
	}

	back := reflect.New(typ).Interface()

	if err := bson.UnmarshalExtJSON(out, true, back); err != nil {
		t.FailNow()
	}

	back2 := Convert(back, reflect.TypeFor[TestStruct]()).Interface()

	_ = back2
	fmt.Println(string(out))

	if !reflect.DeepEqual(sample, value) {
		t.FailNow()
	}
}
