package mango

type (
	BsonSchema struct {
		Title        *string                `json:"title,omitempty"`
		Description  *string                `json:"description,omitempty"`
		Type         []string               `json:"bsonType,omitempty"`
		Properties   map[string]*BsonSchema `json:"properties,omitempty"`
		Items        []*BsonSchema          `json:"items,omitempty"`
		Required     []string               `json:"required,omitempty"`
		Dependencies map[string][]string    `json:"dependencies,omitempty"`
		Enum         []any                  `json:"enum,omitempty"`

		MinLen  *int    `json:"minLength,omitempty"`
		MaxLen  *int    `json:"maxLength,omitempty"`
		Pattern *string `json:"pattern,omitempty"`

		Min          *int     `json:"minimum,omitempty"`
		Max          *int     `json:"maximum,omitempty"`
		ExclusiveMin *bool    `json:"exclusiveMinimum,omitempty"`
		ExclusiveMax *bool    `json:"exclusiveMaximum,omitempty"`
		MultipleOf   *float64 `json:"multipleOf,omitempty"`

		MinItems    *int  `json:"minItems,omitempty"`
		MaxItems    *int  `json:"maxItems,omitempty"`
		UniqueItems *bool `json:"uniqueItems,omitempty"`

		MaxProperties *int `json:"maxProperties,omitempty"`
		MinProperties *int `json:"minProperties,omitempty"`

		AdditionalItems      *bool `json:"additionalItems,omitempty"`
		AdditionalProperties *bool `json:"additionalProperties,omitempty"`

		AnyOf []*BsonSchema `json:"anyOf,omitempty"`
		OneOf []*BsonSchema `json:"oneOf,omitempty"`
		AllOf []*BsonSchema `json:"allOf,omitempty"`
	}
)

func ToBsonSchema(in TypedParam) (*BsonSchema, error) {
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
	out.Type = typ

	out.Properties = make(map[string]*BsonSchema)
	out.Items = make([]*BsonSchema, 0)
	out.Required = make([]string, 0)

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

	for _, item := range items {
		bsonSchema, err := ToBsonSchema(item)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, bsonSchema)
	}

	return out, nil
}
