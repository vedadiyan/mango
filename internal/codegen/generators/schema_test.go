package generators

import (
	"github.com/vedadiyan/mango/static"
)

func DefineExport() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"first_name": static.Scalar{
				Type: static.BasicTypes{static.TypeString},
			},
			"last_name": static.Scalar{
				Type: static.BasicTypes{static.TypeString},
			},
			"products": static.Array[static.Composite]{
				Items: static.Composite{
					Properties: static.Properties{
						"id": static.Scalar{
							Type: static.BasicTypes{static.TypeInt},
						},
						"name": static.Scalar{
							Type: static.BasicTypes{static.TypeString},
						},
					},
				},
			},
			"subscription": static.Scalar{
				Enum: []any{"Normal", "Advanced"},
			},
			"address": static.Composite{
				Properties: static.Properties{
					"postal_code": static.Scalar{
						Type: static.BasicTypes{static.TypeString},
					},
					"street": static.Scalar{
						Type: static.BasicTypes{static.TypeString},
					},
				},
			},
			"details": static.Combinator[static.OneOf]{
				Specs: static.OneOf{
					static.Composite{
						ObjectName: "individual",
						Properties: static.Properties{
							"customer_name": static.Scalar{
								Type:     static.BasicTypes{static.TypeString},
								Required: true,
							},
							"email": static.Scalar{
								Type:     static.BasicTypes{static.TypeString},
								Required: true,
							},
						},
					},
					static.Composite{
						ObjectName: "legal",
						Properties: static.Properties{
							"company_name": static.Scalar{
								Type:     static.BasicTypes{static.TypeString},
								Required: true,
							},
							"tax_id": static.Scalar{
								Type:     static.BasicTypes{static.TypeString},
								Required: true,
							},
						},
					},
				},
			},
		},
		OneOf: static.Combinator[static.OneOf]{
			Specs: static.OneOf{
				static.Composite{
					Required: []string{"hobbie"},
					Properties: static.Properties{
						"hobbie": static.Scalar{
							Type: static.BasicTypes{static.TypeString},
						},
						"hobbie2": static.Scalar{
							Type: static.BasicTypes{static.TypeString},
						},
					},
				},
				static.Composite{
					Properties: static.Properties{
						"skills": static.Scalar{
							Type:     static.BasicTypes{static.TypeString},
							Required: true,
						},
					},
				},
			},
		},
		Conditions: static.Conditions{
			{"$exists": static.Condition{
				"X": static.Condition{
					"Y": static.Conditions{
						{
							"Z": static.Condition{
								"V": 1,
							},
						},
					},
				},
			}},
		},
	}
}
