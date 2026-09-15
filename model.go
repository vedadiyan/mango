package mango

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
