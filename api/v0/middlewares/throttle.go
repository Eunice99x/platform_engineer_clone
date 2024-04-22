package middlewares

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"net/http"
	"platform_engineer_clone/api/helpers"
	"time"
)

var (
	ErrThrottleLimitExceeded = errors.New("please limit your requests to 5 per 5 seconds")
)

func Throttle() func(ctx *fiber.Ctx) error {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 5 * time.Second,
		LimitReached: func(ctx *fiber.Ctx) error {
			return ctx.Status(http.StatusForbidden).JSON(helpers.WrapErrInErrMap(ErrThrottleLimitExceeded))
		},
	})
}
