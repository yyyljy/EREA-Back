package handlers

import (
	"erea-api/config"
	"erea-api/models"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PlaceBid creates a new bid for a property
func PlaceBid(c *gin.Context) {
	var req models.CreateBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BidResponse{
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
		c.JSON(http.StatusNotFound, models.BidResponse{
			Success: false,
			Message: "Property not found",
			Error:   err.Error(),
		})
		return
	}

	var property models.Property
	if err := property.FromJSON(propertyJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to parse property data",
			Error:   err.Error(),
		})
		return
	}

	// Check if auction is still active
	if property.Status != "Active" || time.Now().After(property.EndDate) {
		c.JSON(http.StatusBadRequest, models.BidResponse{
			Success: false,
			Message: "Auction is not active",
		})
		return
	}

	// Check if bid is higher than current price
	if req.Amount <= property.CurrentPrice {
		c.JSON(http.StatusBadRequest, models.BidResponse{
			Success: false,
			Message: "Bid must be higher than current price",
		})
		return
	}

	// Create new bid
	bid := models.Bid{
		ID:            uuid.New().String(),
		PropertyID:    req.PropertyID,
		BidderID:      req.BidderID,
		Amount:        req.Amount,
		Status:        "Pending",
		IsEncrypted:   req.IsEncrypted,
		EncryptedData: req.EncryptedData,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Simulate blockchain transaction hash
	bid.TxHash = fmt.Sprintf("0x%s", uuid.New().String()[:32])

	// Save bid to Redis
	bidJSON, err := bid.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to encode bid data",
			Error:   err.Error(),
		})
		return
	}

	bidKey := fmt.Sprintf("bid:%s", bid.ID)
	if err := redis.Set(ctx, bidKey, bidJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to save bid",
			Error:   err.Error(),
		})
		return
	}

	// Add bid to property's bid list
	propertyBidsKey := fmt.Sprintf("property_bids:%s", req.PropertyID)
	redis.SAdd(ctx, propertyBidsKey, bid.ID)

	// Update property's current price
	property.CurrentPrice = req.Amount
	property.UpdatedAt = time.Now()

	updatedPropertyJSON, err := property.ToJSON()
	if err == nil {
		redis.Set(ctx, "property:"+req.PropertyID, updatedPropertyJSON, 0)
	}

	// Update bid status to confirmed
	bid.Status = "Confirmed"
	bid.UpdatedAt = time.Now()
	confirmedBidJSON, _ := bid.ToJSON()
	redis.Set(ctx, bidKey, confirmedBidJSON, 0)

	// Broadcast bid update via WebSocket
	BroadcastBidUpdate(req.PropertyID, bid)

	c.JSON(http.StatusCreated, models.BidResponse{
		Success: true,
		Message: "Bid placed successfully",
		Data:    bid,
	})
}

// GetBidHistory retrieves bid history for a property
func GetBidHistory(c *gin.Context) {
	propertyID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Check if property exists
	if exists, err := redis.Exists(ctx, "property:"+propertyID).Result(); err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, models.BidResponse{
			Success: false,
			Message: "Property not found",
		})
		return
	}

	// Get all bid IDs for this property
	propertyBidsKey := fmt.Sprintf("property_bids:%s", propertyID)
	bidIDs, err := redis.SMembers(ctx, propertyBidsKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to retrieve bid history",
			Error:   err.Error(),
		})
		return
	}

	var bids []models.Bid
	var highestBid int64
	var latestBidder string

	for _, bidID := range bidIDs {
		bidJSON, err := redis.Get(ctx, "bid:"+bidID).Result()
		if err != nil {
			continue
		}

		var bid models.Bid
		if err := bid.FromJSON(bidJSON); err != nil {
			continue
		}

		bids = append(bids, bid)

		if bid.Amount > highestBid {
			highestBid = bid.Amount
			latestBidder = bid.BidderID
		}
	}

	// Sort bids by creation time (newest first)
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].CreatedAt.After(bids[j].CreatedAt)
	})

	bidHistory := models.BidHistory{
		PropertyID:   propertyID,
		BidCount:     len(bids),
		HighestBid:   highestBid,
		LatestBidder: latestBidder,
		Bids:         bids,
	}

	c.JSON(http.StatusOK, models.BidResponse{
		Success: true,
		Message: "Bid history retrieved successfully",
		Data:    bidHistory,
	})
}

// GetUserBids retrieves all bids for a specific user
func GetUserBids(c *gin.Context) {
	userID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get all bid keys
	bidKeys, err := redis.Keys(ctx, "bid:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to retrieve user bids",
			Error:   err.Error(),
		})
		return
	}

	var userBids []models.Bid

	for _, key := range bidKeys {
		bidJSON, err := redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var bid models.Bid
		if err := bid.FromJSON(bidJSON); err != nil {
			continue
		}

		if bid.BidderID == userID {
			userBids = append(userBids, bid)
		}
	}

	// Sort bids by creation time (newest first)
	sort.Slice(userBids, func(i, j int) bool {
		return userBids[i].CreatedAt.After(userBids[j].CreatedAt)
	})

	c.JSON(http.StatusOK, models.BidResponse{
		Success: true,
		Message: "User bids retrieved successfully",
		Data:    userBids,
	})
}

// GetBid retrieves a specific bid by ID
func GetBid(c *gin.Context) {
	bidID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	bidJSON, err := redis.Get(ctx, "bid:"+bidID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.BidResponse{
			Success: false,
			Message: "Bid not found",
			Error:   err.Error(),
		})
		return
	}

	var bid models.Bid
	if err := bid.FromJSON(bidJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to parse bid data",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BidResponse{
		Success: true,
		Message: "Bid retrieved successfully",
		Data:    bid,
	})
}

// UpdateBidStatus updates the status of a bid (for blockchain confirmations)
func UpdateBidStatus(c *gin.Context) {
	bidID := c.Param("id")
	status := c.PostForm("status")
	txHash := c.PostForm("tx_hash")

	if status == "" {
		c.JSON(http.StatusBadRequest, models.BidResponse{
			Success: false,
			Message: "Status is required",
		})
		return
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	bidJSON, err := redis.Get(ctx, "bid:"+bidID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.BidResponse{
			Success: false,
			Message: "Bid not found",
			Error:   err.Error(),
		})
		return
	}

	var bid models.Bid
	if err := bid.FromJSON(bidJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to parse bid data",
			Error:   err.Error(),
		})
		return
	}

	bid.Status = status
	if txHash != "" {
		bid.TxHash = txHash
	}
	bid.UpdatedAt = time.Now()

	updatedBidJSON, err := bid.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to encode updated bid data",
			Error:   err.Error(),
		})
		return
	}

	if err := redis.Set(ctx, "bid:"+bidID, updatedBidJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to update bid status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BidResponse{
		Success: true,
		Message: "Bid status updated successfully",
		Data:    bid,
	})
}

// GetTopBids retrieves top bids across all properties
func GetTopBids(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	bidKeys, err := redis.Keys(ctx, "bid:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BidResponse{
			Success: false,
			Message: "Failed to retrieve bids",
			Error:   err.Error(),
		})
		return
	}

	var bids []models.Bid

	for _, key := range bidKeys {
		bidJSON, err := redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var bid models.Bid
		if err := bid.FromJSON(bidJSON); err != nil {
			continue
		}

		if bid.Status == "Confirmed" {
			bids = append(bids, bid)
		}
	}

	// Sort by amount (highest first)
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Amount > bids[j].Amount
	})

	// Limit results
	if len(bids) > limit {
		bids = bids[:limit]
	}

	c.JSON(http.StatusOK, models.BidResponse{
		Success: true,
		Message: "Top bids retrieved successfully",
		Data:    bids,
	})
}
