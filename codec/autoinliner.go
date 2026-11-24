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

// Package codec provides BSON codecs that automatically inline anonymous
// (embedded) struct fields, without requiring `bson:",inline"` tags.
package codec

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"

	"github.com/vedadiyan/mango/internal/hack"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	// AutoInlinerCodec encodes and decodes values of type T, inlining
	// anonymous struct fields. It works by viewing T as an equivalent type
	// with the same memory layout and inline tags applied.
	AutoInlinerCodec[T any] struct {
		srcType     reflect.Type
		dynamicType reflect.Type
		registry    *bson.Registry
	}

	autoInlinerCodecOptions struct {
		crhp     hack.LoopHandlerPolicy
		registry *bson.Registry
	}

	// AutoInlinerCodecOption configures an AutoInlinerCodec.
	AutoInlinerCodecOption func(*autoInlinerCodecOptions)
)

var circRefHandlerOnce sync.Once

// WithCircRefBlock makes codec creation fail if T contains a circular type
// reference. This is the default.
func WithCircRefBlock() AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.crhp = hack.LoopHandlerPolicyBlock
	}
}

// WithCircRefIgnore excludes circular fields from encoding and decoding.
func WithCircRefIgnore() AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.crhp = hack.LoopHandlerPolicyIgnore
	}
}

// WithCircRefHandle supports circular fields by patching runtime type
// metadata. It enables ABI patching for the whole process the first time it
// is used.
//
// This relies on Go runtime internals and may break with new Go releases.
func WithCircRefHandle() AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.crhp = hack.LoopHandlerPolicyHandle
		circRefHandlerOnce.Do(func() {
			hack.EnableAbiPatch(true)
		})
	}
}

// WithRegistry sets the registry used by Encode, Decode, EncodeValue and
// DecodeValue, so that custom codecs apply to nested values. If unset, the
// driver's default registry is used. RegisterCodec sets this automatically.
func WithRegistry(registry *bson.Registry) AutoInlinerCodecOption {
	return func(aico *autoInlinerCodecOptions) {
		aico.registry = registry
	}
}

// RegisterCodec creates an AutoInlinerCodec for T and registers it as the
// encoder and decoder for T in registry.
//
// Values are encoded and decoded in a single pass using the registry and
// options of the calling encoder or decoder, so other codecs registered in
// registry apply to nested values.
func RegisterCodec[T any](registry *bson.Registry, opts ...AutoInlinerCodecOption) error {
	opts = append(opts, WithRegistry(registry))
	alc, err := NewAutoInlinerCodec[T](opts...)
	if err != nil {
		return err
	}

	// T needed no rewriting, so the default codec already handles it.
	// Registering would make the encoder look itself up and recurse forever.
	if alc.dynamicType == alc.srcType {
		return nil
	}

	// Encode T by viewing it as the rewritten type and delegating to the
	// rewritten type's encoder in the same registry.
	registry.RegisterTypeEncoder(alc.srcType, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			if v.Type() != alc.srcType {
				return fmt.Errorf("codec: cannot encode %s as %s", v.Type(), alc.srcType)
			}
			enc, err := ec.LookupEncoder(alc.dynamicType)
			if err != nil {
				return err
			}
			return enc.EncodeValue(ec, vw, alc.asDynamic(v))
		},
	))

	// Decode T in place by viewing it as the rewritten type and delegating
	// to the rewritten type's decoder in the same registry.
	registry.RegisterTypeDecoder(alc.srcType, bson.ValueDecoderFunc(
		func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
			if v.Type() != alc.srcType || !v.CanSet() {
				return fmt.Errorf("codec: cannot decode into %s: expected settable %s", v.Type(), alc.srcType)
			}
			dec, err := dc.LookupDecoder(alc.dynamicType)
			if err != nil {
				return err
			}
			return dec.DecodeValue(dc, vr, alc.asDynamic(v))
		},
	))

	return nil
}

// NewAutoInlinerCodec returns a codec for T configured by opts. It returns
// an error if T cannot be rewritten, for example because it contains a
// circular reference under the default policy.
func NewAutoInlinerCodec[T any](opts ...AutoInlinerCodecOption) (*AutoInlinerCodec[T], error) {
	options := &autoInlinerCodecOptions{}
	for _, opt := range opts {
		opt(options)
	}

	srcType := reflect.TypeFor[T]()
	dynamicType, err := hack.Rewrite(srcType, options.crhp)
	if err != nil {
		return nil, err
	}
	return &AutoInlinerCodec[T]{
		srcType:     srcType,
		dynamicType: dynamicType,
		registry:    options.registry,
	}, nil
}

// Encode marshals value to BSON with anonymous fields inlined.
func (x *AutoInlinerCodec[T]) Encode(value *T) ([]byte, error) {
	out := convert(value, x.dynamicType)
	return x.marshal(out.Interface())
}

// EncodeValue is like Encode but takes a reflect.Value. v must be a pointer
// to a value of type T.
func (x *AutoInlinerCodec[T]) EncodeValue(v reflect.Value) ([]byte, error) {
	out := convertValue(v, x.dynamicType)
	return x.marshal(out.Interface())
}

// Decode unmarshals BSON into a new T, treating anonymous fields as inlined.
func (x *AutoInlinerCodec[T]) Decode(value []byte) (*T, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := x.unmarshal(value, v); err != nil {
		return nil, err
	}
	out := convert(v, x.srcType).Interface()
	return out.(*T), nil
}

// DecodeValue is like Decode but returns the result as a reflect.Value
// holding a *T.
func (x *AutoInlinerCodec[T]) DecodeValue(value []byte) (reflect.Value, error) {
	v := reflect.New(x.dynamicType).Interface()
	if err := x.unmarshal(value, v); err != nil {
		return reflect.Value{}, err
	}
	out := convert(v, x.srcType)
	return out, nil
}

// asDynamic returns v, a value of type T, viewed as the rewritten type. The
// result shares v's memory, so writes through it update v. If v is not
// addressable, it is copied first.
func (x *AutoInlinerCodec[T]) asDynamic(v reflect.Value) reflect.Value {
	if !v.CanAddr() {
		p := reflect.New(x.srcType)
		p.Elem().Set(v)
		v = p.Elem()
	}
	return reflect.NewAt(x.dynamicType, v.Addr().UnsafePointer()).Elem()
}

// marshal encodes v using the codec's registry, or the default registry if
// none is set.
func (x *AutoInlinerCodec[T]) marshal(v any) ([]byte, error) {
	if x.registry == nil {
		return bson.Marshal(v)
	}
	buf := new(bytes.Buffer)
	enc := bson.NewEncoder(bson.NewDocumentWriter(buf))
	enc.SetRegistry(x.registry)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// unmarshal decodes data into v using the codec's registry, or the default
// registry if none is set.
func (x *AutoInlinerCodec[T]) unmarshal(data []byte, v any) error {
	if x.registry == nil {
		return bson.Unmarshal(data, v)
	}
	dec := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(data)))
	dec.SetRegistry(x.registry)
	return dec.Decode(v)
}

// convert reinterprets the pointer v as a pointer to typ. The pointed-to
// types must have identical memory layouts.
func convert(v any, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, reflect.ValueOf(v).UnsafePointer())
}

// convertValue is like convert but takes a reflect.Value.
func convertValue(v reflect.Value, typ reflect.Type) reflect.Value {
	return reflect.NewAt(typ, v.UnsafePointer())
}
