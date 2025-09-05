package handlers

import (
	"erea-api/config"
	"erea-api/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateDemoData creates demo data for testing
func CreateDemoData(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Demo users
	users := []models.User{
		{
			ID:        uuid.New().String(),
			Name:      "John Smith",
			Email:     "john.smith@erea.gov",
			Age:       35,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Name:      "Sarah Johnson",
			Email:     "sarah.johnson@erea.gov",
			Age:       42,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Name:      "Michael Davis",
			Email:     "michael.davis@erea.gov",
			Age:       28,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Save demo users
	for _, user := range users {
		userJSON, _ := user.ToJSON()
		redis.Set(ctx, "user:"+user.ID, userJSON, 0)
	}

	// Demo properties (matching EREA frontend data)
	properties := []models.Property{
		{
			ID:            uuid.New().String(),
			Title:         "Gangnam District Premium Officetel",
			Location:      "Sinsa-dong, Gangnam-gu, Seoul, South Korea",
			Description:   "Modern officetel in the heart of Gangnam district with excellent transportation access.",
			Type:          "Officetel",
			Area:          45.2,
			StartingPrice: 500000000,
			CurrentPrice:  650000000,
			ImageURL:      "/api/placeholder/300/200",
			Features:      []string{"Near Subway Station", "24/7 Security", "Parking Available", "Modern Facilities"},
			Status:        "Active",
			EndDate:       time.Now().Add(10 * 24 * time.Hour), // 10 days from now
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			OwnerID:       users[0].ID,
		},
		{
			ID:            uuid.New().String(),
			Title:         "Bundang New Town Apartment Complex",
			Location:      "Jeongja-dong, Bundang-gu, Seongnam-si, Gyeonggi-do",
			Description:   "Spacious apartment in prestigious Bundang new town development.",
			Type:          "Apartment",
			Area:          84.3,
			StartingPrice: 800000000,
			CurrentPrice:  920000000,
			ImageURL:      "/api/placeholder/300/200",
			Features:      []string{"School District", "Park View", "Underground Parking", "Elevator"},
			Status:        "Active",
			EndDate:       time.Now().Add(8 * 24 * time.Hour), // 8 days from now
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			OwnerID:       users[1].ID,
		},
		{
			ID:            uuid.New().String(),
			Title:         "Hongdae Commercial Property",
			Location:      "Hapjeong-dong, Mapo-gu, Seoul, South Korea",
			Description:   "Prime commercial space in vibrant Hongdae entertainment district.",
			Type:          "Commercial",
			Area:          32.1,
			StartingPrice: 300000000,
			CurrentPrice:  380000000,
			ImageURL:      "/api/placeholder/300/200",
			Features:      []string{"High Foot Traffic", "Corner Location", "Restaurant Permitted", "Night Business"},
			Status:        "Closed",
			EndDate:       time.Now().Add(-2 * 24 * time.Hour), // 2 days ago
			CreatedAt:     time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:     time.Now(),
			OwnerID:       users[2].ID,
		},
		{
			ID:            uuid.New().String(),
			Title:         "Jeju Island Resort Villa",
			Location:      "Seogwipo-si, Jeju-do, South Korea",
			Description:   "Luxury resort villa with ocean view on beautiful Jeju Island.",
			Type:          "Villa",
			Area:          150.5,
			StartingPrice: 1200000000,
			CurrentPrice:  1350000000,
			ImageURL:      "/api/placeholder/300/200",
			Features:      []string{"Ocean View", "Private Garden", "Resort Amenities", "Tourist Zone"},
			Status:        "Active",
			EndDate:       time.Now().Add(11 * 24 * time.Hour), // 11 days from now
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			OwnerID:       users[0].ID,
		},
		{
			ID:            uuid.New().String(),
			Title:         "Busan Haeundae Beachfront Condo",
			Location:      "Haeundae-gu, Busan, South Korea",
			Description:   "Modern condominium with direct beach access in Haeundae.",
			Type:          "Condominium",
			Area:          65.8,
			StartingPrice: 700000000,
			CurrentPrice:  780000000,
			ImageURL:      "/api/placeholder/300/200",
			Features:      []string{"Beach Access", "Ocean View", "Resort Facilities", "Investment Property"},
			Status:        "Pending",
			EndDate:       time.Now().Add(15 * 24 * time.Hour), // 15 days from now
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			OwnerID:       users[1].ID,
		},
	}

	// Save demo properties
	for _, property := range properties {
		propertyJSON, _ := property.ToJSON()
		redis.Set(ctx, "property:"+property.ID, propertyJSON, 0)
		redis.SAdd(ctx, "properties", property.ID)
	}

	// Create demo auctions for active properties
	for _, property := range properties {
		if property.Status == "Active" {
			auction := models.Auction{
				ID:             uuid.New().String(),
				PropertyID:     property.ID,
				Status:         "Active",
				StartTime:      property.CreatedAt,
				EndTime:        property.EndDate,
				MinIncrement:   10000000, // 10M KRW minimum increment
				ReservePrice:   property.StartingPrice,
				CurrentHighest: property.CurrentPrice,
				BidCount:       3, // Simulate some bids
				CreatedAt:      property.CreatedAt,
				UpdatedAt:      time.Now(),
			}

			auctionJSON, _ := auction.ToJSON()
			redis.Set(ctx, "auction:"+auction.ID, auctionJSON, 0)
			redis.Set(ctx, "property_auction:"+property.ID, auction.ID, 0)
			redis.SAdd(ctx, "active_auctions", auction.ID)
		}
	}

	// Create demo bids
	bidCount := 0
	for _, property := range properties {
		if property.Status == "Active" || property.Status == "Closed" {
			// Create multiple bids per property
			bidAmounts := []int64{
				property.StartingPrice + 10000000,
				property.StartingPrice + 30000000,
				property.CurrentPrice,
			}

			for j, amount := range bidAmounts {
				bid := models.Bid{
					ID:          uuid.New().String(),
					PropertyID:  property.ID,
					BidderID:    users[j%len(users)].ID,
					Amount:      amount,
					TxHash:      "0x" + uuid.New().String()[:32],
					Status:      "Confirmed",
					IsEncrypted: true,
					CreatedAt:   time.Now().Add(-time.Duration(3-j) * time.Hour),
					UpdatedAt:   time.Now(),
				}

				bidJSON, _ := bid.ToJSON()
				redis.Set(ctx, "bid:"+bid.ID, bidJSON, 0)
				redis.SAdd(ctx, "property_bids:"+property.ID, bid.ID)
				bidCount++
			}
		}
	}

	response := gin.H{
		"success": true,
		"message": "Demo data created successfully",
		"data": gin.H{
			"users_created":      len(users),
			"properties_created": len(properties),
			"auctions_created":   3, // Active properties count
			"bids_created":       bidCount,
		},
	}

	c.JSON(http.StatusCreated, response)
}

// ClearDemoData clears all demo data
func ClearDemoData(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get all keys to delete
	patterns := []string{"user:*", "property:*", "auction:*", "bid:*", "property_bids:*", "property_auction:*"}
	deletedCount := 0

	for _, pattern := range patterns {
		keys, err := redis.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		if len(keys) > 0 {
			deleted, _ := redis.Del(ctx, keys...).Result()
			deletedCount += int(deleted)
		}
	}

	// Clear sets
	redis.Del(ctx, "properties", "active_auctions", "closed_auctions")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Demo data cleared successfully",
		"data": gin.H{
			"deleted_keys": deletedCount,
		},
	})
}

// GetDemoStatus checks if demo data exists
func GetDemoStatus(c *gin.Context) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	userKeys, _ := redis.Keys(ctx, "user:*").Result()
	propertyKeys, _ := redis.Keys(ctx, "property:*").Result()
	auctionKeys, _ := redis.Keys(ctx, "auction:*").Result()
	bidKeys, _ := redis.Keys(ctx, "bid:*").Result()

	status := gin.H{
		"demo_data_exists": len(userKeys) > 0 || len(propertyKeys) > 0,
		"users_count":      len(userKeys),
		"properties_count": len(propertyKeys),
		"auctions_count":   len(auctionKeys),
		"bids_count":       len(bidKeys),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Demo status retrieved successfully",
		"data":    status,
	})
}
