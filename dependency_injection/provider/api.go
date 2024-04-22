package provider

import (
	"github.com/sarulabs/dingo/v4"
	"platform_engineer_clone/api/v0/middlewares"
	"platform_engineer_clone/api/v0/token"
	BusinessToken "platform_engineer_clone/business/v0/token"
	"platform_engineer_clone/src/persistence/mysql/v0/user"
)

const (
	apiToken       = "api_token"
	apiMiddlewares = "api_middlewares"
)

func getAPILayers() *[]dingo.Def {
	return &[]dingo.Def{
		{
			Name: apiToken,
			Build: func(businessToken *BusinessToken.BusinessToken) (*token.APIToken, error) {
				return token.NewAPIToken(businessToken), nil
			},
		},
		{
			Name: apiMiddlewares,
			Build: func(user *user.PersistenceUser) (*middlewares.AuthRoutes, error) {
				return middlewares.NewAuthRoutes(user), nil
			},
		},
	}
}
