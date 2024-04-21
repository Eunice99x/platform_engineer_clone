package token

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"platform_engineer_clone/business/v0/token/tokenfakes"
	"platform_engineer_clone/models"
	"testing"
	"time"
)

func TestBusinessToken_Generate_HappyPath(t *testing.T) {
	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GenerateReturns("1234", nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	_, err := businessToken.Generate(context.Background(), &models.User{Id: 3})
	t.Run("Test Generate - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestBusinessToken_Generate_FailPath(t *testing.T) {
	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GenerateReturns("", errGenerateToken)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	_, err := businessToken.Generate(context.Background(), &models.User{Id: 3})
	t.Run("Test Generate - Happy Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errGenerateToken.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestBusinessToken_GetAll_HappyPath(t *testing.T) {
	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetAllReturns([]models.Token{
		{
			Id: 1,
		},
	}, nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	_, err := businessToken.GetAll(context.Background())
	t.Run("Test Get - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestBusinessToken_GetAll_FailPath(t *testing.T) {
	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetAllReturns([]models.Token{
		{
			Id: 1,
		},
	}, errGetTokens)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	_, err := businessToken.GetAll(context.Background())
	t.Run("Test Get - Happy Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errGetTokens.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestBusinessToken_Revoke_Fail(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.RevokeTokenReturns(errTokenRevoked)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Revoke(context.Background(), tokenKey)
	t.Run("Test Revoke - Fail Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errTokenRevoked.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestBusinessToken_Revoke_HappyPath(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.RevokeTokenReturns(nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Revoke(context.Background(), tokenKey)
	t.Run("Test Revoke - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestBusinessToken_Validate_HappyPath(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetTokenReturns(&models.Token{
		Id:        1,
		Key:       tokenKey,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		Revoked:   false,
		Expired:   false,
		CreatedBy: "Demby",
	}, nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Validate(context.Background(), tokenKey)
	t.Run("Test Validate - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestBusinessToken_Validate_FailPath_UpdateTokenToExpired(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetTokenReturns(&models.Token{
		Id:        1,
		Key:       tokenKey,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		Revoked:   false,
		Expired:   false,
		CreatedBy: "Demby",
	}, errUpdateTokenToExpired)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Validate(context.Background(), tokenKey)
	t.Run("Test Validate - Update Token To Expired", func(t *testing.T) {
		defer func() {
			require.Error(t, err)
		}()
	})
}

func TestBusinessToken_Validate_FailPath_GetToken(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetTokenReturns(&models.Token{
		Id:        1,
		Key:       tokenKey,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		Revoked:   false,
		Expired:   false,
		CreatedBy: "Demby",
	}, nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Validate(context.Background(), tokenKey)
	t.Run("Test Validate - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestBusinessToken_Validate_FailPath_Revoked(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetTokenReturns(&models.Token{
		Id:        1,
		Key:       tokenKey,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		Revoked:   true,
		Expired:   false,
		CreatedBy: "Demby",
	}, nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Validate(context.Background(), tokenKey)
	t.Run("Test Validate - Fail Path Revoked", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errTokenRevoked.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestBusinessToken_Validate_FailPath_Expired(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetTokenReturns(&models.Token{
		Id:        1,
		Key:       tokenKey,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		Revoked:   false,
		Expired:   true,
		CreatedBy: "Demby",
	}, nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Validate(context.Background(), tokenKey)
	t.Run("Test Validate - Fail Path Expired", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errTokenExpired.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestBusinessToken_Validate_FailPath_DeterminedExpired(t *testing.T) {
	tokenKey := "123456"

	fakeDataPersistence := tokenfakes.FakeDataPersistence{}
	fakeDataPersistence.GetTokenReturns(&models.Token{
		Id:        1,
		Key:       tokenKey,
		ExpiresAt: time.Now().AddDate(0, 0, -1),
		Revoked:   false,
		Expired:   false,
		CreatedBy: "Demby",
	}, nil)

	businessToken := NewBusinessToken(&fakeDataPersistence, 7, 6, 12)
	err := businessToken.Validate(context.Background(), tokenKey)
	fmt.Println("err err err", err)
	t.Run("Test Validate - Fail Path Determined Expired", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errTokenDeterminedExpired.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}
