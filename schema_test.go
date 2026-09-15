package mango

import "github.com/vedadiyan/mango/static"

func Export() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: map[string]static.Property{
			"FirstName": static.Scalar{
				Title:    "Test First Name",
				Type:     []static.Type{static.TypeString},
				Required: true,
			},
		},
	}
}
