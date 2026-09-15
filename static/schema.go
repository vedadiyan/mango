package static

type (
	Type interface {
		void()
	}

	Property interface {
		void()
	}

	StringDefinition struct {
		MinLen  int
		MaxLen  int
		Pattern string
	}

	DoubleDefinition struct {
		Min          int
		Max          int
		ExclusiveMin bool
		ExclusiveMax bool
		MultipleOf   float64
	}

	DecimalDefinition struct {
		DoubleDefinition
	}

	LongDefinition struct {
		IntDefinition
	}

	IntDefinition struct {
		Min          int
		Max          int
		ExclusiveMin bool
		ExclusiveMax bool
		MultipleOf   int
	}

	Scalar struct {
		Title       string
		Description string
		BsonName    string
		Type        []Type
		Required    bool
	}

	Composite struct {
		Title        string
		Description  string
		BsonName     string
		Required     bool
		Properties   map[string]Property
		Dependencies map[string][]string
	}

	Array[T Scalar | Composite] struct {
		Value T
	}

	OneOf map[string]Property
	AnyOf map[string]Property
	AllOf map[string]Property

	Schema struct {
		Title        string
		Description  string
		Properties   map[string]Property
		Dependencies map[string][]string
	}

	BasicType string
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

func (BasicType) void() {
	justPanic()
}

func (StringDefinition) void() {
	justPanic()
}

func (DoubleDefinition) void() {
	justPanic()
}

func (IntDefinition) void() {
	justPanic()
}

func (Scalar) void() {
	justPanic()
}

func (Composite) void() {
	justPanic()
}

func (Array[T]) void() {
	justPanic()
}
