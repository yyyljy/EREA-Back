package models

import (
	"encoding/json"
	"time"
)

// Bid represents a bid in the auction system
type Bid struct {
	ID           string    `json:"id"`
	PropertyID   string    `json:"property_id" binding:"required"`
	BidderID     string    `json:"bidder_id" binding:"required"`
	Amount       int64     `json:"amount" binding:"required,min=0"`
	TxHash       string    `json:"tx_hash"` // Blockchain transaction hash
	Status       string    `json:"status"`  // Pending, Confirmed, Failed
	IsEncrypted  bool      `json:"is_encrypted"`
	EncryptedData string   `json:"encrypted_data,omitempty"` // EERC encrypted bid data
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToJSON converts Bid struct to JSON string
func (b *Bid) ToJSON() (string, error) {
	jsonData, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON converts JSON string to Bid struct
func (b *Bid) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), b)
}

// CreateBidRequest represents a request to place a bid
type CreateBidRequest struct {
	PropertyID    string `json:"property_id" binding:"required"`
	BidderID      string `json:"bidder_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,min=0"`
	IsEncrypted   bool   `json:"is_encrypted"`
	EncryptedData string `json:"encrypted_data,omitempty"`
}

// BidResponse represents API response for bid operations
type BidResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BidHistory represents bid history for a property
type BidHistory struct {
	PropertyID   string `json:"property_id"`
	BidCount     int    `json:"bid_count"`
	HighestBid   int64  `json:"highest_bid"`
	LatestBidder string `json:"latest_bidder"`
	Bids         []Bid  `json:"bids"`
}
