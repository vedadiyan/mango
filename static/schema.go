package static

type (
	Type interface {
		typ()
	}

	Index struct {
		Order      int
		Unique     bool
		Sparse     bool
		Background bool
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
		Index        Index
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
		Index                Index
	}

	Combinator[T OneOf | AnyOf | AllOf | Not] struct {
		Title       string
		Description string
		Required    bool
		Specs       T
		Index       Index
	}

	Array struct {
		Items           Items
		MinItems        int
		MaxItems        int
		UniqueItems     bool
		AdditionalItems bool
		Index           Index
	}

	OneOf Items
	AnyOf Items
	AllOf Items
	Not   [1]Type

	Schema struct {
		Title        string
		Description  string
		Properties   Properties
		Dependencies Dependencies
	}

	BasicType string

	Items        []Type
	Properties   map[string]Type
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

func (Scalar) typ() {
	justPanic()
}

func (Composite) typ() {
	justPanic()
}

func (Array) typ() {
	justPanic()
}
