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

		ObjectName string
	}

	Composite struct {
		Title                string
		Description          string
		Required             any
		Properties           Properties
		Dependencies         Dependencies
		MaxProperties        int
		MinProperties        int
		AdditionalProperties bool
		Index                Index

		OneOf Combinator[OneOf]
		AnyOf Combinator[AnyOf]
		AllOf Combinator[AllOf]
		Not   Combinator[Not]

		ObjectName string
	}

	Combinator[T OneOf | AnyOf | AllOf | Not] struct {
		Title       string
		Description string
		Specs       T
		Index       Index

		ObjectName string
	}

	Array[T Composite | Scalar] struct {
		Required        bool
		Items           T
		MinItems        int
		MaxItems        int
		UniqueItems     bool
		AdditionalItems bool
		Index           Index

		ObjectName string
	}

	NotCondition struct {
		Required []string
	}

	OneOf Items
	AnyOf Items
	AllOf Items
	Not   [1]NotCondition

	Schema struct {
		Title        string
		Description  string
		Properties   Properties
		Dependencies Dependencies
		Conditions   Conditions
		OneOf        Combinator[OneOf]
		AnyOf        Combinator[AnyOf]
		AllOf        Combinator[AllOf]
		Not          Combinator[Not]
	}

	BasicType  string
	BasicTypes []BasicType

	Items        []Type
	Properties   map[string]Type
	Dependencies map[string][]string

	Condition  map[string]any
	Conditions []Condition
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

func (Combinator[T]) typ() {
	justPanic()
}

func (Array[T]) typ() {
	justPanic()
}
