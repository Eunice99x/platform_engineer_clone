package middlewares

import (
	"fmt"
	"net/http"

	"github.com/dembygenesis/platform_engineer_exam/api/helpers"
	"github.com/dembygenesis/platform_engineer_exam/models"
	"github.com/dembygenesis/platform_engineer_exam/src/utils/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/sirupsen/logrus"
)

const (
	userKey     = "user"
	passKey     = "pass"
	UserMetaKey = "userMeta"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . authFunctions
type authFunctions interface {
	BasicAuth(user, pass string) (bool, *models.User, error)
}

type AuthRoutes struct {
	authData authFunctions
}

func NewAuthRoutes(authData authFunctions) *AuthRoutes {
	return &AuthRoutes{authData}
}

// ProtectedRoute guards a route using the "Basic Auth" protocol
func (a *AuthRoutes) ProtectedRoute() func(ctx *fiber.Ctx) error {

	return basicauth.New(basicauth.Config{
		Authorizer: func(user, pass string) bool {
			matched, _, err := a.authData.BasicAuth(user, pass)
			if err != nil {
				return false
			}
			if !matched {
				return false
			}
			return true
		},
		Unauthorized: func(ctx *fiber.Ctx) error {

			reqHeaders := ctx.GetReqHeaders()
			fmt.Println("reqHeaders reqHeaders reqHeaders", reqHeaders)

			logger := common.GetLogger(ctx.Context())
			logger.WithFields(logrus.Fields{
				"msg": "Unauthorized",
			}).Error("error_protected_route")
			return ctx.Status(http.StatusUnauthorized).JSON(helpers.WrapStrInErrMap("Unauthorized"))
		},
		ContextUsername: userKey,
		ContextPassword: passKey,
	})
}

func (a *AuthRoutes) AttachUserMeta(ctx *fiber.Ctx) error {
	logger := common.GetLogger(ctx.Context())
	user := ctx.Locals(userKey).(string)
	pass := ctx.Locals(passKey).(string)

	_, userMeta, err := a.authData.BasicAuth(user, pass)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("error_extract_authed_user_meta")
		return ctx.Status(http.StatusUnauthorized).JSON(helpers.WrapErrInErrMap(err))
	}
	ctx.Locals(UserMetaKey, &models.User{
		Id: userMeta.Id,
	})
	return ctx.Next()
}
