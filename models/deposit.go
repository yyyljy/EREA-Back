package models

import (
	"encoding/json"
	"time"
)

// Deposit represents a deposit in the auction system
type Deposit struct {
	ID           string    `json:"id"`
	PropertyID   string    `json:"property_id" binding:"required"`
	UserID       string    `json:"user_id" binding:"required"`
	Amount       int64     `json:"amount" binding:"required,min=0"`
	TokenType    string    `json:"token_type"`   // wKRW, EERC20
	TxHash       string    `json:"tx_hash"`      // Blockchain transaction hash
	Status       string    `json:"status"`       // Pending, Confirmed, Failed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToJSON converts Deposit struct to JSON string
func (d *Deposit) ToJSON() (string, error) {
	jsonData, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON converts JSON string to Deposit struct
func (d *Deposit) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), d)
}

// CreateDepositRequest represents a request to create a deposit
type CreateDepositRequest struct {
	PropertyID string `json:"property_id" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
	Amount     int64  `json:"amount" binding:"required,min=0"`
	TokenType  string `json:"token_type" binding:"required"`
}

// DepositResponse represents API response for deposit operations
type DepositResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// DepositListResponse represents response for multiple deposits
type DepositListResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Data     []Deposit `json:"data,omitempty"`
	Total    int       `json:"total"`
	Error    string    `json:"error,omitempty"`
}
