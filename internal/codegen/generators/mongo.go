package generators

import (
	"encoding/json"
	"fmt"

	"github.com/vedadiyan/mango/internal/codegen"
)

type (
	Index struct {
		Order      *int
		Unique     *bool
		Sparse     *bool
		Background *bool
	}
	BsonSchema struct {
		Title        *string                `json:"title,omitempty"`
		Description  *string                `json:"description,omitempty"`
		Type         []string               `json:"bsonType,omitempty"`
		Properties   map[string]*BsonSchema `json:"properties,omitempty"`
		Items        any                    `json:"items,omitempty"`
		Required     []string               `json:"required,omitempty"`
		Dependencies map[string][]string    `json:"dependencies,omitempty"`
		Enum         []any                  `json:"enum,omitempty"`

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

		AnyOf []*BsonSchema `json:"anyOf,omitempty"`
		OneOf []*BsonSchema `json:"oneOf,omitempty"`
		AllOf []*BsonSchema `json:"allOf,omitempty"`
		Not   *BsonSchema   `json:"Not,omitempty"`

		index Index

		Conditions []map[string]any `json:"-"`
	}
)

func ToMongoValidationSchema(in *BsonSchema) (string, error) {
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

func ToBsonSchema(in codegen.TypedParam) (*BsonSchema, error) {
	out := &BsonSchema{}
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
	oneOfs, err := in.GetOneOfs()
	if err != nil {
		return nil, err
	}

	_ = oneOfs

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

	out.Properties = make(map[string]*BsonSchema)

	required, err := typeSchema.GetRequiredStringArray()
	if err != nil {
		return nil, err
	}
	out.Required = append(out.Required, required...)

	if oneOfs != nil {
		items := make([]*BsonSchema, 0)
		for _, item := range oneOfs {
			bsonSchema, err := ToBsonSchema(item)
			if err != nil {
				return nil, err
			}
			for _, oneOf := range bsonSchema.OneOf {
				oneOf.Required = append(oneOf.Required, bsonSchema.Required...)
				items = append(items, oneOf)
			}
		}
		out.OneOf = items
	}

	switch firstOrDefault(typ) {
	case "oneof":
		{
			items := make([]*BsonSchema, 0)
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
			items := make([]*BsonSchema, 0)
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

			items := make([]*BsonSchema, 0)
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
			for key, value := range properties {
				bsonSchema, err := ToBsonSchema(value)
				if err != nil {
					return nil, err
				}
				required, err := value.GetRequired()
				if err != nil {
					return nil, err
				}
				if required != nil && *required == true {
					out.Required = append(out.Required, key)
				}
				out.Properties[key] = bsonSchema
			}
			if items != nil {

				switch t := items.(type) {
				case codegen.Items:
					{
						out.Type = []string{"array"}
						items := make([]*BsonSchema, 0)
						for _, item := range t {
							bsonSchema, err := ToBsonSchema(item)
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
						val, err := ToBsonSchema(t)
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

func firstOrDefault[T any](in []T) T {
	if len(in) == 0 {
		var zero T
		return zero
	}
	return in[0]
}
