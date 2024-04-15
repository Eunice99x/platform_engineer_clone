package provider

import "github.com/sarulabs/dingo/v4"

type Provider struct {
	dingo.BaseProvider
}

func getServices() (*[]dingo.Def, error) {
	var services []dingo.Def
	services = append(services, *getHello()...)
	services = append(services, *getWorld()...)

	return &services, nil
}
