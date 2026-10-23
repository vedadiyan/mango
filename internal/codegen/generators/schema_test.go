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
				Type:   static.BasicTypes{static.TypeString},
				MinLen: 10,
			},
			"last_name": static.Scalar{
				Type: static.BasicTypes{static.TypeString},
			},
			"products": static.Array[static.Composite]{
				Required: true,
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
			"subscription": static.Enum[static.Int]{
				Enum: []static.Int{1, 2},
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
							"inner_field": static.Composite{
								Properties: static.Properties{
									"id": static.Scalar{
										Type: static.BasicTypes{static.TypeInt},
									},
								},
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
					OneOf: static.Combinator[static.OneOf]{
						Specs: static.OneOf{
							static.Composite{
								Properties: static.Properties{
									"role": static.Scalar{
										Type:     static.BasicTypes{static.TypeString},
										Required: true,
									},
								},
							},
							static.Composite{
								Properties: static.Properties{
									"responsibility": static.Scalar{
										Type:     static.BasicTypes{static.TypeTimeStamp},
										Required: true,
									},
								},
							},
						},
					},
					Not: static.Combinator[static.Not]{
						Specs: static.Not{
							{
								Required: []string{"experience"},
							},
						},
					},
				},
			},
		},
		Not: static.Combinator[static.Not]{
			Specs: static.Not{
				{
					Required: []string{"new_field"},
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
