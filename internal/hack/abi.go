package hack

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

const (
	abiPackage = "internal/abi"
	abiType    = "StructType"
)

var (
	abi    reflect.Type
	abiErr error
)

//go:linkname typelinks reflect.typelinks
func typelinks() (sections []unsafe.Pointer, offset [][]int32)

//go:linkname resolveTypeOff reflect.resolveTypeOff
func resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer

//go:linkname toType reflect.toType
func toType(t unsafe.Pointer) reflect.Type

func init() {
	abiErr = discoverABI()
	if abiErr == nil {
		abiErr = validateABI()
	}
}

func discoverABI() error {
	sections, offsets := typelinks()

	for i, base := range sections {
		for _, off := range offsets[i] {
			t := toType(unsafe.Add(base, uintptr(off)))
			if t == nil || t.Kind() != reflect.Pointer {
				continue
			}

			t = t.Elem()

			if t.Kind() != reflect.Struct || t.Name() != abiType || t.PkgPath() != abiPackage {
				continue
			}

			abi = t
			return nil
		}
	}

	return fmt.Errorf("abi.StructType not found in runtime typelinks")
}

func validateABI() error {
	if abi == nil {
		return fmt.Errorf("ABI type not initialized")
	}

	if abi.Kind() != reflect.Struct || abi.Name() != abiType || abi.PkgPath() != abiPackage {
		return fmt.Errorf("unexpected ABI StructType: %v", abi)
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

	if typ.Type.Elem().Name() != "Type" || typ.Type.Elem().PkgPath() != abiPackage {
		return fmt.Errorf("unexpected StructField.Typ type")
	}

	return nil
}

func abiPatch(structType *reflect.Type, fieldIndex int, fieldType reflect.Type, concreteType *reflect.Type) error {
	if abiErr != nil {
		return fmt.Errorf("ABI unavailable: %w", abiErr)
	}

	typ := reflect.NewAt(abi, reflect.ValueOf(*structType).UnsafePointer()).Elem()

	fields := typ.FieldByName("Fields")

	if fieldIndex >= fields.Len() {
		return fmt.Errorf("ABI field index out of bounds")
	}

	f := fields.Index(fieldIndex).FieldByName("Typ")

	if !f.IsValid() || !f.CanAddr() {
		return fmt.Errorf("ABI field Typ is inaccessible")
	}

	addr := (*unsafe.Pointer)(f.Addr().UnsafePointer())

	ptr := *concreteType

	for _, i := range bluePrint(fieldType) {
		switch i.Kind() {
		case reflect.Pointer:
			{
				ptr = reflect.PointerTo(ptr)
			}
		case reflect.Slice:
			{
				ptr = reflect.SliceOf(ptr)
			}
		case reflect.Array:
			{
				ptr = reflect.ArrayOf(i.Len(), ptr)
			}
		}
	}

	*addr = reflect.ValueOf(ptr).UnsafePointer()

	runtime.KeepAlive(ptr)
	runtime.KeepAlive(*structType)

	return nil
}
