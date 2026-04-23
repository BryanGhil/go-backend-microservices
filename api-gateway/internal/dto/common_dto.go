package dto

// ==========================================
// UNIVERSAL RESPONSES (Used by all handlers)
// ==========================================

// ErrorResponse represents your utils.ErrorResponse
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error description goes here"`
}

// IDData is the generic { "id": 123 } payload
type IDData struct {
	ID int64 `json:"id" example:"123"`
}

// IDResponse is used for POST, PUT, and DELETE successes
type IDResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation successful"`
	Data    IDData `json:"data"`
}