package helper

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestSuccessResponse(t *testing.T) {
	app := fiber.New()

	tests := []struct {
		name       string
		statusCode int
		message    string
		data       interface{}
	}{
		{
			name:       "Success with data",
			statusCode: fiber.StatusOK,
			message:    "Operation successful",
			data:       map[string]string{"key": "value"},
		},
		{
			name:       "Success with nil data",
			statusCode: fiber.StatusOK,
			message:    "Success",
			data:       nil,
		},
		{
			name:       "Created response",
			statusCode: fiber.StatusCreated,
			message:    "Resource created",
			data:       map[string]int{"id": 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(c)

			err := SuccessResponse(c, tt.statusCode, tt.message, tt.data)
			if err != nil {
				t.Errorf("SuccessResponse() error = %v", err)
			}

			if c.Response().StatusCode() != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, c.Response().StatusCode())
			}

			var response map[string]interface{}
			if err := json.Unmarshal(c.Response().Body(), &response); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if status, ok := response["status"].(string); !ok || status != "success" {
				t.Errorf("Expected status=success, got %v", response["status"])
			}

			if msg, ok := response["message"].(string); !ok || msg != tt.message {
				t.Errorf("Expected message=%s, got %v", tt.message, response["message"])
			}

			if tt.data != nil {
				if _, ok := response["data"]; !ok {
					t.Error("Expected data field in response")
				}
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	app := fiber.New()

	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "Bad request error",
			statusCode: fiber.StatusBadRequest,
			message:    "Invalid input",
		},
		{
			name:       "Unauthorized error",
			statusCode: fiber.StatusUnauthorized,
			message:    "Authentication required",
		},
		{
			name:       "Not found error",
			statusCode: fiber.StatusNotFound,
			message:    "Resource not found",
		},
		{
			name:       "Internal server error",
			statusCode: fiber.StatusInternalServerError,
			message:    "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(c)

			err := ErrorResponse(c, tt.statusCode, tt.message)
			if err != nil {
				t.Errorf("ErrorResponse() error = %v", err)
			}

			if c.Response().StatusCode() != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, c.Response().StatusCode())
			}

			var response map[string]interface{}
			if err := json.Unmarshal(c.Response().Body(), &response); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if success, ok := response["success"].(string); !ok || success != "error" {
				t.Errorf("Expected success=error, got %v", response["success"])
			}

			if msg, ok := response["message"].(string); !ok || msg != tt.message {
				t.Errorf("Expected message=%s, got %v", tt.message, response["message"])
			}
		})
	}
}

func TestResponseContentType(t *testing.T) {
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	SuccessResponse(c, fiber.StatusOK, "Test", nil)

	contentType := string(c.Response().Header.ContentType())
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
