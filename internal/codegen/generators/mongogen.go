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

package generators

import (
	"encoding/json"
	"fmt"

	"github.com/vedadiyan/mango/internal/codegen"
)

type (
	// Index describes the index options for a field.
	Index struct {
		Order      *int
		Unique     *bool
		Sparse     *bool
		Background *bool
	}

	// BsonSchemaModel is a MongoDB $jsonSchema document. Unset fields are
	// omitted from the JSON output.
	BsonSchemaModel struct {
		Title        *string                     `json:"title,omitempty"`
		Description  *string                     `json:"description,omitempty"`
		Type         []string                    `json:"bsonType,omitempty"`
		Properties   map[string]*BsonSchemaModel `json:"properties,omitempty"`
		Items        any                         `json:"items,omitempty"` // *BsonSchemaModel or []*BsonSchemaModel
		Required     []string                    `json:"required,omitempty"`
		Dependencies map[string][]string         `json:"dependencies,omitempty"`
		Enum         []any                       `json:"enum,omitempty"`

		MinLen  *int64  `json:"minLength,omitempty"`
		MaxLen  *int64  `json:"maxLength,omitempty"`
		Pattern *string `json:"pattern,omitempty"`

		Min          *int64   `json:"minimum,omitempty"`
		Max          *int64   `json:"maximum,omitempty"`
		ExclusiveMin *bool    `json:"exclusiveMinimum,omitempty"`
		ExclusiveMax *bool    `json:"exclusiveMaximum,omitempty"`
		MultipleOf   *float64 `json:"multipleOf,omitempty"`

		MinItems    *int64 `json:"minItems,omitempty"`
		MaxItems    *int64 `json:"maxItems,omitempty"`
		UniqueItems *bool  `json:"uniqueItems,omitempty"`

		MaxProperties *int64 `json:"maxProperties,omitempty"`
		MinProperties *int64 `json:"minProperties,omitempty"`

		AdditionalItems      *bool `json:"additionalItems,omitempty"`
		AdditionalProperties *bool `json:"additionalProperties,omitempty"`

		AnyOf []*BsonSchemaModel `json:"anyOf,omitempty"`
		OneOf []*BsonSchemaModel `json:"oneOf,omitempty"`
		AllOf []*BsonSchemaModel `json:"allOf,omitempty"`
		Not   *BsonSchemaModel   `json:"not,omitempty"`

		index Index // index options; not part of the schema

		// Conditions are query conditions that must hold in addition to the
		// schema. They are not part of the JSON Schema itself; see
		// ToMongoValidationSchema.
		Conditions []map[string]any `json:"-"`
	}

	// BsonSchemaGenerationOptions configures ToBsonSchema.
	BsonSchemaGenerationOptions struct {
		nullableItems bool
	}

	// BsonSchemaGenerationOption configures ToBsonSchema.
	BsonSchemaGenerationOption func(*BsonSchemaGenerationOptions)
)

// WithNullableItems makes the generated schema also accept null if b is true.
func WithNullableItems(b bool) BsonSchemaGenerationOption {
	return func(bsgo *BsonSchemaGenerationOptions) {
		bsgo.nullableItems = b
	}
}

// ToMongoValidationSchema returns in as indented JSON for use as a MongoDB
// collection validator. If in has Conditions, the schema and the conditions
// are combined with $and.
func ToMongoValidationSchema(in *BsonSchemaModel) (string, error) {
	val := any(in)

	if in.Conditions != nil {
		val = map[string][]any{"$and": {in, in.Conditions}}
	}

	out, err := json.MarshalIndent(val, "", "\t")
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// ToBsonSchema converts the parsed schema definition in to a
// BsonSchemaModel, converting nested properties, items and combinators
// recursively.
func ToBsonSchema(in codegen.TypedParam, opts ...BsonSchemaGenerationOption) (*BsonSchemaModel, error) {
	options := &BsonSchemaGenerationOptions{}
	for _, opt := range opts {
		opt(options)
	}

	out := &BsonSchemaModel{}
	typeSchema, err := in.GetSchema()
	if err != nil {
		return nil, err
	}
	title, err := typeSchema.GetTitle()
	if err != nil {
		return nil, err
	}
	description, err := typeSchema.GetDescription()
	if err != nil {
		return nil, err
	}
	properties, err := typeSchema.GetProperties()
	if err != nil {
		return nil, err
	}
	items, err := typeSchema.GetItems()
	if err != nil {
		return nil, err
	}
	specs, err := typeSchema.GetSpecs()
	if err != nil {
		return nil, err
	}
	dependencies, err := typeSchema.GetDependencies()
	if err != nil {
		return nil, err
	}
	conditions, err := typeSchema.GetConditions()
	if err != nil {
		return nil, err
	}
	typ, err := in.GetType()
	if err != nil {
		return nil, err
	}
	minValue, err := in.GetMin()
	if err != nil {
		return nil, err
	}
	maxValue, err := in.GetMax()
	if err != nil {
		return nil, err
	}
	exclusiveMinValue, err := in.GetExclusiveMin()
	if err != nil {
		return nil, err
	}
	exclusiveMaxValue, err := in.GetExclusiveMax()
	if err != nil {
		return nil, err
	}
	minLenValue, err := in.GetMinLen()
	if err != nil {
		return nil, err
	}
	maxLenValue, err := in.GetMaxLen()
	if err != nil {
		return nil, err
	}
	patternValue, err := in.GetPattern()
	if err != nil {
		return nil, err
	}

	allowNull, err := typeSchema.GetAllowNull()
	if err != nil {
		return nil, err
	}

	minItemsValue, err := in.GetMinItems()
	if err != nil {
		return nil, err
	}
	maxItemsValue, err := in.GetMaxItems()
	if err != nil {
		return nil, err
	}
	uniqueItemsValue, err := in.GetUniqueItems()
	if err != nil {
		return nil, err
	}
	additionalItemsValue, err := in.GetAdditionalItems()
	if err != nil {
		return nil, err
	}
	additionalPropertiesValue, err := in.GetAdditionalProperties()
	if err != nil {
		return nil, err
	}
	maxPropertiesValue, err := in.GetMaxProperties()
	if err != nil {
		return nil, err
	}
	minPropertiesValue, err := in.GetMinProperties()
	if err != nil {
		return nil, err
	}
	enum, err := in.GetEnum()
	if err != nil {
		return nil, err
	}
	rootCombinators, err := in.GetRootCombinators()
	if err != nil {
		return nil, err
	}

	out.Title = title
	out.Description = description
	out.Min = minValue
	out.Max = maxValue
	out.ExclusiveMin = exclusiveMinValue
	out.ExclusiveMax = exclusiveMaxValue
	out.MinLen = minLenValue
	out.MaxLen = maxLenValue
	out.Pattern = patternValue
	out.MinItems = minItemsValue
	out.MaxItems = maxItemsValue
	out.UniqueItems = uniqueItemsValue
	out.AdditionalItems = additionalItemsValue
	out.AdditionalProperties = additionalPropertiesValue
	out.MaxProperties = maxPropertiesValue
	out.MinProperties = minPropertiesValue
	out.Enum = enum
	out.Dependencies = dependencies
	out.Conditions = conditions

	out.Properties = make(map[string]*BsonSchemaModel)

	required, err := typeSchema.GetRequiredStringArray()
	if err != nil {
		return nil, err
	}
	out.Required = append(out.Required, required...)

	// Merge combinators declared on the schema itself.
	for _, rootCombinator := range rootCombinators {
		bsonSchema, err := ToBsonSchema(rootCombinator)
		if err != nil {
			return nil, err
		}
		out.OneOf = append(out.OneOf, bsonSchema.OneOf...)
		out.AnyOf = append(out.AnyOf, bsonSchema.AnyOf...)
		out.AllOf = append(out.AllOf, bsonSchema.AllOf...)
		out.Not = bsonSchema.Not
	}

	// Combinator schemas are built from their specs; all others from their
	// type, properties and items.
	switch firstOrDefault(typ) {
	case "oneof":
		{
			items := make([]*BsonSchemaModel, 0)
			for _, item := range specs {
				bsonSchema, err := ToBsonSchema(item)
				if err != nil {
					return nil, err
				}
				items = append(items, bsonSchema)
			}
			out.OneOf = items
		}
	case "anyof":
		{
			items := make([]*BsonSchemaModel, 0)
			for _, item := range specs {
				bsonSchema, err := ToBsonSchema(item)
				if err != nil {
					return nil, err
				}
				items = append(items, bsonSchema)
			}
			out.AnyOf = items
		}
	case "allof":
		{

			items := make([]*BsonSchemaModel, 0)
			for _, item := range specs {
				bsonSchema, err := ToBsonSchema(item)
				if err != nil {
					return nil, err
				}
				items = append(items, bsonSchema)
			}
			out.AllOf = items
		}
	case "not":
		{
			// Not takes a single schema; only the first spec is used.
			for i := 0; i < 1 && i < len(specs); i++ {
				bsonSchema, err := ToBsonSchema(specs[i])
				if err != nil {
					return nil, err
				}
				out.Not = bsonSchema
			}
		}
	default:
		{
			out.Type = typ
			if options.nullableItems {
				out.Type = append(out.Type, "null")
			}
			for key, value := range properties {
				bsonSchema, err := ToBsonSchema(value, WithNullableItems(allowNull))
				if err != nil {
					return nil, err
				}
				required, err := value.GetRequiredOrFalse()
				if err != nil {
					return nil, err
				}
				if required == true {
					out.Required = append(out.Required, key)
				}

				out.Properties[key] = bsonSchema
			}
			if items != nil {

				switch t := items.(type) {
				case codegen.Items:
					{
						out.Type = []string{"array"}
						if options.nullableItems {
							out.Type = append(out.Type, "null")
						}
						items := make([]*BsonSchemaModel, 0)
						for _, item := range t {
							bsonSchema, err := ToBsonSchema(item, WithNullableItems(allowNull))
							if err != nil {
								return nil, err
							}
							items = append(items, bsonSchema)
						}
						out.Items = items
					}
				case codegen.TypedParam:
					{
						out.Type = []string{"array"}
						if options.nullableItems {
							out.Type = append(out.Type, "null")
						}
						val, err := ToBsonSchema(t, WithNullableItems(allowNull))
						if err != nil {
							return nil, err
						}
						out.Items = val
					}
				default:
					{
						return nil, fmt.Errorf("expected either `Items` or `TypedParam` but got `%T`", t)
					}
				}
			}

		}
	}

	return out, nil
}

// firstOrDefault returns the first element of in, or the zero value if in is empty.
func firstOrDefault[T any](in []T) T {
	if len(in) == 0 {
		var zero T
		return zero
	}
	return in[0]
}
