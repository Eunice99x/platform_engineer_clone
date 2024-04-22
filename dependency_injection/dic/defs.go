package dic

import (
	"errors"

	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"

	middlewares "platform_engineer_clone/api/v0/middlewares"
	token1 "platform_engineer_clone/api/v0/token"
	token "platform_engineer_clone/business/v0/token"
	config "platform_engineer_clone/src/config"
	mysql "platform_engineer_clone/src/persistence/mysql"
	token2 "platform_engineer_clone/src/persistence/mysql/v0/token"
	user "platform_engineer_clone/src/persistence/mysql/v0/user"
)

func getDiDefs(provider dingo.Provider) []di.Def {
	return []di.Def{
		{
			Name:  "api_middlewares",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("api_middlewares")
				if err != nil {
					var eo *middlewares.AuthRoutes
					return eo, err
				}
				pi0, err := ctn.SafeGet("mysql_user_persistence")
				if err != nil {
					var eo *middlewares.AuthRoutes
					return eo, err
				}
				p0, ok := pi0.(*user.PersistenceUser)
				if !ok {
					var eo *middlewares.AuthRoutes
					return eo, errors.New("could not cast parameter 0 to *user.PersistenceUser")
				}
				b, ok := d.Build.(func(*user.PersistenceUser) (*middlewares.AuthRoutes, error))
				if !ok {
					var eo *middlewares.AuthRoutes
					return eo, errors.New("could not cast build function to func(*user.PersistenceUser) (*middlewares.AuthRoutes, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
		{
			Name:  "api_token",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("api_token")
				if err != nil {
					var eo *token1.APIToken
					return eo, err
				}
				pi0, err := ctn.SafeGet("business_token")
				if err != nil {
					var eo *token1.APIToken
					return eo, err
				}
				p0, ok := pi0.(*token.BusinessToken)
				if !ok {
					var eo *token1.APIToken
					return eo, errors.New("could not cast parameter 0 to *token.BusinessToken")
				}
				b, ok := d.Build.(func(*token.BusinessToken) (*token1.APIToken, error))
				if !ok {
					var eo *token1.APIToken
					return eo, errors.New("could not cast build function to func(*token.BusinessToken) (*token1.APIToken, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
		{
			Name:  "business_token",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("business_token")
				if err != nil {
					var eo *token.BusinessToken
					return eo, err
				}
				pi0, err := ctn.SafeGet("config")
				if err != nil {
					var eo *token.BusinessToken
					return eo, err
				}
				p0, ok := pi0.(*config.Config)
				if !ok {
					var eo *token.BusinessToken
					return eo, errors.New("could not cast parameter 0 to *config.Config")
				}
				pi1, err := ctn.SafeGet("mysql_token_persistence")
				if err != nil {
					var eo *token.BusinessToken
					return eo, err
				}
				p1, ok := pi1.(*token2.PersistenceToken)
				if !ok {
					var eo *token.BusinessToken
					return eo, errors.New("could not cast parameter 1 to *token2.PersistenceToken")
				}
				b, ok := d.Build.(func(*config.Config, *token2.PersistenceToken) (*token.BusinessToken, error))
				if !ok {
					var eo *token.BusinessToken
					return eo, errors.New("could not cast build function to func(*config.Config, *token2.PersistenceToken) (*token.BusinessToken, error)")
				}
				return b(p0, p1)
			},
			Unshared: false,
		},
		{
			Name:  "config",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("config")
				if err != nil {
					var eo *config.Config
					return eo, err
				}
				b, ok := d.Build.(func() (*config.Config, error))
				if !ok {
					var eo *config.Config
					return eo, errors.New("could not cast build function to func() (*config.Config, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "mysql_connection",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("mysql_connection")
				if err != nil {
					var eo *mysql.MYSQLConnection
					return eo, err
				}
				pi0, err := ctn.SafeGet("config")
				if err != nil {
					var eo *mysql.MYSQLConnection
					return eo, err
				}
				p0, ok := pi0.(*config.Config)
				if !ok {
					var eo *mysql.MYSQLConnection
					return eo, errors.New("could not cast parameter 0 to *config.Config")
				}
				b, ok := d.Build.(func(*config.Config) (*mysql.MYSQLConnection, error))
				if !ok {
					var eo *mysql.MYSQLConnection
					return eo, errors.New("could not cast build function to func(*config.Config) (*mysql.MYSQLConnection, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
		{
			Name:  "mysql_token_persistence",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("mysql_token_persistence")
				if err != nil {
					var eo *token2.PersistenceToken
					return eo, err
				}
				pi0, err := ctn.SafeGet("mysql_connection")
				if err != nil {
					var eo *token2.PersistenceToken
					return eo, err
				}
				p0, ok := pi0.(*mysql.MYSQLConnection)
				if !ok {
					var eo *token2.PersistenceToken
					return eo, errors.New("could not cast parameter 0 to *mysql.MYSQLConnection")
				}
				b, ok := d.Build.(func(*mysql.MYSQLConnection) (*token2.PersistenceToken, error))
				if !ok {
					var eo *token2.PersistenceToken
					return eo, errors.New("could not cast build function to func(*mysql.MYSQLConnection) (*token2.PersistenceToken, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
		{
			Name:  "mysql_user_persistence",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("mysql_user_persistence")
				if err != nil {
					var eo *user.PersistenceUser
					return eo, err
				}
				pi0, err := ctn.SafeGet("mysql_connection")
				if err != nil {
					var eo *user.PersistenceUser
					return eo, err
				}
				p0, ok := pi0.(*mysql.MYSQLConnection)
				if !ok {
					var eo *user.PersistenceUser
					return eo, errors.New("could not cast parameter 0 to *mysql.MYSQLConnection")
				}
				b, ok := d.Build.(func(*mysql.MYSQLConnection) (*user.PersistenceUser, error))
				if !ok {
					var eo *user.PersistenceUser
					return eo, errors.New("could not cast build function to func(*mysql.MYSQLConnection) (*user.PersistenceUser, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
	}
}
