package generators

import (
	"github.com/vedadiyan/mango/static"
)

func DefineExport() static.Schema {
	return static.Schema{
		Title:       "Test Title",
		Description: "Test Description",
		Properties: static.Properties{
			"products": static.Composite{
				Required: true,
				Properties: static.Properties{
					"first_name": static.Scalar{
						Type: static.BasicTypes{static.TypeString},
					},
					"last_name": static.Scalar{
						Type: static.BasicTypes{static.TypeString},
					},
				},
			},
		},
	}
}
