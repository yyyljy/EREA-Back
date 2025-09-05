package models

import (
	"encoding/json"
	"time"
)

// Property represents a real estate property in the auction system
type Property struct {
	ID            string    `json:"id"`
	Title         string    `json:"title" binding:"required"`
	Location      string    `json:"location" binding:"required"`
	Description   string    `json:"description"`
	Type          string    `json:"type" binding:"required"` // Apartment, Officetel, Commercial, Villa
	Area          float64   `json:"area" binding:"required,min=0"`
	StartingPrice int64     `json:"starting_price" binding:"required,min=0"`
	CurrentPrice  int64     `json:"current_price"`
	ImageURL      string    `json:"image_url"`
	Features      []string  `json:"features"`
	Status        string    `json:"status"` // Active, Closed, Pending
	EndDate       time.Time `json:"end_date" binding:"required"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	OwnerID       string    `json:"owner_id"`
}

// ToJSON converts Property struct to JSON string
func (p *Property) ToJSON() (string, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON converts JSON string to Property struct
func (p *Property) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), p)
}

// CreatePropertyRequest represents a request to create a new property
type CreatePropertyRequest struct {
	Title         string    `json:"title" binding:"required"`
	Location      string    `json:"location" binding:"required"`
	Description   string    `json:"description"`
	Type          string    `json:"type" binding:"required"`
	Area          float64   `json:"area" binding:"required,min=0"`
	StartingPrice int64     `json:"starting_price" binding:"required,min=0"`
	ImageURL      string    `json:"image_url"`
	Features      []string  `json:"features"`
	EndDate       time.Time `json:"end_date" binding:"required"`
	OwnerID       string    `json:"owner_id" binding:"required"`
}

// UpdatePropertyRequest represents a request to update property
type UpdatePropertyRequest struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	Features    []string `json:"features,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// PropertyResponse represents API response for property operations
type PropertyResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
