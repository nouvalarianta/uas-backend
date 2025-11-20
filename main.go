package main

import "github.com/gofiber/fiber/v2"

func main() {
	app := fiber.New()

	// Route GET /
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, Fiber!",
		})
	})

	// Jalankan server pada port 3000
	app.Listen(":3000")
}
