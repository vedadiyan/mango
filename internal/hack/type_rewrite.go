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

// Package hack rewrites Go types at runtime so that anonymous (embedded)
// struct fields are inlined for BSON encoding.
package hack

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"unsafe"
)

type (
	// Metadata is a marker interface used to attach struct tags to a type.
	// Fields whose type implements Metadata are excluded from encoding.
	Metadata interface {
		dummyMethod(Metadata)
	}

	// Rewriter holds the state of a single Rewrite call: circular-reference
	// tracking, the active policy, and deferred patch handlers.
	Rewriter struct {
		ld  LoopDetector
		lhp LoopHandlerPolicy
		lh  LoopHandler
	}

	// StructField describes a single field of a struct being rewritten.
	StructField struct {
		Field      *reflect.StructField
		FieldIndex int
		StructType *reflect.Type
		Rewriter   *Rewriter
	}

	// LoopHandlerPolicy controls how Rewrite handles circular type references.
	LoopHandlerPolicy int
)

const (
	// LoopHandlerPolicyBlock causes Rewrite to return an error when a
	// circular reference is found.
	LoopHandlerPolicyBlock LoopHandlerPolicy = 0

	// LoopHandlerPolicyIgnore replaces circular fields with an unexported
	// placeholder.
	LoopHandlerPolicyIgnore LoopHandlerPolicy = 1

	// LoopHandlerPolicyHandle replaces circular fields with a placeholder
	// that is patched to the rewritten type once rewriting completes.
	LoopHandlerPolicyHandle LoopHandlerPolicy = 3
)

var (
	metadataType = reflect.TypeFor[Metadata]()
	typeCache    = make(map[reflect.Type]reflect.Type) // guarded by mut
	nilType      = reflect.TypeFor[unsafe.Pointer]()   // placeholder for circular fields
	mut          sync.Mutex
)

// Rewrite returns a dynamically constructed equivalent of src in which
// anonymous struct fields are inlined. crhp determines how circular
// references are handled. Results are cached, and Rewrite is safe for
// concurrent use.
func Rewrite(src reflect.Type, crhp LoopHandlerPolicy) (reflect.Type, error) {
	mut.Lock()
	defer mut.Unlock()
	if value, ok := typeCache[src]; ok {
		return value, nil
	}

	w := &Rewriter{}
	w.ld = make(LoopDetector)
	w.lhp = crhp

	typ, err := w.typeRewrite(src)
	if err != nil {
		return nil, err
	}
	if err := w.lh.inject(); err != nil {
		return nil, err
	}
	typeCache[src] = typ
	return typ, nil
}

// typeRewrite recursively rewrites src, descending into slices, arrays and
// structs while preserving pointer indirection. The result is cached.
func (w *Rewriter) typeRewrite(src reflect.Type) (reflect.Type, error) {
	if value, ok := typeCache[src]; ok {
		return value, nil
	}

	indirectType, n := indirect(src)
	var out reflect.Type

	switch indirectType.Kind() {
	case reflect.Slice:
		{
			res, err := w.typeRewrite(indirectType.Elem())
			if err != nil {
				return nil, err
			}
			out = reflect.SliceOf(res)
		}
	case reflect.Array:
		{
			res, err := w.typeRewrite(indirectType.Elem())
			if err != nil {
				return nil, err
			}
			out = reflect.ArrayOf(indirectType.Len(), res)
		}
	case reflect.Struct:
		{
			res, err := w.structRewrite(indirectType)
			if err != nil {
				return nil, err
			}
			out = res
		}
	default:
		{
			out = indirectType
		}
	}

	if n == 0 {
		typeCache[src] = out
		return out, nil
	}
	ptr := pointerTo(out, n)
	typeCache[src] = ptr
	return ptr, nil
}

// structRewrite rebuilds the struct type src, rewriting each field with
// StructField.Build. src is tracked for circular-reference detection while
// its fields are processed.
func (w *Rewriter) structRewrite(src reflect.Type) (reflect.Type, error) {
	if value, ok := typeCache[src]; ok {
		return value, nil
	}

	var finalResult reflect.Type
	out := make([]reflect.StructField, 0)
	w.ld.track(src, &finalResult)
	defer w.ld.release(src)

	for i := range src.NumField() {
		field := src.Field(i)
		structField := StructField{&field, i, &finalResult, w}
		f, err := structField.Build()
		if err != nil {
			return nil, err
		}
		out = append(out, *f)
	}
	finalResult = reflect.StructOf(out)
	typeCache[src] = finalResult
	return finalResult, nil
}

// Build returns the rewritten field. Fields whose bson tag includes the
// "opaque" option are returned unchanged. Fields that refer to a struct
// currently being rewritten are handled according to the Rewriter's
// CircRefHndlrPolicy.
func (sf *StructField) Build() (*reflect.StructField, error) {
	// Opaque fields are left unchanged.
	if tag, ok := sf.Field.Tag.Lookup("bson"); ok && isOpaque(tag) {
		return sf.Field, nil
	}

	// Not a circular reference: rewrite the field type normally.
	val, ok := sf.Rewriter.ld.has(concreteType(sf.Field.Type))
	if !ok {
		res, err := sf.Rewriter.typeRewrite(sf.Field.Type)
		if err != nil {
			return nil, err
		}
		sf.Field.Type = res
		sf.Field.Tag = getTagValue(sf.Field)
		return sf.Field, nil
	}

	// Circular reference: apply the configured policy.
	if sf.Rewriter.lhp == LoopHandlerPolicyBlock {
		return nil, fmt.Errorf("circular references are not allowed")
	}

	// Defer patching the field type until all types have been built.
	if sf.Rewriter.lhp == LoopHandlerPolicyHandle {
		ref := sf.Field.Type
		sf.Rewriter.lh = append(sf.Rewriter.lh, func() error {
			return abiPatch(sf.StructType, sf.FieldIndex, ref, val)
		})
	}

	// Hide the field by making it blank and unexported.
	if sf.Rewriter.lhp == LoopHandlerPolicyIgnore {
		sf.Field.Name = "_"
		sf.Field.PkgPath = "_"
	}

	sf.Field.Type = nilType
	return sf.Field, nil
}

// getTagValue returns the struct tag for a rewritten field. In order of
// precedence:
//   - the tag supplied through Metadata, for non-anonymous fields;
//   - `bson:"-"` for Metadata fields and blank (_) fields;
//   - `bson:",inline"` for anonymous fields;
//   - otherwise, the field's original tag.
func getTagValue(field *reflect.StructField) reflect.StructTag {
	// A Metadata tag overrides the field's own tag, allowing tags in places
	// the Go language does not permit them.
	if tagValue, ok := isMetadata(field.Type); ok && !field.Anonymous {
		return tagValue
	}

	// Metadata fields are used for tagging only and are never encoded.
	if field.Type.Implements(metadataType) || field.Name == "_" {
		return reflect.StructTag(`bson:"-"`)
	}

	if field.Anonymous {
		return reflect.StructTag(`bson:",inline"`)
	}

	return field.Tag
}

// isMetadata reports whether src, or the element type of a pointer, slice or
// array, embeds a Metadata type. If so, it returns the tag declared on that
// type's Metadata field.
func isMetadata(src reflect.Type) (reflect.StructTag, bool) {
	switch src.Kind() {
	case reflect.Pointer:
		{
			return isMetadata(src.Elem())
		}
	case reflect.Slice, reflect.Array:
		{
			return isMetadata(src.Elem())
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

// isOpaque reports whether the bson tag value str includes the "opaque" option.
func isOpaque(str string) bool {
	return slices.Contains(strings.Split(str, ","), "opaque")
}
