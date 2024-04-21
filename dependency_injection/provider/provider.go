package provider

import (
	"github.com/pkg/errors"
	"github.com/sarulabs/dingo/v4"
)

type Provider struct {
	dingo.BaseProvider
}

// getServices is the main configuration func that produces the singleton
func getServices() (*[]dingo.Def, error) {
	var services []dingo.Def

	services = append(services, *getConfigLayers()...)
	services = append(services, *getBusinessLayers()...)
	services = append(services, *getPersistenceLayers()...)
	services = append(services, *getAPILayers()...)

	return &services, nil
}

// Load bootstrap the dependencies
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
