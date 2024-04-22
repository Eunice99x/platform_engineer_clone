package provider

import (
	"github.com/sarulabs/dingo/v4"
	"log"
	"platform_engineer_clone/src/config"
	PersistenceMYSQL "platform_engineer_clone/src/persistence/mysql"
	PersistenceToken "platform_engineer_clone/src/persistence/mysql/v0/token"
	"platform_engineer_clone/src/persistence/mysql/v0/user"
)

const (
	mysqlConnection            = "mysql_connection"
	mysqlTokenPersistenceLayer = "mysql_token_persistence"
	mysqlUserPersistenceLayer  = "mysql_user_persistence"
)

func getPersistenceLayers() *[]dingo.Def {
	return &[]dingo.Def{
		{
			Name: mysqlConnection,
			Build: func(config *config.Config) (*PersistenceMYSQL.MYSQLConnection, error) {
				mysql, err := PersistenceMYSQL.NewMYSQLConnection(config.DatabaseCredentials)
				if err != nil {
					log.Fatalf("error establishing connection to MYSQL: :%v", err.Error())
				}
				return mysql, err
			},
		},
		{
			Name: mysqlTokenPersistenceLayer,
			Build: func(connection *PersistenceMYSQL.MYSQLConnection) (*PersistenceToken.PersistenceToken, error) {
				return PersistenceToken.NewPersistenceToken(connection.DB), nil
			},
		},
		{
			Name: mysqlUserPersistenceLayer,
			Build: func(connection *PersistenceMYSQL.MYSQLConnection) (*user.PersistenceUser, error) {
				return user.NewPersistenceUser(connection.DB), nil
			},
		},
	}
}
