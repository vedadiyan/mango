package hack

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

var (
	abi    reflect.Type
	abiErr error
)

func init() {
	EnableAbiPatch(false)
}

func EnableAbiPatch(b bool) error {
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

func discoverABI() error {
	target := reflect.TypeOf(reflect.TypeFor[struct{}]()).Elem()
	if target.NumField() != 1 {
		return fmt.Errorf("")
	}
	ptr := reflect.PointerTo(target.Field(0).Type)
	method, ok := ptr.MethodByName("StructType")
	if !ok {
		return fmt.Errorf("")
	}
	if method.Type.NumOut() != 1 {
		return fmt.Errorf("")
	}

	abi = method.Type.Out(0).Elem()
	return nil
}

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
