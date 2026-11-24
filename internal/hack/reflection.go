// Copyright 2026 Pouya Vedadiyan. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hack

import (
	"reflect"
	"slices"
)

// indirect strips all pointer indirection from src and returns the
// underlying type together with the number of pointer levels removed.
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

// pointerTo wraps src in n levels of pointer indirection. It is the inverse
// of indirect.
func pointerTo(src reflect.Type, n int) reflect.Type {
	typ := src
	for range n {
		typ = reflect.PointerTo(typ)
	}
	return typ
}

// bluePrint returns the chain of pointer, slice and array types that wrap
// t's concrete type, ordered from innermost to outermost. For example,
// []*T yields [*T, []*T].
func bluePrint(t reflect.Type) []reflect.Type {
	var out []reflect.Type
	for hasElem(t.Kind()) {
		out = append(out, t)
		t = t.Elem()
	}
	slices.Reverse(out)
	return out
}

// concreteType returns the innermost element type of t, stripping all
// pointer, slice and array wrappers.
func concreteType(t reflect.Type) reflect.Type {
	for hasElem(t.Kind()) {
		t = t.Elem()
	}
	return t
}

// hasElem reports whether k is a kind that wraps an element type and is
// unwrapped by concreteType: pointer, array or slice.
func hasElem(k reflect.Kind) bool {
	return k == reflect.Pointer || k == reflect.Array || k == reflect.Slice
}
