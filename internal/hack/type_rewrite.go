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
		loopDetector      LoopDetector
		loopHandlerPolicy LoopHandlerPolicy
		loopHandler       LoopHandler
		cleanUp           map[reflect.Type]LoopHandlerPolicy
	}

	// StructField describes a single field of a struct being rewritten.
	StructField struct {
		field      *reflect.StructField
		fieldIndex int
		structType *reflect.Type
		rewriter   *Rewriter
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
	typeCache    = make(map[reflect.Type]map[LoopHandlerPolicy]reflect.Type) // guarded by mut
	nilType      = reflect.TypeFor[unsafe.Pointer]()                         // placeholder for circular fields
	mut          sync.Mutex
)

// Rewrite returns a dynamically constructed equivalent of src in which
// anonymous struct fields are inlined. crhp determines how circular
// references are handled. Results are cached, and Rewrite is safe for
// concurrent use.
func Rewrite(src reflect.Type, crhp LoopHandlerPolicy) (reflect.Type, error) {
	mut.Lock()
	defer mut.Unlock()
	value, ok := typeCache[src]
	if !ok {
		value = make(map[LoopHandlerPolicy]reflect.Type)
		typeCache[src] = value
	} else {
		if out, ok := value[crhp]; ok {
			return out, nil
		}
	}

	w := &Rewriter{}
	w.cleanUp = make(map[reflect.Type]LoopHandlerPolicy)
	w.loopDetector = make(LoopDetector)
	w.loopHandlerPolicy = crhp

	typ, err := w.typeRewrite(src, crhp)
	if err != nil {
		w.cleanTypeCache()
		return nil, err
	}
	if err := w.loopHandler.inject(); err != nil {
		w.cleanTypeCache()
		return nil, err
	}
	typeCache[src][crhp] = typ
	return typ, nil
}

func (w *Rewriter) cleanTypeCache() {
	for key, value := range w.cleanUp {
		delete(typeCache[key], value)
		if len(typeCache[key]) == 0 {
			delete(typeCache, key)
		}
	}
}

// typeRewrite recursively rewrites src, descending into slices, arrays and
// structs while preserving pointer indirection. The result is cached.
func (w *Rewriter) typeRewrite(src reflect.Type, crhp LoopHandlerPolicy) (reflect.Type, error) {
	value, ok := typeCache[src]
	if !ok {
		value = make(map[LoopHandlerPolicy]reflect.Type)
		typeCache[src] = value
	} else {
		if out, ok := value[crhp]; ok {
			return out, nil
		}
	}

	indirectType, n := indirect(src)
	var out reflect.Type

	switch indirectType.Kind() {
	case reflect.Slice:
		{
			res, err := w.typeRewrite(indirectType.Elem(), crhp)
			if err != nil {
				return nil, err
			}
			out = reflect.SliceOf(res)
		}
	case reflect.Array:
		{
			res, err := w.typeRewrite(indirectType.Elem(), crhp)
			if err != nil {
				return nil, err
			}
			out = reflect.ArrayOf(indirectType.Len(), res)
		}
	case reflect.Struct:
		{
			res, err := w.structRewrite(indirectType, crhp)
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

	w.cleanUp[src] = crhp
	if n == 0 {
		typeCache[src][crhp] = out
		return out, nil
	}
	ptr := pointerTo(out, n)
	typeCache[src][crhp] = ptr
	return ptr, nil
}

// structRewrite rebuilds the struct type src, rewriting each field with
// StructField.Build. src is tracked for circular-reference detection while
// its fields are processed.
func (w *Rewriter) structRewrite(src reflect.Type, crhp LoopHandlerPolicy) (reflect.Type, error) {
	value, ok := typeCache[src]
	if !ok {
		value = make(map[LoopHandlerPolicy]reflect.Type)
		typeCache[src] = value
	} else {
		if out, ok := value[crhp]; ok {
			return out, nil
		}
	}

	var finalResult reflect.Type
	out := make([]reflect.StructField, 0)
	w.loopDetector.track(src, &finalResult)
	defer w.loopDetector.release(src)

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
	w.cleanUp[src] = crhp
	typeCache[src][crhp] = finalResult
	return finalResult, nil
}

// Build returns the rewritten field. Fields whose bson tag includes the
// "opaque" option are returned unchanged. Fields that refer to a struct
// currently being rewritten are handled according to the Rewriter's
// CircRefHndlrPolicy.
func (sf *StructField) Build() (*reflect.StructField, error) {
	// Opaque fields are left unchanged.
	if tag, ok := sf.field.Tag.Lookup("bson"); ok && isOpaque(tag) {
		return sf.field, nil
	}

	// Not a circular reference: rewrite the field type normally.
	val, ok := sf.rewriter.loopDetector.has(concreteType(sf.field.Type))
	if !ok {
		sf.field.Anonymous, sf.field.Tag = getAnonymityAndTag(sf.field)
		res, err := sf.rewriter.typeRewrite(sf.field.Type, sf.rewriter.loopHandlerPolicy)
		if err != nil {
			return nil, err
		}
		sf.field.Type = res
		return sf.field, nil
	}

	// Circular reference: apply the configured policy.
	if sf.rewriter.loopHandlerPolicy == LoopHandlerPolicyBlock {
		return nil, fmt.Errorf("circular references are not allowed")
	}

	// Defer patching the field type until all types have been built.
	if sf.rewriter.loopHandlerPolicy == LoopHandlerPolicyHandle {
		ref := sf.field.Type
		sf.rewriter.loopHandler = append(sf.rewriter.loopHandler, func() error {
			return abiPatch(sf.structType, sf.fieldIndex, ref, val)
		})
	}

	// Hide the field by making it blank and unexported.
	if sf.rewriter.loopHandlerPolicy == LoopHandlerPolicyIgnore {
		sf.field.Name = "_"
		sf.field.PkgPath = "_"
	}

	sf.field.Type = typeCascade(sf.field.Type, nilType)

	return sf.field, nil
}

// getAnonymityAndTag returns the struct tag for a rewritten field. In order of
// precedence:
//   - the tag supplied through Metadata, for non-anonymous fields;
//   - `bson:"-"` for Metadata fields and blank (_) fields;
//   - `bson:",inline"` for anonymous fields;
//   - otherwise, the field's original tag.
func getAnonymityAndTag(field *reflect.StructField) (bool, reflect.StructTag) {
	// A Metadata tag overrides the field's own tag, allowing tags in places
	// the Go language does not permit them.
	if tagValue, ok := isMetadata(field.Type); ok && !field.Anonymous {
		return false, tagValue
	}

	// Metadata fields are used for tagging only and are never encoded.
	if field.Type.Implements(metadataType) || field.Name == "_" {
		return false, reflect.StructTag(`bson:"-"`)
	}

	if field.Anonymous {
		return true, reflect.StructTag(`bson:",inline"`)
	}

	return field.Anonymous, field.Tag
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
				if f.Anonymous && f.Type == metadataType {
					return f.Tag, true
				}
				if f.Type.Implements(metadataType) {
					tag, ok := isMetadata(f.Type)
					if ok {
						return tag, ok
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
