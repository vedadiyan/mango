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
			"NestObject": static.Combinator[static.AnyOf]{
				Title: "Test Nested Object",
				Specs: static.AnyOf{
					static.Scalar{
						Type:     []static.BasicType{static.TypeString},
						MinLen:   3,
						MaxLen:   100,
						Required: true,
					},
				},
			},
		},
	}
}
