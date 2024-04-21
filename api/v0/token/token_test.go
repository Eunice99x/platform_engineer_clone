package token

import (
	"github.com/dembygenesis/platform_engineer_exam/api/v0/token/tokenfakes"
	"github.com/dembygenesis/platform_engineer_exam/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestGetToken_StatusCreated(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.GenerateReturns("12345", nil)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Post("/", func(ctx *fiber.Ctx) error {
		ctx.Locals("userMeta", &models.User{Id: 1})
		return ctx.Next()
	}, apiToken.GetToken)

	req := httptest.NewRequest("POST", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test Validate - StatusCreated", func(t *testing.T) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestGetToken_InternalServerError_GenerateFails(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.GenerateReturns("", errMockGenerate)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Post("/", func(ctx *fiber.Ctx) error {
		ctx.Locals("userMeta", &models.User{Id: 1})
		return ctx.Next()
	}, apiToken.GetToken)

	req := httptest.NewRequest("POST", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test Validate - Internal Server Error", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetToken_InternalServerError_UserMetaFails(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Get("/", apiToken.GetToken)

	req := httptest.NewRequest("GET", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test Validate - Internal Server Error", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

var m sync.RWMutex
var wg sync.WaitGroup

func TestValidate_StatusOk(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.ValidateReturns(nil)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Get("/:token/validate", apiToken.ValidateToken)

	req := httptest.NewRequest("GET", "/mock_token_value/validate", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test Validate - StatusOk", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestValidate_InternalServerError(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.ValidateReturns(errMockValidate)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Get("/:token/validate", apiToken.ValidateToken)

	req := httptest.NewRequest("GET", "/mock_token_value/validate", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test Validate - Internal Server Error", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestRevoke_StatusOk(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.RevokeReturns(nil)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Delete("/:token/revoke", apiToken.Revoke)

	req := httptest.NewRequest("DELETE", "/mock_token_value/revoke", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test GetAll - StatusOk", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestRevoke_InternalServerError(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.RevokeReturns(errMockRevoke)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Delete("/:token/revoke", apiToken.Revoke)

	req := httptest.NewRequest("DELETE", "/mock_token_value/revoke", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test GetAll - Internal Server Error", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetAll_StatusOk(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.GetAllReturns(nil, nil)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Get("/", apiToken.GetAll)

	req := httptest.NewRequest("GET", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test GetAll - Ok", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetAll_InternalServerError(t *testing.T) {
	fakeBizFunctions := &tokenfakes.FakeBizFunctions{}
	fakeBizFunctions.GetAllReturns(nil, errMockGetAll)

	apiToken := NewAPIToken(fakeBizFunctions)

	app := fiber.New()
	app.Get("/", apiToken.GetAll)

	req := httptest.NewRequest("GET", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test GetAll - Internal Server Error", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
