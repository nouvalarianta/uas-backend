package helper

import (
	"github.com/gofiber/fiber/v2"
)

type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// response helper
func SuccessResponse(c *fiber.Ctx, code int, message string, data interface{}) error {
	return c.Status(code).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

func ErrorResponse(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"success": "error",
		"message": message,
	})
}
