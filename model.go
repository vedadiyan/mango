package mango

type (
	BsonSchema struct {
		Title        *string
		Description  *string
		Type         []string
		Properties   map[string]*BsonSchema
		Items        []*BsonSchema
		Required     []string
		Dependencies map[string][]string

		MinLen  *int
		MaxLen  *int
		Pattern *string

		Min          *int
		Max          *int
		ExclusiveMin *bool
		ExclusiveMax *bool
		MultipleOf   *float64

		MinItems    *int
		MaxItems    *int
		UniqueItems *bool

		AnyOf []*BsonSchema
		OneOf []*BsonSchema
		AllOf []*BsonSchema
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
