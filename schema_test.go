package mango

import "github.com/vedadiyan/mango/static"

func DefineExport() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"first_name": static.Array{
				Items: static.Items{
					static.Scalar{
						Type:     []static.BasicType{static.TypeString},
						MinLen:   3,
						MaxLen:   100,
						Required: true,
					},
				},
			},
			"last_name": static.Array{
				Items: static.Items{
					static.Composite{
						Properties: static.Properties{
							"title": static.Scalar{
								Type: []static.BasicType{static.TypeString},
							},
						},
					},
				},
			},
			"nested_object": static.Composite{
				Title: "Test Nested Object",
				Properties: static.Properties{
					"A": static.Scalar{
						Type:     []static.BasicType{static.TypeInt},
						Min:      10,
						Max:      10000,
						Required: true,
					},
					"C": static.Scalar{
						Type:     []static.BasicType{static.TypeInt},
						Min:      10,
						Max:      10000,
						Required: true,
					},
					"B": static.Composite{
						Properties: static.Properties{
							"Z": static.Scalar{
								Type: []static.BasicType{static.TypeBool},
							},
						},
					},
				},
				Dependencies: static.Dependencies{
					"A": {"email"},
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
