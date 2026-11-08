package tools

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func WriteArray(in any) (bson.A, error) {
	val := reflect.ValueOf(in)
	typ := val.Type()

	if !isSliceOrArray(typ) {
		return nil, fmt.Errorf("expected slice type, got %s", typ)
	}

	for isSliceOrArray(typ) {
		typ = typ.Elem()
	}

	res, err := writeArray(val, typ)
	if err != nil {
		return nil, err
	}

	out, ok := res.(bson.A)
	if !ok {
		return nil, fmt.Errorf("could not cast %T to %s", res, typ.Name())
	}

	return out, nil
}

func writeArray(in reflect.Value, typ reflect.Type) (any, error) {
	if !in.IsValid() {
		return nil, nil
	}
	t := in.Type()
	if t == typ {
		return in.Interface(), nil
	}

	switch in.Kind() {
	case reflect.Pointer, reflect.Interface:
		{
			if in.IsNil() {
				return nil, nil
			}
			return writeArray(in.Elem(), typ)
		}

	case reflect.Array, reflect.Slice:
		out := make(bson.A, in.Len())

		if t.Elem() == typ {
			for i := range out {
				out[i] = in.Index(i).Interface()
			}
			return out, nil
		}

		for i := range out {
			res, err := writeArray(in.Index(i), typ)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			out[i] = res
		}
		return out, nil
	}

	return nil, fmt.Errorf("cannot use %s as %s or array of it", t, typ)
}

func ReadArray[T any](in bson.A) (T, error) {
	typ := reflect.TypeFor[T]()
	res, err := readArray(in, typ)
	if err != nil {
		return zero[T](), err
	}

	out, ok := reflect.TypeAssert[T](res)
	if !ok {
		return zero[T](), fmt.Errorf("could not cast %s to %s", res.Type().Name(), typ.Name())
	}

	return out, nil
}

func readArray(in bson.A, typ reflect.Type) (reflect.Value, error) {
	if typ.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("expected slice type, got %s", typ)
	}
	elemType := typ.Elem()
	out := reflect.MakeSlice(typ, len(in), len(in))

	for i, item := range in {
		if item == nil {
			continue
		}
		v, err := convertValue(item, elemType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("index %d: %w", i, err)
		}
		out.Index(i).Set(v)
	}
	return out, nil
}

func convertValue(item any, target reflect.Type) (reflect.Value, error) {
	if target.Kind() == reflect.Pointer {
		inner, err := convertValue(item, target.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		p := reflect.New(target.Elem())
		p.Elem().Set(inner)
		return p, nil
	}

	if arr, ok := item.(bson.A); ok && target.Kind() == reflect.Slice {
		return readArray(arr, target)
	}

	v := reflect.ValueOf(item)

	if v.Type().AssignableTo(target) {
		return v, nil
	}

	if v.Type().ConvertibleTo(target) {
		return v.Convert(target), nil
	}
	return reflect.Value{}, fmt.Errorf("cannot use %s as %s", v.Type(), target)
}

func zero[T any]() T {
	var zero T
	return zero
}

func isSliceOrArray(v reflect.Type) bool {
	k := v.Kind()
	return k == reflect.Slice || k == reflect.Array
}
