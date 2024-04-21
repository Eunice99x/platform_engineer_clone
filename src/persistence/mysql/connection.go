package mysql

import (
	"database/sql"
	"fmt"
	"github.com/dembygenesis/platform_engineer_exam/src/config"
	"github.com/dembygenesis/platform_engineer_exam/src/utils/strings"
	"github.com/friendsofgo/errors"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type MYSQLConnection struct {
	DB *sql.DB
}

var (
	errDatabasePropertyNil = errors.New("db property is nil")
	errDatabasePingError   = errors.New("db ping failure")
)

// Ping checks if the sql instance can be reached
func (c *MYSQLConnection) Ping() error {
	if c.DB == nil {
		return errDatabasePropertyNil
	}
	err := c.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

// NewMYSQLConnection returns a struct with a mysql instance
func NewMYSQLConnection(c config.DatabaseCredentials) (*MYSQLConnection, error) {
	conn := MYSQLConnection{}
	connString := c.User + ":" +
		c.Pass + "@tcp(" +
		c.Host + ":" +
		c.Port + ")/" +
		c.Database + "?charset=utf8&parseTime=true"
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, errors.Wrap(err, "error establishing mysql connection")
	}
	db.SetConnMaxLifetime(time.Minute * 60)
	conn.DB = db
	err = conn.Ping()
	if err != nil {
		fmt.Println("============ Ping err with creds", strings.GetJSON(c))
		return nil, errors.Wrap(err, errDatabasePingError.Error())
	}
	return &conn, nil
}
