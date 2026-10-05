package mango

import (
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	AutoInlinerCodec[T any] struct {
		srcType     reflect.Type
		dynamicType reflect.Type
	}

	Metadata interface {
		dummyMethod(Metadata)
	}
)

var (
	metadataType = reflect.TypeFor[Metadata]()
)

func RegisterCodec[T any](registry *bson.Registry) error {
	alc := NewAutoInlinerCodec[T]()

	encoder, err := registry.LookupEncoder(reflect.TypeFor[bson.Raw]())
	if err != nil {
		return err
	}

	decoder, err := registry.LookupDecoder(reflect.TypeFor[bson.Raw]())
	if err != nil {
		return err
	}

	_ = decoder
	registry.RegisterTypeEncoder(alc.srcType, bson.ValueEncoderFunc(func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
		val := reflect.New(alc.srcType)
		val.Elem().Set(v)
		out, err := alc.EncodeValue(val)
		if err != nil {
			return err
		}
		return encoder.EncodeValue(ec, vw, reflect.ValueOf(bson.Raw(out)))
	}))

	registry.RegisterTypeDecoder(alc.srcType, bson.ValueDecoderFunc(func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
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
	}))

	return nil
}

func NewAutoInlinerCodec[T any]() *AutoInlinerCodec[T] {
	srcType := reflect.TypeFor[T]()
	dynamicType := AutoInliner(srcType)
	return &AutoInlinerCodec[T]{
		srcType:     srcType,
		dynamicType: dynamicType,
	}
}

func (x *AutoInlinerCodec[T]) Encode(value *T) ([]byte, error) {
	out := Convert(value, x.dynamicType)
	return bson.Marshal(out.Interface())
}

func (x *AutoInlinerCodec[T]) EncodeValue(v reflect.Value) ([]byte, error) {
	out := ConvertValue(v, x.dynamicType)
	return bson.Marshal(out.Interface())
}

func (x *AutoInlinerCodec[T]) Decode(value []byte) (*T, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.Unmarshal(value, v); err != nil {
		return nil, err
	}
	out := Convert(v, x.srcType).Interface()
	return out.(*T), nil
}

func (x *AutoInlinerCodec[T]) DecodeValue(value []byte) (reflect.Value, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.Unmarshal(value, v); err != nil {
		return reflect.Value{}, err
	}
	out := Convert(v, x.srcType)
	return out, nil
}

func AutoInliner(src reflect.Type) reflect.Type {
	indirectType, n := Indirect(src)
	var out reflect.Type

	switch indirectType.Kind() {
	case reflect.Slice:
		{
			out = reflect.SliceOf(AutoInliner(indirectType.Elem()))
		}
	case reflect.Array:
		{
			out = reflect.ArrayOf(indirectType.Len(), AutoInliner(indirectType.Elem()))
		}
	case reflect.Struct:
		{
			out = RewriteStruct(indirectType)
		}
	default:
		{
			out = indirectType
		}
	}

	if n == 0 {
		return out
	}
	return PointerTo(out, n)
}

func RewriteStruct(src reflect.Type) reflect.Type {
	out := make([]reflect.StructField, 0)
	for field := range src.Fields() {
		if !field.Anonymous {
			tag := field.Tag
			if tagValue, ok := IsMetadata(field.Type); ok {
				tag = tagValue
			}
			out = append(out, reflect.StructField{
				Name:      field.Name,
				Anonymous: false,
				Index:     field.Index,
				Offset:    field.Offset,
				PkgPath:   field.PkgPath,
				Type:      AutoInliner(field.Type),
				Tag:       tag,
			})
			continue
		}
		if field.Type.Implements(metadataType) {
			out = append(out, reflect.StructField{
				Name:      field.Name,
				Anonymous: false,
				Index:     field.Index,
				Offset:    field.Offset,
				PkgPath:   field.PkgPath,
				Type:      AutoInliner(field.Type),
				Tag:       reflect.StructTag(`bson:"-"`),
			})
			continue
		}
		out = append(out, reflect.StructField{
			Name:      field.Name,
			Anonymous: true,
			Index:     field.Index,
			Offset:    field.Offset,
			PkgPath:   field.PkgPath,
			Type:      AutoInliner(field.Type),
			Tag:       reflect.StructTag(`bson:",inline"`),
		})

	}
	return reflect.StructOf(out)
}

func Indirect(src reflect.Type) (reflect.Type, int) {
	current := src
	n := 0
	if src.Kind() == reflect.Pointer {
		for current.Kind() == reflect.Pointer {
			current = current.Elem()
			n++
		}
	}
	return current, n
}

func PointerTo(src reflect.Type, n int) reflect.Type {
	typ := src
	for range n {
		typ = reflect.PointerTo(typ)
	}
	return typ
}

func IsMetadata(src reflect.Type) (reflect.StructTag, bool) {
	switch src.Kind() {
	case reflect.Pointer:
		{
			return IsMetadata(src.Elem())
		}
	case reflect.Slice, reflect.Array:
		{
			return IsMetadata(src.Elem())
		}
	case reflect.Struct:
		{
			for f := range src.Fields() {
				if !f.Anonymous || !f.Type.AssignableTo(metadataType) {
					continue
				}
				for innerField := range f.Type.Fields() {
					if innerField.Type.Implements(metadataType) {
						return innerField.Tag, true
					}
				}
			}
			return "", false
		}
	default:
		{
			return "", false
		}
	}
}

func Convert(v any, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, reflect.ValueOf(v).UnsafePointer())
}

func ConvertValue(v reflect.Value, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, v.UnsafePointer())
}
