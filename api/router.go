package api

import (
	"github.com/gofiber/fiber/v2"
	"platform_engineer_clone/api/v0/middlewares"
	"platform_engineer_clone/dependency_injection/dic"
)

func GetRouter(app *fiber.App, ctn *dic.Container) {
	api := app.Group("/api")
	v0 := api.Group("/v0")

	apiToken := ctn.GetApiToken()
	authMiddlewares := ctn.GetApiMiddlewares()

	v0token := v0.Group("/token")
	v0token.Get("/", authMiddlewares.ProtectedRoute(), authMiddlewares.AttachUserMeta, apiToken.GetAll)
	v0token.Post("/", authMiddlewares.ProtectedRoute(), authMiddlewares.AttachUserMeta, apiToken.GetToken)
	v0token.Get("/:token/validate", middlewares.Throttle(), apiToken.ValidateToken)
	v0token.Delete("/:token/revoke", authMiddlewares.ProtectedRoute(), authMiddlewares.AttachUserMeta, apiToken.Revoke)
}
