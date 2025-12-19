package helper

import (
	// "fmt"

	// "github.com/go-playground/validator/v10"
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

// func ValidationError(c *fiber.Ctx, errors interface{}) error {
// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 		"success": false,
// 		"message": "Request body tidak valid",
// 		"errors":  errors,
// 	})
// }

// // validator
// func FormatValidationErrors(err error) []ValidationErrorResponse {
// 	var errors []ValidationErrorResponse

// 	if validationErrors, ok := err.(validator.ValidationErrors); ok {
// 		for _, fieldErr := range validationErrors {
// 			errors = append(errors, ValidationErrorResponse{
// 				Field:   fieldErr.Field(),
// 				Message: fmt.Sprintf("Aturan validasi '%s' gagal", fieldErr.Tag()),
// 			})
// 		}
// 	}
// 	return errors
// }
