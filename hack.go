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
)

var (
	metadataType = reflect.TypeFor[Metadata]()
)

type Metadata interface {
	getMetadata(Metadata)
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
	return bson.MarshalExtJSON(out.Interface(), true, true)
}

func (x *AutoInlinerCodec[T]) Decode(value []byte) (*T, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.UnmarshalExtJSON(value, true, v); err != nil {
		return nil, err
	}
	out := Convert(v, x.srcType).Interface().(T)
	return &out, nil
}

func AutoInliner(src reflect.Type) reflect.Type {
	if src.Kind() != reflect.Struct {
		return src
	}

	out := make([]reflect.StructField, 0)
	for field := range src.Fields() {
		if !field.Anonymous {
			tag := field.Tag
			if field.Type.Implements(metadataType) {
				for f := range field.Type.Fields() {
					if f.Type.Implements(metadataType) {
						tag = f.Type.Field(0).Tag
						break
					}
				}
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

func Convert(v any, typ reflect.Type) reflect.Value {
	src := reflect.ValueOf(v)

	ptr := unsafe.Pointer(src.Pointer())
	return reflect.NewAt(typ, ptr).Elem()
}
