package token

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"platform_engineer_clone/api/helpers"
)

func GetToken(ctx *fiber.Ctx) error {
	ctn, err := helpers.GetContainer(ctx)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(err.Error())
	}
	biz, err := ctn.SafeGetBusinessToken()
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(err.Error())
	}

	generatedToken, err := biz.Generate()
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(err.Error())
	}

	return ctx.JSON(generatedToken)
}
