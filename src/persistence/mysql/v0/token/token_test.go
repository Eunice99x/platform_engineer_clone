package token

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dembygenesis/platform_engineer_exam/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"time"
)

func configureMockGenerateFailFetchToken(mock sqlmock.Sqlmock) {
	sqlToken := "SELECT `token`.*"
	mock.ExpectQuery(regexp.QuoteMeta(sqlToken)).WillReturnError(errFetchToken)
}

func configureMockGenerateFailInsertToken(mock sqlmock.Sqlmock, randomString string, createdAt time.Time) {
	var mockIdReturned int64 = 1
	sqlInsert := "INSERT INTO `token` (`key`,`created_at`,`created_by`,`expires_at`) VALUES (?,?,?,?)"
	mock.ExpectExec(regexp.QuoteMeta(sqlInsert)).WithArgs(
		randomString,
		createdAt,
		3,
		createdAt.Add(7*time.Hour*24),
	).WillReturnResult(sqlmock.NewResult(mockIdReturned, 1)).WillReturnError(errInsertNewToken)
}

func configureMockGeneratePassInsertToken(mock sqlmock.Sqlmock, randomString string, createdBy int, createdAt time.Time) {
	var mockIdReturned int64 = 1
	sqlInsert := "INSERT INTO `token` (`key`,`created_at`,`created_by`,`expires_at`) VALUES (?,?,?,?)"
	mock.ExpectExec(regexp.QuoteMeta(sqlInsert)).WithArgs(
		randomString,
		createdAt,
		3,
		createdAt.Add(7*time.Hour*24),
	).WillReturnResult(sqlmock.NewResult(mockIdReturned, 1))

	sqlPostSelectAfterSQLBoilerInsert := "SELECT `id`,`revoked`,`expired` FROM `token` WHERE `id`=?"
	rows := sqlmock.NewRows([]string{"id", "revoked", "expired"})
	rows.AddRow(mockIdReturned, false, false)
	mock.ExpectQuery(regexp.QuoteMeta(sqlPostSelectAfterSQLBoilerInsert)).WithArgs(
		mockIdReturned,
	).WillReturnRows(rows)
}

func configureMockGeneratePassFetchToken(mock sqlmock.Sqlmock, randomString string) {
	sqlToken := "SELECT `token`.* FROM `token` WHERE (`token`.`key` = ?);"
	rows := sqlmock.NewRows([]string{
		"id",
		"key",
		"created_at",
		"revoked",
		"expired",
		"created_by",
		"expires_at",
	})
	mock.ExpectQuery(regexp.QuoteMeta(sqlToken)).WithArgs(randomString).WillReturnRows(rows)
}

func TestPersistenceToken_Generate_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	randomString := generateRandomCharacters(12)
	createdAt := time.Now()
	createdById := 3

	sqlToken := "SELECT `token`.* FROM `token` WHERE (`token`.`key` = ?);"
	rows := sqlmock.NewRows([]string{
		"id",
	})
	rows.AddRow("123")
	mock.ExpectQuery(regexp.QuoteMeta(sqlToken)).WithArgs(randomString).WillReturnRows(rows)

	sqlToken = "SELECT `token`.* FROM `token` WHERE (`token`.`key` = ?);"
	rows = sqlmock.NewRows([]string{
		"id",
	})
	mock.ExpectQuery(regexp.QuoteMeta(sqlToken)).WithArgs(randomString).WillReturnRows(rows)

	configureMockGeneratePassInsertToken(mock, randomString, createdById, createdAt)

	persistenceToken := PersistenceToken{db: db, mockRandomString: randomString, mockCreatedTime: createdAt}
	_, err = persistenceToken.Generate(context.Background(), createdById, 6, 12)
	t.Run("Test Generate Happy Path", func(t *testing.T) {
		require.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPersistenceToken_Generate_FailCheckUniqueToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	configureMockGenerateFailFetchToken(mock)

	createdById := 3

	persistenceToken := PersistenceToken{db: db}
	_, err = persistenceToken.Generate(context.Background(), createdById, 6, 12)
	t.Run("Test Generate Fail Check Unique Token", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errFetchToken.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestPersistenceToken_Generate_FailInsertNewToken(t *testing.T) {
	randomString := generateRandomCharacters(12)
	createdAt := time.Now()
	createdById := 3

	db, mock, err := sqlmock.New()
	configureMockGeneratePassFetchToken(mock, randomString)
	configureMockGenerateFailInsertToken(mock, randomString, createdAt)

	persistenceToken := PersistenceToken{db: db, mockCreatedTime: createdAt, mockRandomString: randomString}
	t.Run("Test Generate Fail Insert New Token", func(t *testing.T) {
		_, err = persistenceToken.Generate(context.Background(), createdById, 6, 12)
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errInsertNewToken.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func configureMockGetAllFetchTokensSuccess(mock sqlmock.Sqlmock) {
	sqlFetchTokens := "SELECT token.id AS id, token.key AS `key`, token.created_at AS created_at, token.revoked AS revoked, token.expired AS expired, token.expires_at AS expires_at, u.name AS created_by FROM `token` INNER JOIN user u ON u.id = token.created_by;"

	headers := []string{
		"id",
		"`key`",
		"created_at",
		"revoked",
		"expired",
		"expires_at",
		"created_by",
	}
	data := []driver.Value{
		1,
		"abc",
		time.Now(),
		true,
		true,
		time.Now(),
		"Demby",
	}
	rows := sqlmock.NewRows(headers).AddRow(data...)
	mock.ExpectQuery(regexp.QuoteMeta(sqlFetchTokens)).WillReturnRows(rows)
}

func TestPersistenceToken_GetAll_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	configureMockGetAllFetchTokensSuccess(mock)

	persistenceToken := PersistenceToken{db: db}
	res, err := persistenceToken.GetAll(context.Background())

	t.Run("Test GetAll Happy Path", func(t *testing.T) {
		require.NoError(t, err)

		resLength := len(res)
		require.Equal(t, true, resLength > 0)
	})
}

func configureMockGetAllFetchTokensFail(mock sqlmock.Sqlmock) {
	sqlFetchTokens := "SELECT token.id AS id, token.key AS `key`, token.created_at AS created_at, token.revoked AS revoked, token.expired AS expired, token.expires_at AS expires_at, u.name AS created_by FROM `token` INNER JOIN user u ON u.id = token.created_by;"

	mock.ExpectQuery(regexp.QuoteMeta(sqlFetchTokens)).WillReturnError(errFetchToken)
}

func TestPersistenceToken_GetAll_FailPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	configureMockGetAllFetchTokensFail(mock)

	persistenceToken := PersistenceToken{db: db}
	_, err = persistenceToken.GetAll(context.Background())

	t.Run("Test GetAll Fail Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errFetchTokens.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestNewPersistenceToken(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)

	persistenceToken := NewPersistenceToken(db)
	t.Run("Test NewPersistenceToken", func(t *testing.T) {
		require.NotNil(t, persistenceToken)
	})
}

func configureMockUpdateTokenToExpiredFailFindToken(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("select (.+) FROM `token.*").WillReturnError(errFetchToken)
}

func TestPersistenceToken_UpdateTokenToExpired_FailFindToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	configureMockUpdateTokenToExpiredFailFindToken(mock)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.UpdateTokenToExpired(context.Background(), &models.Token{Key: "123456"})
	t.Run("Test GetAll Fail Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errFetchToken.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func configureMockUpdateTokenToExpiredPassFindToken(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{"id"})
	rows.AddRow(1)
	mock.ExpectQuery("select (.+) .*").WillReturnRows(rows)
}

func configureMockUpdateTokenToExpiredFailUpdate(mock sqlmock.Sqlmock) {
	mock.ExpectExec("UPDATE `token.*").WillReturnError(errUpdateTokenToExpired)
}

func TestPersistenceToken_UpdateTokenToExpired_FailUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	configureMockUpdateTokenToExpiredPassFindToken(mock)
	configureMockUpdateTokenToExpiredFailUpdate(mock)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.UpdateTokenToExpired(context.Background(), &models.Token{Key: "123456"})
	t.Run("Test GetAll Fail Update", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errUpdateTokenToExpired.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func configureMockUpdateTokenToExpiredPassUpdate(mock sqlmock.Sqlmock) {
	rows := sqlmock.NewRows([]string{"A"})
	rows.AddRow("1")
	mock.ExpectExec("UPDATE `token.*").WillReturnResult(sqlmock.NewResult(1, 1))
}

func TestPersistenceToken_UpdateTokenToExpired_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	configureMockUpdateTokenToExpiredPassFindToken(mock)
	configureMockUpdateTokenToExpiredPassUpdate(mock)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.UpdateTokenToExpired(context.Background(), &models.Token{Key: "123456"})
	t.Run("Test UpdateTokenToExpired - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestPersistenceToken_RevokeToken_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"})
	rows.AddRow("3")
	mock.ExpectQuery("SELECT `token.*").WillReturnRows(rows)
	mock.ExpectExec("UPDATE `token` .*").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("select * from `token`").WillReturnRows(rows)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.RevokeToken(context.Background(), "123456")
	t.Run("Test RevokeToken - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestPersistenceToken_RevokeToken_Fail_FindToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectQuery("select (.*) from `token.*").WillReturnError(errFetchToken)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.RevokeToken(context.Background(), "123456")
	t.Run("Test RevokeToken - Happy Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errFetchToken.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestPersistenceToken_RevokeToken_Fail_FindTokenNoResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectQuery("SELECT `token.*").WillReturnError(sql.ErrNoRows)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.RevokeToken(context.Background(), "123456")
	fmt.Println("err", err)
	t.Run("Test RevokeToken - Fail Find Token No Results", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errTokenNotFound.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestPersistenceToken_RevokeToken_Fail_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"})
	rows.AddRow(0)
	mock.ExpectQuery("SELECT `token.*").WillReturnRows(rows)
	mock.ExpectExec("UPDATE `token` .*").WillReturnError(errUpdateTokenToRevoked)

	persistenceToken := PersistenceToken{db: db}
	err = persistenceToken.RevokeToken(context.Background(), "123456")
	t.Run("Test RevokeToken - Fail Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errUpdateTokenToRevoked.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestPersistenceToken_GetToken_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"})
	rows.AddRow(0)
	mock.ExpectQuery("SELECT `token.*").WillReturnRows(rows)

	persistenceToken := PersistenceToken{db: db}
	_, err = persistenceToken.GetToken(context.Background(), "12345")
	t.Run("Test GetToken - Happy Path", func(t *testing.T) {
		require.NoError(t, err)
	})
}

func TestPersistenceToken_GetToken_FailPath_FetchTokenByKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"})
	rows.AddRow(0)
	mock.ExpectQuery("SELECT `token.*").WillReturnError(errFetchTokenByKey)

	persistenceToken := PersistenceToken{db: db}
	_, err = persistenceToken.GetToken(context.Background(), "12345")
	t.Run("Test GetToken - Fail Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errFetchTokenByKey.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}

func TestPersistenceToken_GetToken_FailPath_NoResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT `token.*").WillReturnRows(rows)

	persistenceToken := PersistenceToken{db: db}
	_, err = persistenceToken.GetToken(context.Background(), "12345")
	t.Run("Test GetToken - Fail Path", func(t *testing.T) {
		require.Error(t, err)

		errMsg := err.Error()
		wantErrMsg := errFetchTokenByKeyNoResult.Error()
		assert.Containsf(t, errMsg, wantErrMsg, "expected error containing %q, got %s", wantErrMsg, err)
	})
}
