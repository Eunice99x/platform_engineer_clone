package provider

import (
	"fmt"
	"github.com/sarulabs/dingo/v4"
)

type WorldPrinter struct{}

func (hp *WorldPrinter) Print() {
	fmt.Println("Hello")
}

func getWorld() *[]dingo.Def {
	return &[]dingo.Def{
		{Name: "World",
			Build: (*WorldPrinter)(nil),
		},
	}
}
