package token

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"time"
)

type dataPersistence interface {
	// Generate creates a new 6-12 digit authentication token
	Generate() (string, error)

	// Validate checks if a string is registered
	Validate(s string) error
}

type MYSQLPersistence struct {
	db *sql.DB
}

func (m *MYSQLPersistence) Generate() (string, error) {
	return "Generate", nil
}

func (m *MYSQLPersistence) Validate(s string) error {
	return nil
}

// Ping checks if we have a valid, and active MYSQL connection.
func (m *MYSQLPersistence) Ping() error {
	err := m.db.Ping()
	if err != nil {
		return errors.Wrap(err, "error getting response from a ping")
	}
	return nil
}

// establishConnection inits a MYSQL connection
func (m *MYSQLPersistence) establishConnection(
	host string,
	user string,
	pass string,
	database string,
	port string,
) error {
	connString := user + ":" +
		pass + "@tcp(" +
		host + ":" +
		port + ")/" +
		database + "?charset=utf8&parseTime=true"
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return errors.Wrap(err, "error establishing mysql connection")
	}
	db.SetConnMaxLifetime(time.Minute * 60)
	m.db = db
	return nil
}

func NewMYSQLPersistence(
	host string,
	user string,
	pass string,
	database string,
	port string,
) (*MYSQLPersistence, error) {
	m := MYSQLPersistence{}
	err := m.establishConnection(
		host,
		user,
		pass,
		database,
		port,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error establishing mysql connection")
	}
	return &m, nil
}
