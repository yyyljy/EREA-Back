package handlers

import (
	"erea-api/config"
	"erea-api/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAllProperties retrieves all properties from Redis
func GetAllProperties(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get all property IDs
	keys, err := redis.Keys(ctx, "property:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to retrieve properties",
			Error:   err.Error(),
		})
		return
	}

	var properties []models.Property
	for _, key := range keys {
		propertyJSON, err := redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var property models.Property
		if err := property.FromJSON(propertyJSON); err != nil {
			continue
		}

		properties = append(properties, property)
	}

	c.JSON(http.StatusOK, models.PropertyResponse{
		Success: true,
		Message: "Properties retrieved successfully",
		Data:    properties,
	})
}

// GetProperty retrieves a specific property by ID
func GetProperty(c *gin.Context) {
	propertyID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	propertyJSON, err := redis.Get(ctx, "property:"+propertyID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.PropertyResponse{
			Success: false,
			Message: "Property not found",
			Error:   err.Error(),
		})
		return
	}

	var property models.Property
	if err := property.FromJSON(propertyJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to parse property data",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PropertyResponse{
		Success: true,
		Message: "Property retrieved successfully",
		Data:    property,
	})
}

// CreateProperty creates a new property
func CreateProperty(c *gin.Context) {
	var req models.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.PropertyResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Create new property
	property := models.Property{
		ID:            uuid.New().String(),
		Title:         req.Title,
		Location:      req.Location,
		Description:   req.Description,
		Type:          req.Type,
		Area:          req.Area,
		StartingPrice: req.StartingPrice,
		CurrentPrice:  req.StartingPrice,
		ImageURL:      req.ImageURL,
		Features:      req.Features,
		Status:        "Active",
		EndDate:       req.EndDate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		OwnerID:       req.OwnerID,
	}

	// Save to Redis
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	propertyJSON, err := property.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to encode property data",
			Error:   err.Error(),
		})
		return
	}

	if err := redis.Set(ctx, "property:"+property.ID, propertyJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to save property",
			Error:   err.Error(),
		})
		return
	}

	// Add to property list
	redis.SAdd(ctx, "properties", property.ID)

	c.JSON(http.StatusCreated, models.PropertyResponse{
		Success: true,
		Message: "Property created successfully",
		Data:    property,
	})
}

// UpdateProperty updates an existing property
func UpdateProperty(c *gin.Context) {
	propertyID := c.Param("id")
	var req models.UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.PropertyResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get existing property
	propertyJSON, err := redis.Get(ctx, "property:"+propertyID).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.PropertyResponse{
			Success: false,
			Message: "Property not found",
			Error:   err.Error(),
		})
		return
	}

	var property models.Property
	if err := property.FromJSON(propertyJSON); err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to parse property data",
			Error:   err.Error(),
		})
		return
	}

	// Update fields
	if req.Title != "" {
		property.Title = req.Title
	}
	if req.Description != "" {
		property.Description = req.Description
	}
	if req.ImageURL != "" {
		property.ImageURL = req.ImageURL
	}
	if req.Features != nil {
		property.Features = req.Features
	}
	if req.Status != "" {
		property.Status = req.Status
	}
	property.UpdatedAt = time.Now()

	// Save updated property
	updatedJSON, err := property.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to encode updated property data",
			Error:   err.Error(),
		})
		return
	}

	if err := redis.Set(ctx, "property:"+propertyID, updatedJSON, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to update property",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PropertyResponse{
		Success: true,
		Message: "Property updated successfully",
		Data:    property,
	})
}

// DeleteProperty deletes a property
func DeleteProperty(c *gin.Context) {
	propertyID := c.Param("id")
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Check if property exists
	exists, err := redis.Exists(ctx, "property:"+propertyID).Result()
	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, models.PropertyResponse{
			Success: false,
			Message: "Property not found",
		})
		return
	}

	// Delete property
	if err := redis.Del(ctx, "property:"+propertyID).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to delete property",
			Error:   err.Error(),
		})
		return
	}

	// Remove from property list
	redis.SRem(ctx, "properties", propertyID)

	c.JSON(http.StatusOK, models.PropertyResponse{
		Success: true,
		Message: "Property deleted successfully",
	})
}

// GetPropertiesByStatus retrieves properties by status
func GetPropertiesByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		status = "Active"
	}

	redis := config.GetRedisClient()
	ctx := config.GetContext()

	keys, err := redis.Keys(ctx, "property:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PropertyResponse{
			Success: false,
			Message: "Failed to retrieve properties",
			Error:   err.Error(),
		})
		return
	}

	var properties []models.Property
	for _, key := range keys {
		propertyJSON, err := redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var property models.Property
		if err := property.FromJSON(propertyJSON); err != nil {
			continue
		}

		if property.Status == status {
			properties = append(properties, property)
		}
	}

	c.JSON(http.StatusOK, models.PropertyResponse{
		Success: true,
		Message: "Properties retrieved successfully",
		Data:    properties,
	})
}
