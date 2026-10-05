package mango

import (
	"reflect"
	"unsafe"

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
	return bson.MarshalExtJSON(out.Interface(), true, true)
}

func (x *AutoInlinerCodec[T]) Decode(value []byte) (*T, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.UnmarshalExtJSON(value, true, v); err != nil {
		return nil, err
	}
	out := Convert(v, x.srcType).Interface()
	return out.(*T), nil
}

func AutoInliner(src reflect.Type) reflect.Type {
	if src.Kind() == reflect.Slice {
		return reflect.SliceOf(AutoInliner(src.Elem()))
	}

	if src.Kind() == reflect.Array {
		return reflect.ArrayOf(src.Len(), AutoInliner(src.Elem()))
	}

	if src.Kind() != reflect.Struct {
		return src
	}

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

func IsMetadata(field reflect.Type) (reflect.StructTag, bool) {
	if f := field.Kind(); f == reflect.Slice || f == reflect.Array {
		return IsMetadata(field.Elem())
	}
	if field.Kind() != reflect.Struct {
		return "", false
	}
	for f := range field.Fields() {
		if f.Anonymous && f.Type.AssignableTo(metadataType) {
			for innerField := range f.Type.Fields() {
				if innerField.Type.Implements(metadataType) {
					return innerField.Tag, true
				}
			}
		}
	}
	return "", false
}

func Convert(v any, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, unsafe.Pointer(reflect.ValueOf(v).Pointer()))
}
