package generators

import (
	"github.com/vedadiyan/mango/static"
)

func DefineExport() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"products": static.Array[static.Array[static.Composite]]{
				Required: true,
				Items: static.Array[static.Composite]{
					Items: static.Composite{
						Properties: static.Properties{
							"Name": static.Scalar{
								Type: static.BasicTypes{static.TypeString},
							},
						},
					},
				},
			},
		},
	}
}
