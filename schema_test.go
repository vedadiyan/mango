package mango

import "github.com/vedadiyan/mango/static"

func Export() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"FirstName": static.ScalarArray{
				Title:    "Test First Name",
				Type:     []static.Type{static.TypeString},
				MinLen:   3,
				MaxLen:   100,
				Required: true,
			},
			"NestObject": static.Composite{
				Title: "Test Nested Object",
				Properties: static.Properties{
					"A": static.Scalar{
						Type: []static.Type{static.TypeBool},
					},
				},
			},
		},
	}
}
