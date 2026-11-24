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
	Metadata interface {
		dummyMethod(Metadata)
	}

	CircularReferenceHandlingPolicy int
)

const (
	CirRefHdlrPolicyBlock  CircularReferenceHandlingPolicy = 0
	CirRefHdlrPolicyIgnore CircularReferenceHandlingPolicy = 1
	CirRefHdlrPolicyHandle CircularReferenceHandlingPolicy = 3
)

var (
	metadataType = reflect.TypeFor[Metadata]()
	typeCache    = make(map[reflect.Type]reflect.Type)
	emptyType    = reflect.PointerTo(reflect.TypeFor[struct{}]())
	mut          sync.Mutex
)

func Rewrite(src reflect.Type, crhp CircularReferenceHandlingPolicy) (reflect.Type, error) {
	mut.Lock()
	defer mut.Unlock()
	if value, ok := typeCache[src]; ok {
		return value, nil
	}
	typ, crh, err := typeRewrite(src, make(circularReferenceDetector), crhp)
	if err != nil {
		return nil, err
	}
	if err := crh.inject(); err != nil {
		return nil, err
	}
	typeCache[src] = typ
	return typ, nil
}

func typeRewrite(src reflect.Type, cld circularReferenceDetector, crhp CircularReferenceHandlingPolicy) (reflect.Type, circularReferenceHandlers, error) {
	if value, ok := typeCache[src]; ok {
		return value, nil, nil
	}

	var localCrh circularReferenceHandlers

	indirectType, n := indirect(src)
	var out reflect.Type

	switch indirectType.Kind() {
	case reflect.Slice:
		{
			res, crh, err := typeRewrite(indirectType.Elem(), cld, crhp)
			if err != nil {
				return nil, nil, err
			}
			out = reflect.SliceOf(res)
			localCrh = append(localCrh, crh...)
		}
	case reflect.Array:
		{
			res, crh, err := typeRewrite(indirectType.Elem(), cld, crhp)
			if err != nil {
				return nil, nil, err
			}
			out = reflect.ArrayOf(indirectType.Len(), res)
			localCrh = append(localCrh, crh...)
		}
	case reflect.Struct:
		{
			res, crh, err := structRewrite(indirectType, cld, crhp)
			if err != nil {
				return nil, crh, err
			}
			out = res
			localCrh = append(localCrh, crh...)
		}
	default:
		{
			out = indirectType
		}
	}

	if n == 0 {
		typeCache[src] = out
		return out, localCrh, nil
	}
	ptr := pointerTo(out, n)
	typeCache[src] = ptr
	return ptr, localCrh, nil
}

func structRewrite(src reflect.Type, cld circularReferenceDetector, crhp CircularReferenceHandlingPolicy) (reflect.Type, circularReferenceHandlers, error) {
	if value, ok := typeCache[src]; ok {
		return value, nil, nil
	}

	var finalResult reflect.Type
	var crh circularReferenceHandlers
	out := make([]reflect.StructField, 0)
	cld.track(src, &finalResult)
	defer cld.release(src)

	for i := range src.NumField() {
		field := src.Field(i)
		typ := field.Type

		ityp := concreteType(typ)
		if val, ok := cld.has(ityp); ok {
			if crhp == CirRefHdlrPolicyBlock {
				return nil, nil, fmt.Errorf("circular references are not allowed")
			}
			if crhp == CirRefHdlrPolicyHandle {
				crh = append(crh, func() error {
					return abiPatch(&finalResult, i, typ, val)
				})
				out = append(out, reflect.StructField{
					Name:      field.Name,
					Anonymous: false,
					Index:     field.Index,
					Offset:    field.Offset,
					PkgPath:   field.PkgPath,
					Type:      emptyType,
					Tag:       field.Tag,
				})
				continue
			}

			out = append(out, reflect.StructField{
				Name:      field.Name,
				Anonymous: false,
				Index:     field.Index,
				Offset:    field.Offset,
				PkgPath:   field.PkgPath,
				Type:      reflect.TypeFor[unsafe.Pointer](),
				Tag:       reflect.StructTag(`bson:"-"`),
			})
			continue
		}

		if val, ok := field.Tag.Lookup("bson"); !ok || !isOpaque(val) {
			res, fn, err := typeRewrite(field.Type, cld, crhp)
			if err != nil {
				return nil, nil, err
			}
			typ = res
			crh = append(crh, fn...)
		}

		if !field.Anonymous {
			tag := field.Tag
			if tagValue, ok := isMetadata(field.Type); ok {
				tag = tagValue
			}
			out = append(out, reflect.StructField{
				Name:      field.Name,
				Anonymous: false,
				Index:     field.Index,
				Offset:    field.Offset,
				PkgPath:   field.PkgPath,
				Type:      typ,
				Tag:       tag,
			})
			continue
		}
		if field.Type.Implements(metadataType) {
			out = append(out, reflect.StructField{
				Name:      field.Name,
				Anonymous: false,
				Index:     field.Index,
				Offset:    field.Offset,
				PkgPath:   field.PkgPath,
				Type:      typ,
				Tag:       reflect.StructTag(`bson:"-"`),
			})
			continue
		}
		out = append(out, reflect.StructField{
			Name:      field.Name,
			Anonymous: true,
			Index:     field.Index,
			Offset:    field.Offset,
			PkgPath:   field.PkgPath,
			Type:      typ,
			Tag:       reflect.StructTag(`bson:",inline"`),
		})

	}
	finalResult = reflect.StructOf(out)
	typeCache[src] = finalResult
	return finalResult, crh, nil
}

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

func isOpaque(str string) bool {
	return slices.Contains(strings.Split(str, ","), "opaque")
}
