package token

import (
	"context"
	"database/sql"
	"github.com/dembygenesis/platform_engineer_exam/models"
	"github.com/dembygenesis/platform_engineer_exam/src/persistence/mysql"
	"github.com/dembygenesis/platform_engineer_exam/src/persistence/mysql/models_schema"
	"github.com/dembygenesis/platform_engineer_exam/src/utils/common"
	"github.com/friendsofgo/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"time"
)

type PersistenceToken struct {
	db               *sql.DB
	mockRandomString string
	mockCreatedTime  time.Time
}

var (
	errCheckUniqueToken        = errors.New("error checking for unique tokens")
	errFetchToken              = errors.New("error fetching token")
	errFetchTokenByKey         = errors.New("error fetching token by key")
	errFetchTokenByKeyNoResult = errors.New("error, fetching token by key yields no results")
	errFetchTokens             = errors.New("error fetching tokens")
	errInsertNewToken          = errors.New("error inserting new token")
	errTokenNotFound           = errors.New("error, token not found")
	errUpdateTokenToExpired    = errors.New("error updating token as expired")
	errUpdateTokenToRevoked    = errors.New("error updating token as revoked")
)

func (p *PersistenceToken) RevokeToken(ctx context.Context, key string) error {
	token, err := models_schema.Tokens(
		models_schema.TokenWhere.Key.EQ(key),
		models_schema.TokenWhere.Revoked.EQ(false),
	).One(ctx, p.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errTokenNotFound
		}
		return errors.Wrap(err, errFetchToken.Error())
	}
	token.Revoked = true
	_, err = token.Update(ctx, p.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, errUpdateTokenToRevoked.Error())
	}
	return nil
}

func (p *PersistenceToken) UpdateTokenToExpired(ctx context.Context, token *models.Token) error {
	tokenEntry, err := models_schema.FindToken(ctx, p.db, token.Id)
	if err != nil {
		return errors.Wrap(err, errFetchToken.Error())
	}

	tokenEntry.Expired = true
	_, err = tokenEntry.Update(ctx, p.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, errUpdateTokenToExpired.Error())
	}

	return nil
}

func (p *PersistenceToken) GetToken(ctx context.Context, key string) (*models.Token, error) {
	var container []models.Token
	err := models_schema.Tokens(
		models_schema.TokenWhere.Key.EQ(key),
	).Bind(ctx, p.db, &container)
	if err != nil {
		return nil, errors.Wrap(err, errFetchTokenByKey.Error())
	}
	if len(container) == 0 || container == nil {
		return nil, errFetchTokenByKeyNoResult
	}
	return &container[0], nil
}

// GetAll returns all the tokens
func (p *PersistenceToken) GetAll(ctx context.Context) ([]models.Token, error) {
	container := []models.Token{}
	err := models_schema.Tokens(
		qm.InnerJoin("user u ON u.id = token.created_by"),
		qm.Select([]string{
			"token.id AS id",
			"token.key AS `key`",
			"token.created_at AS created_at",
			"token.revoked AS revoked",
			"token.expired AS expired",
			"token.expires_at AS expires_at",
			"u.name AS created_by",
		}...),
	).Bind(mysql.BoilCtx, p.db, &container)
	if err != nil {
		return nil, errors.Wrap(err, errFetchTokens.Error())
	}
	return container, nil
}

// Generate returns a unique string in the length range of 6-12 characters
func (p *PersistenceToken) Generate(ctx context.Context, createdBy int, randomCharMinLength int,
	randomCharMaxLength int) (string, error) {
	logger := common.GetLogger(ctx)
	var randomString string
	tokenVerifiedUnique := false
	loops := 0
	for !tokenVerifiedUnique {
		if loops > 0 {
			logger.WithFields(logrus.Fields{
				"msg":   "number of loops it took to generate the unique key",
				"loops": loops,
			}).Info("info_generate")
		}
		loops++

		randomizedCharLength := rand.Intn(randomCharMaxLength-randomCharMinLength) + randomCharMinLength
		randomString = generateRandomCharacters(randomizedCharLength)

		if p.mockRandomString != "" {
			randomString = p.mockRandomString
		}

		token, err := models_schema.Tokens(
			models_schema.TokenWhere.Key.EQ(randomString),
		).All(mysql.BoilCtx, p.db)
		if err != nil {
			return "", errors.Wrap(err, errCheckUniqueToken.Error())
		}
		if len(token) == 0 {
			tokenVerifiedUnique = true
		}
	}

	createdAt := time.Now()
	if !p.mockCreatedTime.IsZero() {
		createdAt = p.mockCreatedTime
	}

	newToken := models_schema.Token{
		Key:       randomString,
		CreatedBy: createdBy,
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(7 * time.Hour * 24),
	}

	err := newToken.Insert(mysql.BoilCtx, p.db, boil.Infer())
	if err != nil {
		return "", errors.Wrap(err, errInsertNewToken.Error())
	}

	return randomString, nil
}

// NewPersistenceToken returns a new *PersistenceToken instance
func NewPersistenceToken(db *sql.DB) *PersistenceToken {
	return &PersistenceToken{db: db}
}
