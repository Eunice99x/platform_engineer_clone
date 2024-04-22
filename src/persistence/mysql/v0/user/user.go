package user

import (
	"database/sql"
	"github.com/friendsofgo/errors"
	"golang.org/x/crypto/bcrypt"
	"platform_engineer_clone/models"
	"platform_engineer_clone/src/persistence/mysql"
	"platform_engineer_clone/src/persistence/mysql/models_schema"
)

type PersistenceUser struct {
	db *sql.DB
}

var (
	errMatchingEmail = errors.New("error matching email")
)

// BasicAuth authenticates by first matching the user to "email", and the password to it's "encrypted" version
func (p *PersistenceUser) BasicAuth(user, pass string) (bool, *models.User, error) {
	match, err := models_schema.Users(
		models_schema.UserWhere.Email.EQ(user),
	).One(mysql.BoilCtxNoLog, p.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		} else {
			return false, nil, errors.Wrap(errMatchingEmail, err.Error())
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(match.Password), []byte(pass))
	if err != nil {
		return false, nil, nil
	}
	userMeta := models.User{
		Id:    match.ID,
		Name:  match.Name,
		Email: match.Email,
	}
	return true, &userMeta, nil
}

func NewPersistenceUser(db *sql.DB) *PersistenceUser {
	return &PersistenceUser{db}
}
