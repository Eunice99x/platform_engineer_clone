package provider

import (
	"github.com/pkg/errors"
	"github.com/sarulabs/dingo/v4"
)

type Provider struct {
	dingo.BaseProvider
}

func getServices() (*[]dingo.Def, error) {
	var services []dingo.Def
	services = append(services, *getHello()...)
	services = append(services, *getWorld()...)

	return &services, nil
}

func (p *Provider) Load() error {
	services, err := getServices()
	if err != nil {
		return errors.Wrap(err, "error trying to load the provider")
	}

	err = p.AddDefSlice(*services)
	if err != nil {
		return errors.Wrap(err, "error adding dependency definitions")
	}

	return nil
}
