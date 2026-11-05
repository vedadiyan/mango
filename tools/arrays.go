package tools

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Comparable interface {
	string | int | int16 | int32 | int64 | int8 | uint | uint16 | uint32 | uint64 | uint8 | bool | float32 | float64
}

func buildArrayType[T any](dim int) reflect.Type {
	typ := reflect.TypeFor[T]()
	for range dim {
		typ = reflect.SliceOf(typ)
	}
	return typ
}

func ReadDocumentArray[T ~[]bson.E](in bson.A, dims int) (any, error) {
	res, err := readDocumentArray[T](in, buildArrayType[T](dims))
	if err != nil {
		return nil, err
	}
	return res.Interface(), nil
}

func ReadPrimitiveArray[T Comparable](in bson.A, dims int) (any, error) {
	res, err := readPrimitiveArray[T](in, buildArrayType[T](dims))
	if err != nil {
		return nil, err
	}
	return res.Interface(), nil
}

func ReadOptionalPrimitiveArray[T *E, E Comparable](in bson.A, dims int) (any, error) {
	res, err := readOptionalPrimitiveArray[T](in, buildArrayType[T](dims))
	if err != nil {
		return nil, err
	}
	return res.Interface(), nil
}

func readDocumentArray[T ~[]bson.E](in bson.A, typ reflect.Type) (reflect.Value, error) {
	out := reflect.MakeSlice(typ, len(in), len(in))
	for x, i := range in {
		switch t := i.(type) {
		case bson.A:
			{
				res, err := readDocumentArray[T](t, typ.Elem())
				if err != nil {
					return reflect.ValueOf(nil), err
				}
				out.Index(x).Set(res)
			}
		case nil:
			{
				continue
			}
		case bson.D:
			{
				out.Interface().([]T)[x] = T(t)
			}
		case []bson.E:
			{
				out.Interface().([]T)[x] = T(t)
			}
		case T:
			{
				out.Interface().([]T)[x] = t
			}
		default:
			{
				return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
			}
		}
	}
	return out, nil
}

func readPrimitiveArray[T Comparable](in bson.A, typ reflect.Type) (reflect.Value, error) {
	out := reflect.MakeSlice(typ, len(in), len(in))
	for x, i := range in {
		switch t := i.(type) {
		case bson.A:
			{
				res, err := readPrimitiveArray[T](t, typ.Elem())
				if err != nil {
					return reflect.ValueOf(nil), err
				}
				out.Index(x).Set(res)
			}
		case nil:
			{
				continue
			}
		case T:
			{
				out.Interface().([]T)[x] = t
			}
		default:
			{
				val, ok := any(&t).(T)
				if !ok {
					return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
				}
				out.Interface().([]T)[x] = val
			}
		}
	}
	return out, nil
}

func readOptionalPrimitiveArray[T *E, E Comparable](in bson.A, typ reflect.Type) (reflect.Value, error) {
	out := reflect.MakeSlice(typ, len(in), len(in))
	for x, i := range in {
		switch t := i.(type) {
		case bson.A:
			{
				res, err := readOptionalPrimitiveArray[T](t, typ.Elem())
				if err != nil {
					return reflect.ValueOf(nil), err
				}
				out.Index(x).Set(res)
			}
		case nil:
			{
				continue
			}
		case T:
			{
				out.Interface().([]T)[x] = t
			}
		case E:
			{
				out.Interface().([]T)[x] = &t
			}
		default:
			{
				val, ok := any(&t).(T)
				if !ok {
					return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
				}
				out.Interface().([]T)[x] = val
			}
		}
	}
	return out, nil
}

func WriteArray[T any](in any) (bson.A, error) {
	out, err := writeArray[T](reflect.ValueOf(in))
	if err != nil {
		return nil, err
	}
	res, ok := out.(bson.A)
	if !ok {
		return nil, fmt.Errorf("expected bson.A but found another type")
	}
	return res, nil
}

func writeArray[T any](in reflect.Value) (any, error) {
	if in.Type() == reflect.TypeFor[T]() {
		return *(*T)(in.Addr().UnsafePointer()), nil
	}

	if k := in.Kind(); k != reflect.Array && k != reflect.Slice {
		return nil, fmt.Errorf("expected array but found non-array type")
	}

	l := in.Len()
	out := make(bson.A, l)
	for i := range l {
		v := in.Index(i)
		res, err := writeArray[T](v)
		if err != nil {
			return nil, err
		}
		out[i] = res
	}
	return out, nil
}
