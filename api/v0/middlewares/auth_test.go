package middlewares

import (
	"github.com/dembygenesis/platform_engineer_exam/api/helpers"
	"github.com/dembygenesis/platform_engineer_exam/api/v0/middlewares/middlewaresfakes"
	"github.com/dembygenesis/platform_engineer_exam/models"
	"github.com/friendsofgo/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectedRoute_HappyPath(t *testing.T) {
	fakeAuthFunctions := middlewaresfakes.FakeAuthFunctions{}
	fakeAuthFunctions.BasicAuthReturns(true, &models.User{Id: 3}, nil)

	authRoutes := NewAuthRoutes(&fakeAuthFunctions)

	app := fiber.New()
	app.Get("/", authRoutes.ProtectedRoute())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", helpers.AuthorizationHeaderBasicAuth("admin", "123456"))

	resp, _ := app.Test(req, 1)
	t.Run("Test ProtectedRoute Returns True", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestProtectedRoute_FailPath_Error(t *testing.T) {
	fakeAuthFunctions := middlewaresfakes.FakeAuthFunctions{}
	fakeAuthFunctions.BasicAuthReturns(true, &models.User{Id: 3}, errors.New("mock error"))

	authRoutes := NewAuthRoutes(&fakeAuthFunctions)

	app := fiber.New()
	app.Get("/", authRoutes.ProtectedRoute())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", helpers.AuthorizationHeaderBasicAuth("admin", "123456"))

	resp, _ := app.Test(req, 1)
	t.Run("Test ProtectedRoute - Fail Error", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestProtectedRoute_FailPath_NoMatch(t *testing.T) {
	fakeAuthFunctions := middlewaresfakes.FakeAuthFunctions{}
	fakeAuthFunctions.BasicAuthReturns(false, &models.User{Id: 3}, nil)

	authRoutes := NewAuthRoutes(&fakeAuthFunctions)

	app := fiber.New()
	app.Get("/", authRoutes.ProtectedRoute())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", helpers.AuthorizationHeaderBasicAuth("admin", "123456"))

	resp, _ := app.Test(req, 1)
	t.Run("Test ProtectedRoute - No Match", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthRoutes_AttachUserMeta_HappyPath(t *testing.T) {
	fakeAuthFunctions := middlewaresfakes.FakeAuthFunctions{}
	fakeAuthFunctions.BasicAuthReturns(false, &models.User{Id: 3}, nil)

	authRoutes := NewAuthRoutes(&fakeAuthFunctions)

	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		ctx.Locals(userKey, "abc")
		ctx.Locals(passKey, "abc")
		return ctx.Next()
	}, authRoutes.AttachUserMeta)

	req := httptest.NewRequest("GET", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test AuthRoutes - Happy Path", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestAuthRoutes_AttachUserMeta_Fail_ErrorBasicAuth(t *testing.T) {
	fakeAuthFunctions := middlewaresfakes.FakeAuthFunctions{}
	fakeAuthFunctions.BasicAuthReturns(false, &models.User{Id: 3}, errors.New("mock error"))

	authRoutes := NewAuthRoutes(&fakeAuthFunctions)

	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		ctx.Locals(userKey, "abc")
		ctx.Locals(passKey, "abc")
		return ctx.Next()
	}, authRoutes.AttachUserMeta)

	req := httptest.NewRequest("GET", "/", nil)

	resp, _ := app.Test(req, 1)
	t.Run("Test AuthRoutes - Fail, Error Basic Auth", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
