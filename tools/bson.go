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

// Package tools provides helpers used by generated code.
package tools

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	// Converter copies the value src into target. It is called for each
	// value that is not a pointer, slice, array or interface.
	Converter func(src reflect.Value, target reflect.Value) error

	options struct {
		converter Converter
	}

	// Option configures Copy.
	Option func(*options)
)

// WithConverter sets the function Copy uses to copy individual values. By
// default, values are assigned directly.
func WithConverter(fn Converter) Option {
	return func(o *options) {
		o.converter = fn
	}
}

// Copy copies src into the value target points to. Pointers are followed
// and allocated as needed, and slices and arrays are copied element by
// element. All other values are copied by the converter, if one is set, or
// assigned directly.
//
// target must be a non-nil pointer.
func Copy(src any, target any, opts ...Option) error {
	var fn Converter
	if opts != nil {
		options := &options{}
		for _, opt := range opts {
			opt(options)
		}
		fn = options.converter
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer {
		return fmt.Errorf("expected pointer, got %s", value.Type())
	}
	return cp(reflect.ValueOf(src), value.Elem(), fn)
}

// cp copies sv into tv. See Copy.
func cp(sv reflect.Value, tv reflect.Value, fn Converter) error {
	// Allocate target pointers and dereference source pointers.
	for tv.Kind() == reflect.Pointer {
		tv.Set(reflect.New(tv.Type().Elem()))
		tv = tv.Elem()
	}

	for sv.Kind() == reflect.Pointer {
		sv = sv.Elem()
	}

	switch sv.Kind() {
	case reflect.Slice, reflect.Array:
		{
			switch tv.Kind() {
			// Copy into an interface as a slice of the interface type.
			case reflect.Interface:
				{
					c := reflect.MakeSlice(reflect.SliceOf(tv.Type()), sv.Len(), sv.Cap())
					for i := range sv.Len() {
						if err := cp(sv.Index(i), c.Index(i), fn); err != nil {
							return err
						}
					}
					tv.Set(c)
					return nil
				}
			// Copy element by element, preserving length and capacity.
			case reflect.Slice, reflect.Array:
				{
					tv.SetZero()
					tv.Grow(sv.Cap())
					tv.SetLen(sv.Len())
					tv.SetCap(sv.Cap())
					for i := range sv.Len() {
						if err := cp(sv.Index(i), tv.Index(i), fn); err != nil {
							return err
						}
					}
					return nil
				}
			default:
				{
					return fmt.Errorf("expected either slice or array, got %s", tv.Type())
				}
			}
		}
	// Copy the value held by the interface.
	case reflect.Interface:
		{
			return cp(sv.Elem(), tv, fn)
		}
	default:
		{
			if fn != nil {
				return fn(sv, tv)
			}
			tv.Set(sv)
			return nil
		}
	}
}

// FieldName returns the BSON key of the single field of struct type T: the
// name in its bson tag, or the Go field name if the tag sets none. It
// returns "-" if T is not a struct with exactly one field.
func FieldName[T any]() string {
	v := reflect.TypeFor[T]()

	if v.Kind() != reflect.Struct || v.NumField() != 1 {
		return "-"
	}

	field := v.Field(0)

	if value, ok := field.Tag.Lookup("bson"); ok {
		if name, _, _ := strings.Cut(value, ","); name != "" {
			return name
		}
	}

	return field.Name
}
