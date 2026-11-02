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

func ToBsonArray[T ~[]E, E any](in T) (bson.A, error) {
	return toBsonArray(reflect.ValueOf(in))
}

func toBsonArray(in reflect.Value) (bson.A, error) {
	if k := in.Kind(); k != reflect.Array && k != reflect.Slice {
		return nil, fmt.Errorf("expected array but found non-array type")
	}
	l := in.Len()
	out := make(bson.A, l)
	isDimension := false
	if k := in.Type().Elem().Kind(); k == reflect.Array || k == reflect.Slice {
		isDimension = true
	}

	for i := range l {
		v := in.Index(i)
		if isDimension {
			res, err := toBsonArray(v)
			if err != nil {
				return nil, err
			}
			out[i] = res
			continue
		}

		out[i] = v.Interface()
	}

	return out, nil
}
