package middlewares

import (
	"errors"
	"github.com/dembygenesis/platform_engineer_exam/api/helpers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"net/http"
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
