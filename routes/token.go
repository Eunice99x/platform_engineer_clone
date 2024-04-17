package routes

import (
	"context"
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/rand"
	"platform_engineer_clone/models"
	"time"
)

func CreateToken(ctx context.Context, db boil.ContextExecutor, c *fiber.Ctx) error {
	var token models.Token
	if err := c.BodyParser(&token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Set the key
	token.Key = null.StringFrom(randomKey())

	// Insert the token into the database
	if err := token.Insert(ctx, db, boil.Infer()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Return a success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Token created successfully", "token": token})
}

func randomKey() string {
	str := "abcdefghijklmnopqrstuvwxyz"
	num := "0123456789"

	var arr []rune
	arr = append(arr, []rune(str)...)
	arr = append(arr, []rune(num)...)

	rand.Seed(time.Now().UnixNano())

	var token []rune
	for i := 0; i < 7; i++ {
		randomNum := rand.Intn(len(arr))
		tokenLetter := arr[randomNum]
		token = append(token, tokenLetter)
	}

	return string(token)
}

func GetAllTokens(ctx context.Context, db boil.ContextExecutor, c *fiber.Ctx) error {
	// Fetch all tokens from the database
	tokens, err := models.Tokens().All(ctx, db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tokens)
}

func GetToken(ctx context.Context, db boil.ContextExecutor, c *fiber.Ctx) error {
	tokenID := c.Params("id")

	token, err := models.Tokens(qm.Where("id = ?", tokenID)).One(ctx, db)
	if err != nil {
		// If the token is not found or any other error occurs, return an error response
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Token not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(token)
}

func DeleteToken(ctx context.Context, db boil.ContextExecutor, c *fiber.Ctx) error {
	tokenID := c.Params("id")

	token, err := models.Tokens(qm.Where("id = ?", tokenID)).One(ctx, db)
	if err != nil {
		// If the token is not found or any other error occurs, return an error response
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Token not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := token.Delete(ctx, db); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Token deleted successfully"})
}
