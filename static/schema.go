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

// Package static defines the types used to declare MongoDB schemas in Go.
//
// Schemas are declared as composite literals returned from exported
// functions. They are never evaluated at run time: the code generator reads
// them from source and produces Go types and a $jsonSchema validator.
//
//	func DefineUser() static.Schema {
//		return static.Schema{
//			Properties: static.Properties{
//				"name": static.Scalar{Type: []static.BasicType{static.TypeString}, Required: true},
//			},
//		}
//	}
//
// Because values are read statically, fields must be set using literals or
// constants.
package static

type (
	// Type is implemented by every schema node that can appear as a
	// property, array item or combinator spec. It cannot be implemented
	// outside this package.
	Type interface {
		typ()
	}

	// Index describes the index options for a field.
	Index struct {
		Order      int
		Unique     bool
		Sparse     bool
		Background bool
	}

	// Scalar describes a field holding a single value of one or more basic
	// types.
	Scalar struct {
		Title        string
		Description  string
		Type         []BasicType
		Required     bool
		Min          int
		Max          int
		ExclusiveMin bool
		ExclusiveMax bool
		MultipleOf   float64
		MinLen       uint
		MaxLen       int
		Pattern      string
		Index        Index

		// ObjectName names the field when the scalar is a combinator spec.
		ObjectName string
	}

	// Enum describes a field restricted to a fixed set of values.
	Enum[T Int | Float | String] struct {
		Title       string
		Description string
		Required    bool
		Enum        []T
	}

	// Composite describes a nested object.
	Composite struct {
		Title       string
		Description string

		// Required is either a bool, marking the object itself as required,
		// or a []string listing its required properties.
		Required any

		Properties           Properties
		Dependencies         Dependencies
		MaxProperties        int
		MinProperties        int
		AdditionalProperties bool
		Index                Index

		OneOf Combinator[OneOf]
		AnyOf Combinator[AnyOf]
		AllOf Combinator[AllOf]
		Not   Combinator[Not]

		// ObjectName names the object when it is a combinator spec.
		ObjectName string
	}

	// Combinator combines the schemas in Specs. T selects the kind of
	// combination: OneOf, AnyOf, AllOf or Not.
	Combinator[T OneOf | AnyOf | AllOf | Not] struct {
		Title       string
		Description string
		Specs       T
		Index       Index

		// ObjectName names the combinator's generated type.
		ObjectName string
	}

	// Array describes a field holding a list of values matching Items.
	Array[T Type] struct {
		Required        bool
		Items           T
		MinItems        int
		MaxItems        int
		UniqueItems     bool
		AdditionalItems bool
		Index           Index

		// AllowNull permits null elements.
		AllowNull bool

		// ObjectName names the array when it is a combinator spec.
		ObjectName string
	}

	// NotCondition is a condition negated by a Not combinator.
	NotCondition struct {
		// Required lists properties that must not all be present.
		Required []string
	}

	// OneOf, AnyOf and AllOf list the schemas combined by a Combinator.
	OneOf Items
	AnyOf Items
	AllOf Items

	// Not holds the single condition negated by a Combinator.
	Not [1]NotCondition

	// Schema is the root of a schema definition.
	Schema struct {
		Title        string
		Description  string
		Properties   Properties
		Dependencies Dependencies

		// Conditions are MongoDB query conditions that documents must also
		// satisfy.
		Conditions Conditions

		OneOf Combinator[OneOf]
		AnyOf Combinator[AnyOf]
		AllOf Combinator[AllOf]
		Not   Combinator[Not]
	}

	// BasicType is a MongoDB BSON type alias, as used by $jsonSchema's
	// bsonType keyword.
	BasicType string

	// BasicTypes is a list of BSON types.
	BasicTypes []BasicType

	// Items is a list of schemas.
	Items []Type

	// Properties maps property names to their schemas.
	Properties map[string]Type

	// Dependencies maps a property to the properties that must be present
	// whenever it is.
	Dependencies map[string][]string

	// Condition is a MongoDB query expression.
	Condition map[string]any

	// Conditions is a list of query expressions, all of which must hold.
	Conditions []Condition

	// Int, Float and String are the value types allowed in an Enum.
	Int    int64
	Float  float64
	String string
)

// BSON types supported in schemas.
const (
	TypeDouble     BasicType = "double"
	TypeString     BasicType = "string"
	TypeBinary     BasicType = "binData"
	TypeObjectId   BasicType = "objectId"
	TypeBool       BasicType = "bool"
	TypeNil        BasicType = "null"
	TypeRegex      BasicType = "regex"
	TypeJavaScript BasicType = "javascript"
	TypeInt        BasicType = "int"
	TypeTimeStamp  BasicType = "timestamp"
	TypeLong       BasicType = "long"
	TypeDecimal    BasicType = "decimal"
	TypeObject     BasicType = "object"
)

// justPanic panics. The typ methods only mark their types as implementing
// Type and are never meant to be called.
func justPanic() {
	panic("this method should never be called")
}

func (Scalar) typ() {
	justPanic()
}

func (Enum[T]) typ() {
	justPanic()
}

func (Composite) typ() {
	justPanic()
}

func (Combinator[T]) typ() {
	justPanic()
}

func (Array[T]) typ() {
	justPanic()
}
