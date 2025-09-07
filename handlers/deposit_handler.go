package handlers

import (
	"erea-api/config"
	"erea-api/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateDeposit creates a new deposit for a property
func CreateDeposit(c *gin.Context) {
	var req models.CreateDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.DepositResponse{
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
		c.JSON(http.StatusNotFound, models.DepositResponse{
			Success: false,
			Message: "Property not found",
			Error:   err.Error(),
		})
		return
	}

	var property models.Property
	if err := property.FromJSON(propertyJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to parse property data",
			Error:   err.Error(),
		})
		return
	}

	// Calculate minimum deposit (10% of starting price)
	minimumDeposit := property.StartingPrice / 10
	
	// Validate minimum deposit amount
	if req.Amount < minimumDeposit {
		c.JSON(http.StatusBadRequest, models.DepositResponse{
			Success: false,
			Message: fmt.Sprintf("Deposit amount must be at least %d (10%% of starting price)", minimumDeposit),
			Error:   "Insufficient deposit amount",
		})
		return
	}

	// Check if user already has a confirmed deposit for this property
	existingDeposits, err := redis.Keys(ctx, fmt.Sprintf("deposit:*:%s:%s", req.PropertyID, req.UserID)).Result()
	if err == nil {
		for _, key := range existingDeposits {
			depositJSON, err := redis.Get(ctx, key).Result()
			if err == nil {
				var existingDeposit models.Deposit
				if err := existingDeposit.FromJSON(depositJSON); err == nil {
					if existingDeposit.Status == "Confirmed" {
						c.JSON(http.StatusConflict, models.DepositResponse{
							Success: false,
							Message: "You already have a confirmed deposit for this property",
							Error:   "Duplicate deposit",
						})
						return
					}
				}
			}
		}
	}

	// Create new deposit
	deposit := models.Deposit{
		ID:         uuid.New().String(),
		PropertyID: req.PropertyID,
		UserID:     req.UserID,
		Amount:     req.Amount,
		TokenType:  req.TokenType,
		TxHash:     generateTxHash(),
		Status:     "Confirmed", // For demo purposes, set directly to confirmed
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to Redis
	depositJSON, err := deposit.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to serialize deposit data",
			Error:   err.Error(),
		})
		return
	}

	key := fmt.Sprintf("deposit:%s:%s:%s", deposit.ID, deposit.PropertyID, deposit.UserID)
	if err := redis.Set(ctx, key, depositJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to save deposit",
			Error:   err.Error(),
		})
		return
	}

	// Add to user's deposit list
	userDepositKey := fmt.Sprintf("user_deposits:%s", req.UserID)
	redis.SAdd(ctx, userDepositKey, deposit.ID)

	// Add to property's deposit list
	propertyDepositKey := fmt.Sprintf("property_deposits:%s", req.PropertyID)
	redis.SAdd(ctx, propertyDepositKey, deposit.ID)

	c.JSON(http.StatusCreated, models.DepositResponse{
		Success: true,
		Message: "Deposit created successfully",
		Data:    deposit,
	})
}

// GetAllDeposits retrieves all deposits
func GetAllDeposits(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	keys, err := redis.Keys(ctx, "deposit:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositListResponse{
			Success: false,
			Message: "Failed to retrieve deposits",
			Error:   err.Error(),
		})
		return
	}

	var deposits []models.Deposit
	for _, key := range keys {
		depositJSON, err := redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var deposit models.Deposit
		if err := deposit.FromJSON(depositJSON); err != nil {
			continue
		}

		deposits = append(deposits, deposit)
	}

	c.JSON(http.StatusOK, models.DepositListResponse{
		Success: true,
		Message: "Deposits retrieved successfully",
		Data:    deposits,
		Total:   len(deposits),
	})
}

// GetUserDeposits retrieves deposits for a specific user
func GetUserDeposits(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.DepositListResponse{
			Success: false,
			Message: "User ID is required",
			Error:   "Missing user ID",
		})
		return
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get user's deposit IDs
	userDepositKey := fmt.Sprintf("user_deposits:%s", userID)
	depositIDs, err := redis.SMembers(ctx, userDepositKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositListResponse{
			Success: false,
			Message: "Failed to retrieve user deposits",
			Error:   err.Error(),
		})
		return
	}

	var deposits []models.Deposit
	for _, depositID := range depositIDs {
		// Find the deposit key by pattern matching
		keys, err := redis.Keys(ctx, fmt.Sprintf("deposit:%s:*", depositID)).Result()
		if err != nil || len(keys) == 0 {
			continue
		}

		depositJSON, err := redis.Get(ctx, keys[0]).Result()
		if err != nil {
			continue
		}

		var deposit models.Deposit
		if err := deposit.FromJSON(depositJSON); err != nil {
			continue
		}

		deposits = append(deposits, deposit)
	}

	c.JSON(http.StatusOK, models.DepositListResponse{
		Success: true,
		Message: "User deposits retrieved successfully",
		Data:    deposits,
		Total:   len(deposits),
	})
}

// GetDeposit retrieves a specific deposit by ID
func GetDeposit(c *gin.Context) {
	depositID := c.Param("id")
	if depositID == "" {
		c.JSON(http.StatusBadRequest, models.DepositResponse{
			Success: false,
			Message: "Deposit ID is required",
			Error:   "Missing deposit ID",
		})
		return
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Find the deposit key by pattern matching
	keys, err := redis.Keys(ctx, fmt.Sprintf("deposit:%s:*", depositID)).Result()
	if err != nil || len(keys) == 0 {
		c.JSON(http.StatusNotFound, models.DepositResponse{
			Success: false,
			Message: "Deposit not found",
			Error:   "Deposit ID not found",
		})
		return
	}

	depositJSON, err := redis.Get(ctx, keys[0]).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.DepositResponse{
			Success: false,
			Message: "Deposit not found",
			Error:   err.Error(),
		})
		return
	}

	var deposit models.Deposit
	if err := deposit.FromJSON(depositJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to parse deposit data",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DepositResponse{
		Success: true,
		Message: "Deposit retrieved successfully",
		Data:    deposit,
	})
}

// UpdateDepositStatus updates the status of a deposit
func UpdateDepositStatus(c *gin.Context) {
	depositID := c.Param("id")
	if depositID == "" {
		c.JSON(http.StatusBadRequest, models.DepositResponse{
			Success: false,
			Message: "Deposit ID is required",
			Error:   "Missing deposit ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		TxHash string `json:"tx_hash,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.DepositResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Find the deposit key by pattern matching
	keys, err := redis.Keys(ctx, fmt.Sprintf("deposit:%s:*", depositID)).Result()
	if err != nil || len(keys) == 0 {
		c.JSON(http.StatusNotFound, models.DepositResponse{
			Success: false,
			Message: "Deposit not found",
			Error:   "Deposit ID not found",
		})
		return
	}

	depositJSON, err := redis.Get(ctx, keys[0]).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.DepositResponse{
			Success: false,
			Message: "Deposit not found",
			Error:   err.Error(),
		})
		return
	}

	var deposit models.Deposit
	if err := deposit.FromJSON(depositJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to parse deposit data",
			Error:   err.Error(),
		})
		return
	}

	// Update deposit
	deposit.Status = req.Status
	if req.TxHash != "" {
		deposit.TxHash = req.TxHash
	}
	deposit.UpdatedAt = time.Now()

	// Save updated deposit
	updatedJSON, err := deposit.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to serialize updated deposit",
			Error:   err.Error(),
		})
		return
	}

	if err := redis.Set(ctx, keys[0], updatedJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.DepositResponse{
			Success: false,
			Message: "Failed to update deposit",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DepositResponse{
		Success: true,
		Message: "Deposit status updated successfully",
		Data:    deposit,
	})
}

// generateTxHash generates a mock transaction hash for demo purposes
func generateTxHash() string {
	uuid1 := strings.ReplaceAll(uuid.New().String(), "-", "")
	uuid2 := strings.ReplaceAll(uuid.New().String(), "-", "")
	// Take 32 from first UUID and 8 from second to make 40 characters
	return fmt.Sprintf("0x%s%s", uuid1, uuid2[:8])
}
