package hack

import (
	"reflect"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	AutoInlinerCodec[T any] struct {
		srcType     reflect.Type
		dynamicType reflect.Type
	}

	Metadata interface {
		dummyMethod(Metadata)
	}

	CycleLoopDetector map[reflect.Type]bool
)

var (
	bsonRaw      = reflect.TypeFor[bson.Raw]()
	metadataType = reflect.TypeFor[Metadata]()
	typeCache    = make(map[reflect.Type]reflect.Type)
	abi          reflect.Type
	mut          sync.Mutex
)

//go:linkname typelinks reflect.typelinks
func typelinks() (sections []unsafe.Pointer, offset [][]int32)

//go:linkname resolveTypeOff reflect.resolveTypeOff
func resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer

//go:linkname toType reflect.toType
func toType(t unsafe.Pointer) reflect.Type

func init() {
	sections, offsets := typelinks()
	for i, base := range sections {
		for _, off := range offsets[i] {
			if t := toType(unsafe.Add(base, uintptr(off))).Elem(); t.String() == "abi.StructType" {
				abi = t
				return
			}
		}
	}
	panic("abi.StructType not found in runtime typelinks")
}

func (cld CycleLoopDetector) Track(typ reflect.Type) {
	cld[typ] = true
}

func (cld CycleLoopDetector) Has(typ reflect.Type) bool {
	_, ok := cld[typ]
	return ok
}

func (cld CycleLoopDetector) Release(typ reflect.Type) {
	delete(cld, typ)
}

func RegisterCodec[T any](registry *bson.Registry) error {
	alc, err := NewAutoInlinerCodec[T]()
	if err != nil {
		return err
	}

	encoder, err := registry.LookupEncoder(bsonRaw)
	if err != nil {
		return err
	}

	decoder, err := registry.LookupDecoder(bsonRaw)
	if err != nil {
		return err
	}

	registry.RegisterTypeEncoder(alc.srcType, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			val := reflect.New(alc.srcType)
			val.Elem().Set(v)
			out, err := alc.EncodeValue(val)
			if err != nil {
				return err
			}
			return encoder.EncodeValue(ec, vw, reflect.ValueOf(bson.Raw(out)))
		},
	))

	registry.RegisterTypeDecoder(alc.srcType, bson.ValueDecoderFunc(
		func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
			raw := bson.Raw{}
			if err := decoder.DecodeValue(dc, vr, reflect.ValueOf(&raw).Elem()); err != nil {
				return err
			}
			out, err := alc.DecodeValue(raw)
			if err != nil {
				return err
			}
			v.Set(out.Elem())
			return nil
		},
	))

	return nil
}

func NewAutoInlinerCodec[T any]() (*AutoInlinerCodec[T], error) {
	srcType := reflect.TypeFor[T]()
	dynamicType, err := AutoInliner(srcType)
	if err != nil {
		return nil, err
	}
	return &AutoInlinerCodec[T]{
		srcType:     srcType,
		dynamicType: dynamicType,
	}, nil
}

func (x *AutoInlinerCodec[T]) Encode(value *T) ([]byte, error) {
	out := convert(value, x.dynamicType)
	return bson.Marshal(out.Interface())
}

func (x *AutoInlinerCodec[T]) EncodeValue(v reflect.Value) ([]byte, error) {
	out := convertValue(v, x.dynamicType)
	return bson.Marshal(out.Interface())
}

func (x *AutoInlinerCodec[T]) Decode(value []byte) (*T, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.Unmarshal(value, v); err != nil {
		return nil, err
	}
	out := convert(v, x.srcType).Interface()
	return out.(*T), nil
}

func (x *AutoInlinerCodec[T]) DecodeValue(value []byte) (reflect.Value, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := bson.Unmarshal(value, v); err != nil {
		return reflect.Value{}, err
	}
	out := convert(v, x.srcType)
	return out, nil
}

func AutoInliner(src reflect.Type) (reflect.Type, error) {
	mut.Lock()
	defer mut.Unlock()
	if value, ok := typeCache[src]; ok {
		return value, nil
	}
	typ, err := autoInliner(src, make(CycleLoopDetector))
	if err != nil {
		return nil, err
	}
	typeCache[src] = typ
	return typ, nil
}

func autoInliner(src reflect.Type, cld CycleLoopDetector) (reflect.Type, error) {
	if value, ok := typeCache[src]; ok {
		return value, nil
	}

	indirectType, n := indirect(src)
	var out reflect.Type

	switch indirectType.Kind() {
	case reflect.Slice:
		{
			res, err := autoInliner(indirectType.Elem(), cld)
			if err != nil {
				return nil, err
			}
			out = reflect.SliceOf(res)
		}
	case reflect.Array:
		{
			res, err := autoInliner(indirectType.Elem(), cld)
			if err != nil {
				return nil, err
			}
			out = reflect.ArrayOf(indirectType.Len(), res)
		}
	case reflect.Struct:
		{
			res, err := rewriteStruct(indirectType, cld)
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

func rewriteStruct(src reflect.Type, cld CycleLoopDetector) (reflect.Type, error) {
	if value, ok := typeCache[src]; ok {
		return value, nil
	}

	var finalResult reflect.Type
	out := make([]reflect.StructField, 0)
	cld.Track(src)
	defer cld.Release(src)

	finalFields := make([]func(), 0)

	idx := 0
	for field := range src.Fields() {
		idx++
		typ := field.Type

		ityp, _ := indirect(typ)
		if ok := cld.Has(ityp); ok {
			placeholderPtr := reflect.PointerTo(reflect.TypeFor[struct{}]())
			i := idx - 1
			finalFields = append(finalFields, func() {
				typ := reflect.NewAt(abi, reflect.ValueOf(finalResult).UnsafePointer()).Elem()
				f := typ.FieldByName("Fields").Index(i).FieldByName("Typ")
				addr := (*unsafe.Pointer)(f.Addr().UnsafePointer())
				*addr = reflect.ValueOf(reflect.PointerTo(finalResult)).UnsafePointer()
			})

			out = append(out, reflect.StructField{
				Name:      field.Name,
				Anonymous: false,
				Index:     field.Index,
				Offset:    field.Offset,
				PkgPath:   field.PkgPath,
				Type:      placeholderPtr,
				Tag:       "",
			})
			continue
		}

		if val, ok := field.Tag.Lookup("bson"); !ok || !isOpaque(val) {
			res, err := autoInliner(field.Type, cld)
			if err != nil {
				return nil, err
			}
			typ = res
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
	for _, fn := range finalFields {
		fn()
	}
	return finalResult, nil
}

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

func convert(v any, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, reflect.ValueOf(v).UnsafePointer())
}

func convertValue(v reflect.Value, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, v.UnsafePointer())
}
