package utils

import "github.com/gin-gonic/gin"

// APIResponse is the standard envelope for all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	// omitempty means if Data is nil, it won't even show up in the JSON!
	Data    interface{} `json:"data,omitempty"` 
}

// SuccessResponse formats a successful API call
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse formats a failed API call
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
	c.Abort() // Stops any further handlers from running
}