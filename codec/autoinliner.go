package codec

import (
	"reflect"
	"sync"

	"github.com/vedadiyan/mango/internal/hack"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	AutoInlinerCodec[T any] struct {
		srcType     reflect.Type
		dynamicType reflect.Type
	}

	autoInlinerCodecOptions struct {
		crhp hack.CircularReferenceHandlingPolicy
	}

	AutoInlinerCodecOption func(*autoInlinerCodecOptions)
)

var (
	bsonRaw            = reflect.TypeFor[bson.Raw]()
	circRefHandlerOnce sync.Once
)

func WithCircRefBlock() AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.crhp = hack.CirRefHdlrPolicyBlock
	}
}

func WithCircRefIgnore() AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.crhp = hack.CirRefHdlrPolicyIgnore
	}
}

func WithCircRefHandle() AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.crhp = hack.CirRefHdlrPolicyHandle
		circRefHandlerOnce.Do(func() {
			hack.EnableAbiPatch(true)
		})
	}
}

func RegisterCodec[T any](registry *bson.Registry, opts ...AutoInlinerCodecOption) error {
	alc, err := NewAutoInlinerCodec[T](opts...)
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

	registry.RegisterTypeDecoder(reflect.PointerTo(reflect.TypeFor[struct{}]()), bson.ValueDecoderFunc(func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
		return vr.Skip()

	}))

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

func NewAutoInlinerCodec[T any](opts ...AutoInlinerCodecOption) (*AutoInlinerCodec[T], error) {
	autoInlinerCodecOptions := &autoInlinerCodecOptions{}
	for _, opt := range opts {
		opt(autoInlinerCodecOptions)
	}

	srcType := reflect.TypeFor[T]()
	dynamicType, err := hack.Rewrite(srcType, autoInlinerCodecOptions.crhp)
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

func convert(v any, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, reflect.ValueOf(v).UnsafePointer())
}

func convertValue(v reflect.Value, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, v.UnsafePointer())
}
