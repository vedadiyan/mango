package static

type (
	Type interface {
		void()
	}

	Property interface {
		void()
	}

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
		MinLen       int
		MaxLen       int
		Pattern      string
		Enum         []any
	}

	Composite struct {
		Title                string
		Description          string
		Required             bool
		Properties           Properties
		Dependencies         Dependencies
		MaxProperties        int
		MinProperties        int
		AdditionalProperties bool
	}

	Array struct {
		Items           []Type
		MinItems        int
		MaxItems        int
		UniqueItems     bool
		AdditionalItems bool
	}

	OneOf Properties
	AnyOf Properties
	AllOf Properties

	Schema struct {
		Title        string
		Description  string
		Properties   Properties
		Dependencies Dependencies
	}

	BasicType string

	Properties   map[string]Property
	Dependencies map[string][]string
)

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
)

func justPanic() {
	panic("this method should never be called")
}

func (Scalar) void() {
	justPanic()
}

func (Composite) void() {
	justPanic()
}

func (Array) void() {
	justPanic()
}
