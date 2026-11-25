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
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var (
	abi     reflect.Type // runtime's internal/abi.StructType, discovered via reflection
	abiErr  error        // non-nil if ABI patching is disabled or unavailable
	abiMute sync.Mutex
)

func init() {
	EnableAbiPatch(false)
}

// EnableAbiPatch enables or disables runtime ABI patching, which
// CircRefHandlrPolicyHandle requires. When enabling, it locates and validates
// the runtime's internal struct type layout and returns an error if the
// layout is not supported by the running Go version. Patching is disabled
// by default.
//
// ABI patching modifies runtime type metadata through unsafe operations and
// depends on Go internals that may change between releases.
func EnableAbiPatch(b bool) error {
	abiMute.Lock()
	defer abiMute.Unlock()
	if b {
		abiErr = discoverABI()
		if abiErr == nil {
			abiErr = validateABI()
		}
		return abiErr
	}
	abiErr = fmt.Errorf("ABI is disabled")
	return nil
}

// discoverABI locates the runtime's internal/abi.StructType by following
// reflect's rtype to the return type of its StructType method.
func discoverABI() error {
	target := reflect.TypeOf(reflect.TypeFor[struct{}]()).Elem()
	if target.NumField() != 1 {
		return fmt.Errorf("reflect.rtype layout changed: expected 1 field, got %d", target.NumField())
	}
	ptr := reflect.PointerTo(target.Field(0).Type)
	method, ok := ptr.MethodByName("StructType")
	if !ok {
		return fmt.Errorf("%s has no StructType method", ptr)
	}
	if method.Type.NumOut() != 1 {
		return fmt.Errorf("%s.StructType: expected 1 return value, got %d", ptr, method.Type.NumOut())
	}
	if method.Type.Out(0).Kind() != reflect.Pointer {
		return fmt.Errorf("%s.StructType: expected pointer return type, got %s", ptr, method.Type.Out(0))
	}

	abi = method.Type.Out(0).Elem()
	return nil
}

// validateABI checks that the discovered StructType has the layout abiPatch
// relies on: a Fields slice of structs, each with a pointer field Typ.
func validateABI() error {
	if abi == nil {
		return fmt.Errorf("ABI type not initialized")
	}

	fields, ok := abi.FieldByName("Fields")
	if !ok {
		return fmt.Errorf("StructType.Fields not found")
	}

	if fields.Type.Kind() != reflect.Slice {
		return fmt.Errorf("StructType.Fields is not a slice")
	}

	field := fields.Type.Elem()

	if field.Kind() != reflect.Struct {
		return fmt.Errorf("StructType.Fields element is not a struct")
	}

	typ, ok := field.FieldByName("Typ")
	if !ok {
		return fmt.Errorf("StructField.Typ not found")
	}

	if typ.Type.Kind() != reflect.Pointer {
		return fmt.Errorf("StructField.Typ is not a pointer")
	}

	return nil
}

// abiPatch replaces the type of field fieldIndex in *structType in place.
// The new type is built by wrapping *concreteType in the same pointer,
// slice and array layers as fieldType.
//
// The struct's size, field offsets and GC metadata are not updated, so the
// new field type must have the same memory layout as the one it replaces.
func abiPatch(structType *reflect.Type, fieldIndex int, fieldType reflect.Type, concreteType *reflect.Type) error {
	if abiErr != nil {
		return fmt.Errorf("ABI unavailable: %w", abiErr)
	}
	if structType == nil || *structType == nil {
		return fmt.Errorf("ABI patch: struct type is nil")
	}
	if concreteType == nil || *concreteType == nil {
		return fmt.Errorf("ABI patch: target type of %s is nil", *structType)
	}
	if (*structType).Kind() != reflect.Struct {
		return fmt.Errorf("ABI patch: %s is not a struct", *structType)
	}

	// View the struct's runtime type as internal/abi.StructType.
	typ := reflect.NewAt(abi, reflect.ValueOf(*structType).UnsafePointer()).Elem()

	fields := typ.FieldByName("Fields")

	if fieldIndex < 0 || fieldIndex >= fields.Len() {
		return fmt.Errorf("ABI patch: field index %d out of range [0, %d) for %s", fieldIndex, fields.Len(), *structType)
	}

	f := fields.Index(fieldIndex).FieldByName("Typ")

	if !f.IsValid() || !f.CanAddr() {
		return fmt.Errorf("ABI field Typ is inaccessible")
	}

	addr := (*unsafe.Pointer)(f.Addr().UnsafePointer())

	// Rebuild the field type around the rewritten concrete type.
	ptr := typeCascade(fieldType, *concreteType)

	*addr = reflect.ValueOf(ptr).UnsafePointer()

	runtime.KeepAlive(ptr)
	runtime.KeepAlive(*structType)

	return nil
}
