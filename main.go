package main

import (
	"context"
	"database/sql"
	"log"
	"platform_engineer_clone/routes"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize database connection
	db, err := sql.Open("mysql", "root:eunice99x@tcp(127.0.0.1:3306)/platform_engineer_clone")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create Fiber app
	app := fiber.New()

	// Define token creation route
	app.Post("/tokens", func(c *fiber.Ctx) error {
		return routes.CreateToken(context.Background(), db, c)
	})

	// Start Fiber server
	log.Fatal(app.Listen(":3000"))
}
