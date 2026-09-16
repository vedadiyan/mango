package mango

import "github.com/vedadiyan/mango/static"

func Export() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"FirstName": static.Array{
				Items: static.Items{
					static.Scalar{
						Type:     []static.BasicType{static.TypeString},
						MinLen:   3,
						MaxLen:   100,
						Required: true,
					},
				},
			},
			"NestObject": static.Composite{
				Title: "Test Nested Object",
				Properties: static.Properties{
					"A": static.Scalar{
						Type:     []static.BasicType{static.TypeInt},
						Min:      10,
						Max:      10000,
						Required: true,
					},
				},
				Dependencies: static.Dependencies{
					"A": {"email"},
				},
			},
		},
		Conditions: static.Conditions{
			{"$exists": static.Condition{
				"X": 1,
			}},
		},
	}
}
