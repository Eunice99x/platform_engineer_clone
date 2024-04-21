package api

import "github.com/gofiber/fiber/v2"

func GetRouter(app *fiber.App) {
	api := app.Group("/api")
	v0 := api.Group("/v0")

	v0token := v0.Group("/token")
	v0token.Get("/", token.GetToken)
}
