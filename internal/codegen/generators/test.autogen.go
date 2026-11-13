package generators

import (
	"fmt"
	"github.com/vedadiyan/mango/tools"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore"
)

// +--------------------------------------------------+
// | Begin Type Definition For: Export
// +--------------------------------------------------+

type Export bson.D

type ExportType[T any] struct {
	Root T `bson:"inline"`
}

// +--------------------------------------------------+
// |    Begin Field Definition For: Products
// +--------------------------------------------------+

type ExportProductsSpecs[T any] struct {
	Products []*[]*ExportProductsType[T] `bson:"Products,omitempty"`
}

var _ExportProductsName = tools.FieldName[ExportProductsSpecs[any]]()

func ExportProductsName() string {
	return _ExportProductsName
}

func (x Export) SetExportProducts(v []*[]*ExportProducts) (Export, error) {
	for i := range x {
		if x[i].Key == ExportProductsName() {
			x[i].Value = v
			return x, nil
		}
	}
	return append(x, bson.E{Key: ExportProductsName(), Value: v}), nil
}

func (x Export) SetExportProductsIfPresent(v []*[]*ExportProducts) (Export, error) {
	if v == nil {
		return x, nil
	}

	for i := range x {
		if x[i].Key == ExportProductsName() {
			x[i].Value = v
			return x, nil
		}
	}
	return append(x, bson.E{Key: ExportProductsName(), Value: v}), nil
}

func (x Export) GetExportProducts() ([]*[]*ExportProducts, error) {
	for _, i := range x {
		if i.Key != ExportProductsName() {
			continue
		}
		switch t := i.Value.(type) {
		case []*[]*ExportProducts:
			{
				return t, nil
			}
		case nil:
			{
				return nil, nil
			}
		default:
			{
				var out []*[]*ExportProducts
				if err := tools.Copy(t, &out); err != nil {
					return nil, err
				}
				return out, nil
			}
		}
	}

	return nil, bsoncore.ErrElementNotFound
}

func ExportProductsBsonE(v []*[]*ExportProducts) bson.E {
	return bson.E{Key: ExportProductsName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Products
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: Export
// +--------------------------------------------------+

// +--------------------------------------------------+
// | Begin Type Definition For: ExportProducts
// +--------------------------------------------------+

type ExportProducts bson.D

type ExportProductsType[T any] struct {
	Root T `bson:"inline"`
}

// +--------------------------------------------------+
// |    Begin Field Definition For: Name
// +--------------------------------------------------+

type ExportProductsNameSpecs struct {
	Name string `bson:"Name"`
}

var _ExportProductsNameName = tools.FieldName[ExportProductsNameSpecs]()

func ExportProductsNameName() string {
	return _ExportProductsNameName
}

func (x ExportProducts) SetExportProductsName(v string) (ExportProducts, error) {
	for i := range x {
		if x[i].Key == ExportProductsNameName() {
			x[i].Value = v
			return x, nil
		}
	}
	return append(x, bson.E{Key: ExportProductsNameName(), Value: v}), nil
}

func (x ExportProducts) GetExportProductsName() (string, error) {
	var zero string
	for _, i := range x {
		if i.Key != ExportProductsNameName() {
			continue
		}
		switch t := i.Value.(type) {
		case string:
			{
				return t, nil
			}
		default:
			{
				return zero, fmt.Errorf("unexpected BSON value type %T for field %s", t, ExportProductsNameName())
			}
		}
	}

	return zero, bsoncore.ErrElementNotFound
}

func ExportProductsNameBsonE(v string) bson.E {
	return bson.E{Key: ExportProductsNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Name
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportProducts
// +--------------------------------------------------+
