package provider

import (
	"fmt"
	"github.com/sarulabs/dingo/v4"
)

type HelloPrinter struct{}

func (hp *HelloPrinter) Print() {
	fmt.Println("Hello")
}

func getHello() *[]dingo.Def {
	return &[]dingo.Def{
		{Name: "Hello",
			Build: (*HelloPrinter)(nil),
		},
	}
}
