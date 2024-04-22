package token

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"platform_engineer_clone/api/helpers"
	"platform_engineer_clone/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . bizFunctions
type bizFunctions interface {
	Validate(ctx context.Context, key string) error
	GetAll(ctx context.Context) ([]models.Token, error)
	Revoke(ctx context.Context, key string) error
	Generate(ctx context.Context, user *models.User) (string, error)
}

// These error codes are used in tests
var (
	errMockGetAll   = errors.New("error, mock GetAll")
	errMockRevoke   = errors.New("error, mock Revoke")
	errMockValidate = errors.New("error, mock Validate")
	errMockGenerate = errors.New("error, mock Generate")
)

type APIToken struct {
	bizLayer bizFunctions
}

func NewAPIToken(bizLayer bizFunctions) *APIToken {
	return &APIToken{bizLayer}
}

// ValidateToken
// @Id ValidateToken
// @Summary Validate
// @Description Validates a string token passed
// @Tags Token
// @Param token path string true "token"
// @Accept application/json
// @Produce application/json
// @Success 200 {boolean} boolean
// @Failure 400 {object} models.AuthFailBadRequest
// @Failure 500 {object} models.AuthFailInternalServerError
// @Router /v0/token/{token}/validate [get]
func (t *APIToken) ValidateToken(ctx *fiber.Ctx) error {
	token := ctx.Params("token")

	err := t.bizLayer.Validate(ctx.Context(), token)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(helpers.WrapErrInErrMap(err))
	}
	return ctx.Status(http.StatusOK).JSON(true)
}

// GetAll
// @Id GetAll
// @Summary Fetch all
// @Description Fetches all tokens added by the admin user
// @Tags Token
// @Accept application/json
// @Produce application/json
// @Success 200 {object} []models.Token
// @Failure 400 {object} models.AuthFailBadRequest
// @Failure 500 {object} models.AuthFailInternalServerError
// @Security BasicAuth
// @Router /v0/token [get]
func (t *APIToken) GetAll(ctx *fiber.Ctx) error {
	tokens, err := t.bizLayer.GetAll(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(helpers.WrapErrInErrMap(err))
	}
	return ctx.Status(http.StatusOK).JSON(tokens)
}

// Revoke
// @Id Revoke
// @Summary Revoke
// @Description Revokes a token's access
// @Tags Token
// @Accept application/json
// @Produce application/json
// @Param token path string true "token"
// @Success 200 {boolean} boolean
// @Failure 400 {object} models.AuthFailBadRequest
// @Failure 500 {object} models.AuthFailInternalServerError
// @Security BasicAuth
// @Router /v0/token/{token}/revoke [delete]
func (t *APIToken) Revoke(ctx *fiber.Ctx) error {
	token := ctx.Params("token")
	err := t.bizLayer.Revoke(ctx.Context(), token)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(helpers.WrapErrInErrMap(err))
	}
	return ctx.Status(http.StatusOK).SendString("Revoked token access!")
}

// GetToken Creates a new invite token
// @Id GetToken
// @Summary Create
// @Description Creates a new invite token
// @Tags Token
// @Accept application/json
// @Produce application/json
// @Success 201 {string} string
// @Failure 400 {object} models.AuthFailBadRequest
// @Failure 500 {object} models.AuthFailInternalServerError
// @Security BasicAuth
// @Router /v0/token [post]
func (t *APIToken) GetToken(ctx *fiber.Ctx) error {
	userMeta, ok := ctx.Locals("userMeta").(*models.User)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(helpers.WrapStrInErrMap("userMeta conversion fails"))
	}

	generatedToken, err := t.bizLayer.Generate(ctx.Context(), userMeta)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(helpers.WrapErrInErrMap(err))
	}
	return ctx.Status(http.StatusCreated).JSON(generatedToken)
}
