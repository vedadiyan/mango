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
	if src.Kind() == reflect.Pointer {
		return IsMetadata(src.Elem())
	}
	if f := src.Kind(); f == reflect.Slice || f == reflect.Array {
		return IsMetadata(src.Elem())
	}
	if src.Kind() != reflect.Struct {
		return "", false
	}
	for f := range src.Fields() {
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
