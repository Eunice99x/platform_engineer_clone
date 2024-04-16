package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {

	db, err := sql.Open("mysql", "root:eunice99x@tcp(127.0.0.1:3306)/platform_engineer_clone")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":3000")
}
