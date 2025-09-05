# EREA API - Real Estate Auction Backend

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/redis-%23DC382D.svg?style=for-the-badge&logo=redis&logoColor=white)
![Gin](https://img.shields.io/badge/gin-%2300ADD8.svg?style=for-the-badge&logo=gin&logoColor=white)

A high-performance backend API for the EREA (Encrypted Real Estate Auction) platform, built with Go, Gin, and Redis. This API provides comprehensive real estate auction functionality with real-time updates via WebSocket connections.

## 🏗️ Architecture

- **Backend Framework**: Gin (Go)
- **Database**: Redis (In-memory data store)
- **Real-time Communication**: WebSocket
- **Authentication**: Token-based (ready for implementation)
- **API Design**: RESTful with real-time extensions

## 🚀 Features

### 🏠 Property Management
- Create, read, update, delete properties
- Property status management (Active, Closed, Pending)
- Property search and filtering
- Support for various property types (Apartment, Officetel, Commercial, Villa)

### 🔨 Auction System
- Create and manage auctions
- Automatic auction closure
- Bid validation and processing
- Auction statistics and analytics

### 💰 Bidding Engine
- Secure bid placement
- Real-time bid updates
- Bid history tracking
- Encrypted bid support (EERC compatible)

### 📊 Analytics & Statistics
- Dashboard statistics
- Real-time metrics
- User performance tracking
- Property analytics

### 🔄 Real-time Features
- WebSocket connections for live updates
- Real-time bid notifications
- Auction status updates
- Connected client monitoring

## 📡 API Endpoints

### Health Check
```
GET /health
```

### Users
```
POST   /api/v1/users              # Create user
GET    /api/v1/users              # Get all users
GET    /api/v1/users/:id          # Get specific user
PUT    /api/v1/users/:id          # Update user
DELETE /api/v1/users/:id          # Delete user
GET    /api/v1/users/:user_id/bids    # Get user bids
GET    /api/v1/users/:user_id/stats   # Get user statistics
```

### Properties
```
POST   /api/v1/properties                      # Create property
GET    /api/v1/properties                      # Get all properties
GET    /api/v1/properties/status?status=Active # Get properties by status
GET    /api/v1/properties/:id                  # Get specific property
PUT    /api/v1/properties/:id                  # Update property
DELETE /api/v1/properties/:id                  # Delete property
GET    /api/v1/properties/:property_id/auction # Get property auction
GET    /api/v1/properties/:property_id/bids    # Get property bid history
GET    /api/v1/properties/:property_id/stats   # Get property statistics
```

### Bids
```
POST   /api/v1/bids              # Place bid
GET    /api/v1/bids              # Get top bids
GET    /api/v1/bids/:id          # Get specific bid
PUT    /api/v1/bids/:id/status   # Update bid status
```

### Auctions
```
POST   /api/v1/auctions          # Create auction
GET    /api/v1/auctions          # Get active auctions
GET    /api/v1/auctions/:id      # Get specific auction
PUT    /api/v1/auctions/:id/close # Close auction
GET    /api/v1/auctions/stats    # Get auction statistics
```

### Statistics
```
GET    /api/v1/stats/dashboard   # Dashboard statistics
GET    /api/v1/stats/realtime    # Real-time statistics
```

### WebSocket
```
GET    /api/v1/ws/auction?property_id=xxx&user_id=xxx  # WebSocket connection
GET    /api/v1/ws/clients?property_id=xxx              # Connected clients count
```

### Demo Data
```
POST   /api/v1/demo/create       # Create demo data
DELETE /api/v1/demo/clear        # Clear demo data
GET    /api/v1/demo/status       # Check demo data status
```

## 🛠️ Installation & Setup

### Prerequisites
- Go 1.25+ 
- Redis Server
- Git

### 1. Clone Repository
```bash
git clone https://github.com/your-org/erea-api.git
cd erea-api
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Start Redis Server
```bash
# Using Docker
docker run -d -p 6379:6379 redis:alpine

# Or install locally
# macOS: brew install redis && redis-server
# Ubuntu: sudo apt install redis-server && redis-server
```

### 4. Run the API Server
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## 📝 Configuration

### Environment Variables
```bash
REDIS_HOST=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
PORT=8080
```

### Redis Configuration
The API uses Redis as the primary data store with the following key patterns:
- `user:{id}` - User data
- `property:{id}` - Property data
- `auction:{id}` - Auction data
- `bid:{id}` - Bid data
- `property_bids:{property_id}` - Property bid sets
- `property_auction:{property_id}` - Property-auction mapping

## 🧪 Testing

### 1. Check API Health
```bash
curl http://localhost:8080/health
```

### 2. Create Demo Data
```bash
curl -X POST http://localhost:8080/api/v1/demo/create
```

### 3. Get Properties
```bash
curl http://localhost:8080/api/v1/properties
```

### 4. Place a Bid
```bash
curl -X POST http://localhost:8080/api/v1/bids \
  -H "Content-Type: application/json" \
  -d '{
    "property_id": "property-id-here",
    "bidder_id": "user-id-here", 
    "amount": 700000000,
    "is_encrypted": true
  }'
```

### 5. WebSocket Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/auction?property_id=xxx');

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Real-time update:', data);
};
```

## 🔄 WebSocket Events

### Bid Update
```json
{
  "type": "bid_update",
  "data": {
    "property_id": "uuid",
    "new_bid": 750000000,
    "bidder_id": "uuid", 
    "bid_count": 5,
    "time_remaining": "2h30m"
  },
  "message": "New bid placed"
}
```

### Auction Update
```json
{
  "type": "auction_update",
  "data": {
    "property_id": "uuid",
    "status": "Closed",
    "winner_id": "uuid",
    "winning_bid": 1250000000
  },
  "message": "Auction status updated"
}
```

## 🏗️ Data Models

### Property
```json
{
  "id": "uuid",
  "title": "Gangnam District Premium Officetel",
  "location": "Sinsa-dong, Gangnam-gu, Seoul, South Korea",
  "description": "Modern officetel in Gangnam",
  "type": "Officetel",
  "area": 45.2,
  "starting_price": 500000000,
  "current_price": 650000000,
  "features": ["Near Subway", "24/7 Security"],
  "status": "Active",
  "end_date": "2024-12-30T15:00:00Z"
}
```

### Bid
```json
{
  "id": "uuid",
  "property_id": "uuid",
  "bidder_id": "uuid",
  "amount": 650000000,
  "tx_hash": "0x...",
  "status": "Confirmed",
  "is_encrypted": true,
  "created_at": "2024-12-20T14:30:00Z"
}
```

## 🔗 Frontend Integration

This API is designed to work with the EREA frontend React application:

### CORS Support
- Enabled for all origins (development)
- Configurable for production

### Data Synchronization
- Real-time updates via WebSocket
- RESTful API for standard operations
- Consistent data models

### Authentication Ready
- Token-based authentication structure
- User session management
- Role-based access control (ready for implementation)

## 📊 Performance

### Redis Optimization
- Efficient key patterns
- Set operations for relationships
- JSON serialization for complex data

### Concurrent Connections
- WebSocket connection pooling
- Goroutine-based request handling
- Redis connection pooling

## 🔮 Future Enhancements

- [ ] JWT Authentication
- [ ] Rate Limiting
- [ ] API Documentation (Swagger)
- [ ] Database Migration Tools
- [ ] Blockchain Integration
- [ ] Email Notifications
- [ ] File Upload Support
- [ ] Advanced Search
- [ ] Caching Layer
- [ ] Monitoring & Logging

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is in the public domain.

---

**EREA API** | Built with ❤️ for secure real estate auctions