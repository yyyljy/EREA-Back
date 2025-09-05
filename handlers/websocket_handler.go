package handlers

import (
	"encoding/json"
	"erea-api/config"
	"erea-api/models"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		// In production, you should check the origin
		return true
	},
}

// Client represents a websocket client
type Client struct {
	conn       *websocket.Conn
	send       chan []byte
	propertyID string
	userID     string
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Property-specific client groups
	propertyClients map[string]map[*Client]bool

	mutex sync.RWMutex
}

// Global hub instance
var hub = &Hub{
	broadcast:       make(chan []byte),
	register:        make(chan *Client),
	unregister:      make(chan *Client),
	clients:         make(map[*Client]bool),
	propertyClients: make(map[string]map[*Client]bool),
}

// WebSocketMessage represents a websocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// BidUpdate represents a bid update message
type BidUpdate struct {
	PropertyID    string `json:"property_id"`
	NewBid        int64  `json:"new_bid"`
	BidderID      string `json:"bidder_id"`
	BidCount      int    `json:"bid_count"`
	TimeRemaining string `json:"time_remaining"`
}

// AuctionUpdate represents an auction update message
type AuctionUpdate struct {
	PropertyID string `json:"property_id"`
	Status     string `json:"status"`
	WinnerID   string `json:"winner_id,omitempty"`
	WinningBid int64  `json:"winning_bid,omitempty"`
}

func init() {
	go hub.run()
}

// run handles hub operations
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			
			// Add to property-specific group if specified
			if client.propertyID != "" {
				if h.propertyClients[client.propertyID] == nil {
					h.propertyClients[client.propertyID] = make(map[*Client]bool)
				}
				h.propertyClients[client.propertyID][client] = true
			}
			h.mutex.Unlock()
			
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				
				// Remove from property-specific group
				if client.propertyID != "" {
					if clients, exists := h.propertyClients[client.propertyID]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.propertyClients, client.propertyID)
						}
					}
				}
			}
			h.mutex.Unlock()
			
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// broadcastToProperty sends a message to all clients watching a specific property
func (h *Hub) broadcastToProperty(propertyID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if clients, exists := h.propertyClients[propertyID]; exists {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(clients, client)
			}
		}
	}
}

// HandleWebSocket handles websocket connections
func HandleWebSocket(c *gin.Context) {
	propertyID := c.Query("property_id")
	userID := c.Query("user_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}

	client := &Client{
		conn:       conn,
		send:       make(chan []byte, 256),
		propertyID: propertyID,
		userID:     userID,
	}

	hub.register <- client

	// Start goroutines for this client
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocket error: %v", err)
			}
			break
		}
		// Handle incoming messages if needed
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Websocket write error: %v", err)
				return
			}
		}
	}
}

// BroadcastBidUpdate broadcasts a bid update to relevant clients
func BroadcastBidUpdate(propertyID string, bid models.Bid) {
	redis := config.GetRedisClient()
	ctx := config.GetContext()

	// Get bid count for this property
	propertyBidsKey := "property_bids:" + propertyID
	bidCount, _ := redis.SCard(ctx, propertyBidsKey).Result()

	// Get property to calculate time remaining
	propertyJSON, err := redis.Get(ctx, "property:"+propertyID).Result()
	timeRemaining := "Unknown"
	if err == nil {
		var property models.Property
		if property.FromJSON(propertyJSON) == nil {
			remaining := property.EndDate.Sub(property.CreatedAt)
			timeRemaining = remaining.String()
		}
	}

	update := BidUpdate{
		PropertyID:    propertyID,
		NewBid:        bid.Amount,
		BidderID:      bid.BidderID,
		BidCount:      int(bidCount),
		TimeRemaining: timeRemaining,
	}

	message := WebSocketMessage{
		Type:    "bid_update",
		Data:    update,
		Message: "New bid placed",
	}

	// Convert to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal bid update: %v", err)
		return
	}

	// Broadcast to property-specific clients
	hub.broadcastToProperty(propertyID, messageJSON)
}

// BroadcastAuctionUpdate broadcasts an auction status update
func BroadcastAuctionUpdate(auction models.Auction) {
	update := AuctionUpdate{
		PropertyID: auction.PropertyID,
		Status:     auction.Status,
		WinnerID:   auction.WinnerID,
		WinningBid: auction.WinningBid,
	}

	message := WebSocketMessage{
		Type:    "auction_update",
		Data:    update,
		Message: "Auction status updated",
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal auction update: %v", err)
		return
	}

	// Broadcast to all clients
	hub.broadcast <- messageJSON
}

// BroadcastPropertyUpdate broadcasts a property update
func BroadcastPropertyUpdate(property models.Property) {
	message := WebSocketMessage{
		Type:    "property_update",
		Data:    property,
		Message: "Property updated",
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal property update: %v", err)
		return
	}

	// Broadcast to property-specific clients
	hub.broadcastToProperty(property.ID, messageJSON)
}

// GetConnectedClients returns the number of connected clients
func GetConnectedClients(c *gin.Context) {
	propertyID := c.Query("property_id")

	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	var count int
	if propertyID != "" {
		if clients, exists := hub.propertyClients[propertyID]; exists {
			count = len(clients)
		}
	} else {
		count = len(hub.clients)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"connected_clients": count,
		"property_id":      propertyID,
	})
}
