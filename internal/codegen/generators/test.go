package generators

import (
	"reflect"

	"github.com/vedadiyan/mango/tools"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore"
)

// +--------------------------------------------------+
// | Begin Type Definition For: Export
// +--------------------------------------------------+

type Export bson.D

// +--------------------------------------------------+
// |    Begin Field Definition For: Products
// +--------------------------------------------------+

type ExportProductsSpecs[T any] struct {
	Products [][]T `bson:"Products,omitempty"`
}

var _ExportProductsName = tools.FieldName[ExportProductsSpecs[any]]()

func ExportProductsName() string {
	return _ExportProductsName
}

func (x Export) SetExportProducts(v ExportProducts) Export {
	for i := range x {
		if x[i].Key == ExportProductsName() {
			x[i].Value = v
			return x
		}
	}
	return append(x, bson.E{Key: ExportProductsName(), Value: v})
}

func (x Export) GetExportProducts() (ExportProducts, error) {
	for _, i := range x {
		if i.Key != ExportProductsName() {
			continue
		}
		switch t := i.Value.(type) {
		case ExportProducts:
			{
				return t, nil
			}
		case bson.D:
			{
				return ExportProducts(t), nil
			}
		case bson.Raw:
			{
				var d bson.D
				if err := bson.Unmarshal(t, &d); err != nil {
					return nil, err
				}
				return ExportProducts(d), nil
			}
		case nil:
			{
				return nil, nil
			}
		default:
			{
				typ := reflect.TypeFor[ExportProducts]()
				return nil, bson.ValueDecoderError{
					Name:     ExportProductsName(),
					Types:    []reflect.Type{typ},
					Kinds:    []reflect.Kind{typ.Kind()},
					Received: reflect.ValueOf(t),
				}
			}
		}
	}

	return nil, bsoncore.ErrElementNotFound
}

func ExportProductsBsonE(v ExportProducts) bson.E {
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

// +--------------------------------------------------+
// |    Begin Field Definition For: Name
// +--------------------------------------------------+

type ExportProductsNameSpecs struct {
	Name *string `bson:"Name,omitempty"`
}

var _ExportProductsNameName = tools.FieldName[ExportProductsNameSpecs]()

func ExportProductsNameName() string {
	return _ExportProductsNameName
}

func (x ExportProducts) SetExportProductsName(v *string) ExportProducts {
	for i := range x {
		if x[i].Key == ExportProductsNameName() {
			x[i].Value = v
			return x
		}
	}
	return append(x, bson.E{Key: ExportProductsNameName(), Value: v})
}

func (x ExportProducts) GetExportProductsName() (*string, error) {
	for _, i := range x {
		if i.Key != ExportProductsNameName() {
			continue
		}

		switch t := i.Value.(type) {
		case *string:
			{
				return t, nil
			}
		case string:
			{
				return new(t), nil
			}
		case nil:
			{
				var zero *string
				return zero, nil
			}
		default:
			{
				var zero *string
				typ := reflect.TypeFor[*string]()
				return zero, bson.ValueDecoderError{
					Name:     ExportProductsNameName(),
					Types:    []reflect.Type{typ},
					Kinds:    []reflect.Kind{typ.Kind()},
					Received: reflect.ValueOf(t),
				}
			}
		}
	}

	var zero *string
	return zero, bsoncore.ErrElementNotFound
}

func ExportProductsNameBsonE(v *string) bson.E {
	return bson.E{Key: ExportProductsNameName(), Value: v}
}

// +--------------------------------------------------+
// |    End Field Definition For: Name
// +--------------------------------------------------+

// +--------------------------------------------------+
// | End Type Definition For: ExportProducts
// +--------------------------------------------------+
