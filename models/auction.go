package models

import (
	"encoding/json"
	"time"
)

// Auction represents an auction session
type Auction struct {
	ID             string    `json:"id"`
	PropertyID     string    `json:"property_id" binding:"required"`
	Status         string    `json:"status"` // Active, Closed, Cancelled
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	MinIncrement   int64     `json:"min_increment"`   // Minimum bid increment
	ReservePrice   int64     `json:"reserve_price"`   // Reserve price
	CurrentHighest int64     `json:"current_highest"` // Current highest bid
	BidCount       int       `json:"bid_count"`
	WinnerID       string    `json:"winner_id,omitempty"`
	WinningBid     int64     `json:"winning_bid,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ToJSON converts Auction struct to JSON string
func (a *Auction) ToJSON() (string, error) {
	jsonData, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON converts JSON string to Auction struct
func (a *Auction) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), a)
}

// CreateAuctionRequest represents a request to create an auction
type CreateAuctionRequest struct {
	PropertyID   string    `json:"property_id" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	MinIncrement int64     `json:"min_increment"`
	ReservePrice int64     `json:"reserve_price"`
}

// AuctionStats represents auction statistics
type AuctionStats struct {
	TotalAuctions   int     `json:"total_auctions"`
	ActiveAuctions  int     `json:"active_auctions"`
	ClosedAuctions  int     `json:"closed_auctions"`
	TotalVolume     int64   `json:"total_volume"`
	AveragePrice    float64 `json:"average_price"`
	SuccessRate     float64 `json:"success_rate"`
}

// AuctionResponse represents API response for auction operations
type AuctionResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
