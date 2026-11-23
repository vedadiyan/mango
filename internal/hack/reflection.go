package hack

import (
	"reflect"
	"slices"
)

func indirect(src reflect.Type) (reflect.Type, int) {
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

func pointerTo(src reflect.Type, n int) reflect.Type {
	typ := src
	for range n {
		typ = reflect.PointerTo(typ)
	}
	return typ
}

func bluePrint(t reflect.Type) []reflect.Type {
	var out []reflect.Type
	for hasElem(t.Kind()) {
		out = append(out, t)
		t = t.Elem()
	}
	slices.Reverse(out)
	return out
}

func concreteType(t reflect.Type) reflect.Type {
	for hasElem(t.Kind()) {
		t = t.Elem()
	}
	return t
}

func hasElem(k reflect.Kind) bool {
	return k == reflect.Pointer || k == reflect.Array || k == reflect.Slice
}
