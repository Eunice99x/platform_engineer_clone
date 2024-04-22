package provider

import (
	"github.com/sarulabs/dingo/v4"
	BusinessToken "platform_engineer_clone/business/v0/token"
	"platform_engineer_clone/src/config"
	PersistenceToken "platform_engineer_clone/src/persistence/mysql/v0/token"
)

const (
	businessToken = "business_token"
)

func getBusinessLayers() *[]dingo.Def {
	return &[]dingo.Def{
		{
			Name: businessToken,
			Build: func(config *config.Config, persistenceToken *PersistenceToken.PersistenceToken) (*BusinessToken.BusinessToken, error) {
				return BusinessToken.NewBusinessToken(
					persistenceToken,
					config.App.TokenDaysValid,
					config.App.RandomCharMinLength,
					config.App.RandomCharMaxLength,
				), nil
			},
		},
	}
}
