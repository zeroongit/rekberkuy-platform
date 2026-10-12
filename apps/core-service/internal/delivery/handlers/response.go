package handlers

import (
	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondError standardizes error responses across all HTTP handlers.
// Format: { "success": false, "error": { "code": "...", "message": "..." } }
func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// RespondJSON standardizes successful responses.
func RespondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, SuccessResponse{
		Success: true,
		Data:    data,
	})
}
