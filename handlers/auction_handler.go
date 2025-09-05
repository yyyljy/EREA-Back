package handlers

import (
	"erea-api/config"
	"erea-api/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateAuction creates a new auction for a property
func CreateAuction(c *gin.Context) {
	var req models.CreateAuctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.AuctionResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Check if property exists
	propertyJSON, err := redis.Get(ctx, "property:"+req.PropertyID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.AuctionResponse{
			Success: false,
			Message: "Property not found",
			Error:   err.Error(),
		})
		return
	}

	var property models.Property
	if err := property.FromJSON(propertyJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to parse property data",
			Error:   err.Error(),
		})
		return
	}

	// Create new auction
	auction := models.Auction{
		ID:             uuid.New().String(),
		PropertyID:     req.PropertyID,
		Status:         "Active",
		StartTime:      time.Now(),
		EndTime:        req.EndTime,
		MinIncrement:   req.MinIncrement,
		ReservePrice:   req.ReservePrice,
		CurrentHighest: property.StartingPrice,
		BidCount:       0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save auction to Redis
	auctionJSON, err := auction.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to encode auction data",
			Error:   err.Error(),
		})
		return
	}

	auctionKey := fmt.Sprintf("auction:%s", auction.ID)
	if err := redis.Set(ctx, auctionKey, auctionJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to save auction",
			Error:   err.Error(),
		})
		return
	}

	// Link auction to property
	redis.Set(ctx, "property_auction:"+req.PropertyID, auction.ID, 0)

	// Add to active auctions list
	redis.SAdd(ctx, "active_auctions", auction.ID)

	c.JSON(http.StatusCreated, models.AuctionResponse{
		Success: true,
		Message: "Auction created successfully",
		Data:    auction,
	})
}

// GetAuction retrieves a specific auction by ID
func GetAuction(c *gin.Context) {
	auctionID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	auctionJSON, err := redis.Get(ctx, "auction:"+auctionID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.AuctionResponse{
			Success: false,
			Message: "Auction not found",
			Error:   err.Error(),
		})
		return
	}

	var auction models.Auction
	if err := auction.FromJSON(auctionJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to parse auction data",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.AuctionResponse{
		Success: true,
		Message: "Auction retrieved successfully",
		Data:    auction,
	})
}

// GetActiveAuctions retrieves all active auctions
func GetActiveAuctions(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	auctionIDs, err := redis.SMembers(ctx, "active_auctions").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to retrieve active auctions",
			Error:   err.Error(),
		})
		return
	}

	var auctions []models.Auction
	for _, auctionID := range auctionIDs {
		auctionJSON, err := redis.Get(ctx, "auction:"+auctionID).Result()
		if err != nil {
			continue
		}

		var auction models.Auction
		if err := auction.FromJSON(auctionJSON); err != nil {
			continue
		}

		// Check if auction is still active
		if time.Now().After(auction.EndTime) && auction.Status == "Active" {
			// Auto-close expired auction
			auction.Status = "Closed"
			updatedJSON, _ := auction.ToJSON()
			redis.Set(ctx, "auction:"+auctionID, updatedJSON, 0)
			redis.SRem(ctx, "active_auctions", auctionID)
			redis.SAdd(ctx, "closed_auctions", auctionID)
		}

		if auction.Status == "Active" {
			auctions = append(auctions, auction)
		}
	}

	c.JSON(http.StatusOK, models.AuctionResponse{
		Success: true,
		Message: "Active auctions retrieved successfully",
		Data:    auctions,
	})
}

// CloseAuction manually closes an auction
func CloseAuction(c *gin.Context) {
	auctionID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	auctionJSON, err := redis.Get(ctx, "auction:"+auctionID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.AuctionResponse{
			Success: false,
			Message: "Auction not found",
			Error:   err.Error(),
		})
		return
	}

	var auction models.Auction
	if err := auction.FromJSON(auctionJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to parse auction data",
			Error:   err.Error(),
		})
		return
	}

	if auction.Status != "Active" {
		c.JSON(http.StatusBadRequest, models.AuctionResponse{
			Success: false,
			Message: "Auction is not active",
		})
		return
	}

	// Find winning bid
	propertyBidsKey := fmt.Sprintf("property_bids:%s", auction.PropertyID)
	bidIDs, err := redis.SMembers(ctx, propertyBidsKey).Result()
	if err == nil && len(bidIDs) > 0 {
		var highestBid int64
		var winnerID string

		for _, bidID := range bidIDs {
			bidJSON, err := redis.Get(ctx, "bid:"+bidID).Result()
			if err != nil {
				continue
			}

			var bid models.Bid
			if err := bid.FromJSON(bidJSON); err != nil {
				continue
			}

			if bid.Amount > highestBid && bid.Status == "Confirmed" {
				highestBid = bid.Amount
				winnerID = bid.BidderID
			}
		}

		auction.WinnerID = winnerID
		auction.WinningBid = highestBid
	}

	// Close auction
	auction.Status = "Closed"
	auction.UpdatedAt = time.Now()

	// Save updated auction
	updatedJSON, err := auction.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to encode updated auction data",
			Error:   err.Error(),
		})
		return
	}

	if err := redis.Set(ctx, "auction:"+auctionID, updatedJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to close auction",
			Error:   err.Error(),
		})
		return
	}

	// Move from active to closed
	redis.SRem(ctx, "active_auctions", auctionID)
	redis.SAdd(ctx, "closed_auctions", auctionID)

	// Update property status
	propertyJSON, err := redis.Get(ctx, "property:"+auction.PropertyID).Result()
	if err == nil {
		var property models.Property
		if property.FromJSON(propertyJSON) == nil {
			property.Status = "Closed"
			property.UpdatedAt = time.Now()
			updatedPropertyJSON, _ := property.ToJSON()
			redis.Set(ctx, "property:"+auction.PropertyID, updatedPropertyJSON, 0)
		}
	}

	// Broadcast auction update via WebSocket
	BroadcastAuctionUpdate(auction)

	c.JSON(http.StatusOK, models.AuctionResponse{
		Success: true,
		Message: "Auction closed successfully",
		Data:    auction,
	})
}

// GetAuctionStats retrieves auction statistics
func GetAuctionStats(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get all auction keys
	auctionKeys, err := redis.Keys(ctx, "auction:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to retrieve auction statistics",
			Error:   err.Error(),
		})
		return
	}

	var totalAuctions, activeAuctions, closedAuctions int
	var totalVolume int64
	var successfulAuctions int

	for _, key := range auctionKeys {
		auctionJSON, err := redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var auction models.Auction
		if err := auction.FromJSON(auctionJSON); err != nil {
			continue
		}

		totalAuctions++

		switch auction.Status {
		case "Active":
			activeAuctions++
		case "Closed":
			closedAuctions++
			if auction.WinningBid > 0 {
				totalVolume += auction.WinningBid
				successfulAuctions++
			}
		}
	}

	var averagePrice float64
	if successfulAuctions > 0 {
		averagePrice = float64(totalVolume) / float64(successfulAuctions)
	}

	var successRate float64
	if closedAuctions > 0 {
		successRate = float64(successfulAuctions) / float64(closedAuctions) * 100
	}

	stats := models.AuctionStats{
		TotalAuctions:  totalAuctions,
		ActiveAuctions: activeAuctions,
		ClosedAuctions: closedAuctions,
		TotalVolume:    totalVolume,
		AveragePrice:   averagePrice,
		SuccessRate:    successRate,
	}

	c.JSON(http.StatusOK, models.AuctionResponse{
		Success: true,
		Message: "Auction statistics retrieved successfully",
		Data:    stats,
	})
}

// GetPropertyAuction retrieves auction for a specific property
func GetPropertyAuction(c *gin.Context) {
	propertyID := c.Param("property_id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get auction ID for this property
	auctionID, err := redis.Get(ctx, "property_auction:"+propertyID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.AuctionResponse{
			Success: false,
			Message: "No auction found for this property",
			Error:   err.Error(),
		})
		return
	}

	auctionJSON, err := redis.Get(ctx, "auction:"+auctionID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.AuctionResponse{
			Success: false,
			Message: "Auction not found",
			Error:   err.Error(),
		})
		return
	}

	var auction models.Auction
	if err := auction.FromJSON(auctionJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.AuctionResponse{
			Success: false,
			Message: "Failed to parse auction data",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.AuctionResponse{
		Success: true,
		Message: "Property auction retrieved successfully",
		Data:    auction,
	})
}
