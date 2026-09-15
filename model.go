package mango

import (
	"fmt"
)

type (
	BsonSchema struct {
		Title        string
		Description  string
		Type         []string
		Properties   map[string]*BsonSchema
		Items        map[string]*BsonSchema
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

		AnyOf []*BsonSchema
		OneOf []*BsonSchema
		AllOf []*BsonSchema
	}
)

const (
	BasicType         = "github.com/vedadiyan/mango/static.BasicType"
	StringDefinition  = "github.com/vedadiyan/mango/static.StringDefinition"
	DoubleDefinition  = "github.com/vedadiyan/mango/static.DoubleDefinition"
	DecimalDefinition = "github.com/vedadiyan/mango/static.DecimalDefinition"
	LongDefinition    = "github.com/vedadiyan/mango/static.LongDefinition"
	IntDefinition     = "github.com/vedadiyan/mango/static.IntDefinition"
)

func ToBasonSchema(in TypedParam) (*BsonSchema, error) {
	if in.TypeName != "" {
		return nil, fmt.Errorf("")
	}
	typeValue, ok := in.Value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("")
	}

	out := &BsonSchema{}

	for key, value := range typeValue {
		switch key {
		case "Title":
			{
				title, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("title should be a string")
				}
				out.Title = title

			}
		case "Description":
			{
				description, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("description should be a string")
				}
				out.Description = description
			}
		case "Type":
			{
				typeName, ok := value.(TypedParam)
				if !ok {
					return nil, fmt.Errorf("cannot determine type parameter")
				}

				values, ok := typeName.Value.([]any)
				if !ok {
					return nil, fmt.Errorf("expected array but found %T", typeName.Value)
				}
				types := make([]string, 0)
				for _, val := range values {
					typeName, ok := val.(TypedParam)
					if !ok {
						return nil, fmt.Errorf("cannot determine type parameter")
					}

					switch typeName.TypeName {
					case BasicType:
						{
							str, ok := typeName.Value.(string)
							if !ok {
								return nil, fmt.Errorf("expected string but found %T", typeName.Value)
							}
							types = append(types, str)
						}
					case StringDefinition:
						{
							sd, ok := typeName.Value.(map[string]any)
							if !ok {
								return nil, fmt.Errorf("expected map[string]any but found %T", typeName.Value)
							}
							types = append(types, "string")
							for k, v := range sd {
								switch k {
								case "MinLen":
									{
										i, ok := v.(int)
										if !ok {
											return nil, fmt.Errorf("expected int but found %T", v)
										}
										out.MinLen = &i
									}
								case "MaxLen":
									{
										i, ok := v.(int)
										if !ok {
											return nil, fmt.Errorf("expected int but found %T", v)
										}
										out.MaxLen = &i
									}
								case "Pattern":
									{
										i, ok := v.(string)
										if !ok {
											return nil, fmt.Errorf("expected string but found %T", v)
										}
										out.Pattern = &i
									}
								}
							}
						}
					case DoubleDefinition, DecimalDefinition, LongDefinition, IntDefinition:
						{
							sd, ok := typeName.Value.(map[string]any)
							if !ok {
								return nil, fmt.Errorf("expected map[string]any but found %T", typeName.Value)
							}
							types = append(types, "string")
							for k, v := range sd {
								switch k {
								case "Min":
									{
										i, ok := v.(int)
										if !ok {
											return nil, fmt.Errorf("expected int but found %T", v)
										}
										out.Min = &i
									}
								case "Max":
									{
										i, ok := v.(int)
										if !ok {
											return nil, fmt.Errorf("expected int but found %T", v)
										}
										out.Max = &i
									}
								case "ExclusiveMin":
									{
										i, ok := v.(bool)
										if !ok {
											return nil, fmt.Errorf("expected bool but found %T", v)
										}
										out.ExclusiveMin = &i
									}
								case "ExclusiveMax":
									{
										i, ok := v.(bool)
										if !ok {
											return nil, fmt.Errorf("expected bool but found %T", v)
										}
										out.ExclusiveMax = &i
									}
								case "MultipleOf":
									{
										switch t := v.(type) {
										case int:
											{
												f := float64(t)
												out.MultipleOf = &f
											}
										case float64:
											{
												out.MultipleOf = &t
											}
										default:
											{
												return nil, fmt.Errorf("expected string but found %T", v)
											}
										}
									}
								}
							}
						}
					}

				}
			}
		case "Properties":
			{
				typeName, ok := value.(TypedParam)
				if !ok {
					return nil, fmt.Errorf("cannot determine type parameter")
				}
				values, ok := typeName.Value.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("expected map[string]any but found %T", typeName.Value)
				}
				o := make(map[string]*BsonSchema)
				for key, value := range values {
					typeName, ok := value.(TypedParam)
					if !ok {
						return nil, fmt.Errorf("cannot determine type parameter")
					}
					res, err := ToBasonSchema(typeName)
					if err != nil {
						return nil, err
					}
					o[key] = res
				}
				out.Properties = o
			}
		}
	}

	return out, nil
}
