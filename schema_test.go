package mango

import "github.com/vedadiyan/mango/static"

func Export() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: map[string]static.Property{
			"FirstName": static.ScalarArray{
				Title:    "Test First Name",
				Type:     []static.Type{static.TypeString},
				Required: true,
			},
			"NestObject": static.Composite{
				Title: "Test Nested Object",
				Properties: map[string]static.Property{
					"A": static.Scalar{
						Type: []static.Type{static.TypeBool},
					},
				},
			},
		},
	}
}
