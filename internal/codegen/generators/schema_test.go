package generators

import "github.com/vedadiyan/mango/static"

func DefineExport() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"first_name": static.Scalar{
				Type: []static.BasicType{static.TypeString},
			},
			"last_name": static.Scalar{
				Type: []static.BasicType{static.TypeString},
			},
			"products": static.Array{
				Items: static.Items{
					static.Composite{
						Properties: static.Properties{
							"id": static.Scalar{
								Type: []static.BasicType{static.TypeInt},
							},
							"name": static.Scalar{
								Type: []static.BasicType{static.TypeString},
							},
						},
					},
				},
			},
			"address": static.Composite{
				Properties: static.Properties{
					"postal_code": static.Scalar{
						Type: []static.BasicType{static.TypeString},
					},
					"street": static.Scalar{
						Type: []static.BasicType{static.TypeString},
					},
				},
			},
			"nd_array": static.Array{
				Items: static.Items{
					static.Array{
						Items: static.Items{
							static.Array{
								Items: static.Items{
									static.Scalar{
										Type: []static.BasicType{static.TypeInt},
									},
								},
							},
						},
					},
				},
			},
			"details": static.Combinator[static.OneOf]{
				Specs: static.OneOf{
					static.Composite{
						ObjectName: "individual",
						Properties: static.Properties{
							"customer_name": static.Scalar{
								Type: []static.BasicType{static.TypeString},
							},
							"email": static.Scalar{
								Type: []static.BasicType{static.TypeString},
							},
						},
					},
					static.Composite{
						ObjectName: "legal",
						Properties: static.Properties{
							"company_name": static.Scalar{
								Type: []static.BasicType{static.TypeString},
							},
							"tax_id": static.Scalar{
								Type: []static.BasicType{static.TypeString},
							},
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
