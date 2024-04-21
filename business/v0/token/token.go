package token

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type dataPersistence interface {
	GetAll(ctx context.Context) ([]models.Token, error)
	Generate(ctx context.Context, createdBy int, randomCharMinLength int, randomCharMaxLength int) (string, error)
	GetToken(ctx context.Context, key string) (*models.Token, error)
	UpdateTokenToExpired(ctx context.Context, token *models.Token) error
	RevokeToken(ctx context.Context, key string) error
}

type BusinessToken struct {
	dataLayer           dataPersistence
	tokenDaysValid      int
	randomCharMinLength int
	randomCharMaxLength int
}

var (
	errGenerateToken          = errors.New("error generating token")
	errGetToken               = errors.New("error, Get fails")
	errGetTokens              = errors.New("error, get all fails")
	errRevokeToken            = errors.New("error revoking token")
	errTokenRevoked           = errors.New("error, token is revoked")
	errTokenExpired           = errors.New("error, token has already expired")
	errTokenDeterminedExpired = errors.New("error, token has already expired")
	errUpdateTokenToExpired   = errors.New("error, updating token to expired failed")
)

func (b *BusinessToken) GetAll(ctx context.Context) ([]models.Token, error) {
	tokens, err := b.dataLayer.GetAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errGetTokens.Error())
	}
	return tokens, nil
}

func (b *BusinessToken) Generate(ctx context.Context, user *models.User) (string, error) {
	tokenKey, err := b.dataLayer.Generate(ctx, user.Id, b.randomCharMinLength, b.randomCharMaxLength)
	if err != nil {
		return "", errors.Wrap(err, errGenerateToken.Error())
	}
	return tokenKey, nil
}

func (b *BusinessToken) Revoke(ctx context.Context, key string) error {
	err := b.dataLayer.RevokeToken(ctx, key)
	if err != nil {
		return errors.Wrap(err, errRevokeToken.Error())
	}
	return nil
}

func (b *BusinessToken) Validate(ctx context.Context, key string) error {
	logger := common.GetLogger(ctx)
	token, err := b.dataLayer.GetToken(ctx, key)
	if err != nil {
		return errors.Wrap(err, errGetToken.Error())
	}

	if token.Revoked {
		return errTokenRevoked
	}
	if token.Expired {
		logger.WithFields(logrus.Fields{
			"msg": fmt.Sprintf("Token: '%v', has already expired.", token.Key),
		}).Error("error_validate")
		return errTokenExpired
	}
	if time.Now().Unix() > token.ExpiresAt.Unix() {
		defer func() {
			err = b.dataLayer.UpdateTokenToExpired(ctx, token)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"err": err,
				}).Error("error_validate")
			}
		}()
		logger.WithFields(logrus.Fields{
			"err": errors.Wrap(err, errUpdateTokenToExpired.Error()),
		}).Error("error_validate")
		return errTokenDeterminedExpired
	}

	return nil
}

func NewBusinessToken(mysqlDataPersistence dataPersistence, tokenDaysValid int, randomCharMinLength int,
	randomCharMaxLength int) *BusinessToken {
	return &BusinessToken{
		dataLayer:           mysqlDataPersistence,
		tokenDaysValid:      tokenDaysValid,
		randomCharMinLength: randomCharMinLength,
		randomCharMaxLength: randomCharMaxLength,
	}
}
