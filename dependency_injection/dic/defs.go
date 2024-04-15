package dic

import (
	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"
)

func getDiDefs(provider dingo.Provider) []di.Def {
	return []di.Def{
		{
			Name:  "Hello",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &provider.HelloPrinter{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "World",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &provider.WorldPrinter{}, nil
			},
			Unshared: false,
		},
	}
}
