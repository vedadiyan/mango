package static

type (
	Type interface {
		void()
	}

	Property interface {
		property()
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
		Properties           Specs
		Dependencies         Dependencies
		MaxProperties        int
		MinProperties        int
		AdditionalProperties bool
		Index                Index
	}

	Array struct {
		Items           Specs
		MinItems        int
		MaxItems        int
		UniqueItems     bool
		AdditionalItems bool
		Index           Index
	}

	OneOf Items
	AnyOf Items
	AllOf Items

	Schema struct {
		Title        string
		Description  string
		Properties   Specs
		Dependencies Dependencies
	}

	BasicType string

	Specs interface {
		spec()
	}

	Items        []Property
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

func (Scalar) property() {
	justPanic()
}

func (Composite) property() {
	justPanic()
}

func (Array) property() {
	justPanic()
}

func (Items) spec() {
	justPanic()
}

func (Properties) spec() {
	justPanic()
}
