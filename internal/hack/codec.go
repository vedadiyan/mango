package hack

import (
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	AutoInlinerCodec[T any] struct {
		srcType     reflect.Type
		dynamicType reflect.Type
	}
)

var (
	bsonRaw = reflect.TypeFor[bson.Raw]()
)

func RegisterCodec[T any](registry *bson.Registry) error {
	alc, err := NewAutoInlinerCodec[T]()
	if err != nil {
		return err
	}

	encoder, err := registry.LookupEncoder(bsonRaw)
	if err != nil {
		return err
	}

	decoder, err := registry.LookupDecoder(bsonRaw)
	if err != nil {
		return err
	}

	registry.RegisterTypeEncoder(alc.srcType, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			val := reflect.New(alc.srcType)
			val.Elem().Set(v)
			out, err := alc.EncodeValue(val)
			if err != nil {
				return err
			}
			return encoder.EncodeValue(ec, vw, reflect.ValueOf(bson.Raw(out)))
		},
	))

	registry.RegisterTypeDecoder(alc.srcType, bson.ValueDecoderFunc(
		func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
			raw := bson.Raw{}
			if err := decoder.DecodeValue(dc, vr, reflect.ValueOf(&raw).Elem()); err != nil {
				return err
			}
			out, err := alc.DecodeValue(raw)
			if err != nil {
				return err
			}
			v.Set(out.Elem())
			return nil
		},
	))

	return nil
}

func NewAutoInlinerCodec[T any]() (*AutoInlinerCodec[T], error) {
	srcType := reflect.TypeFor[T]()
	dynamicType, err := Rewrite(srcType)
	if err != nil {
		return nil, err
	}
	return &AutoInlinerCodec[T]{
		srcType:     srcType,
		dynamicType: dynamicType,
	}, nil
}

func (x *AutoInlinerCodec[T]) Encode(value *T) ([]byte, error) {
	out := convert(value, x.dynamicType)
	return bson.Marshal(out.Interface())
}

func (x *AutoInlinerCodec[T]) EncodeValue(v reflect.Value) ([]byte, error) {
	out := convertValue(v, x.dynamicType)
	return bson.Marshal(out.Interface())
}

func (x *AutoInlinerCodec[T]) Decode(value []byte) (*T, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.Unmarshal(value, v); err != nil {
		return nil, err
	}
	out := convert(v, x.srcType).Interface()
	return out.(*T), nil
}

func (x *AutoInlinerCodec[T]) DecodeValue(value []byte) (reflect.Value, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.Unmarshal(value, v); err != nil {
		return reflect.Value{}, err
	}
	out := convert(v, x.srcType)
	return out, nil
}

func convert(v any, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, reflect.ValueOf(v).UnsafePointer())
}

func convertValue(v reflect.Value, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, v.UnsafePointer())
}
