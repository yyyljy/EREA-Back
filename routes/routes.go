package routes

import (
	"erea-api/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes API 라우트를 설정합니다
func SetupRoutes() *gin.Engine {
	// Gin 엔진 생성 (릴리즈 모드)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS 미들웨어 설정
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:     true,
		// AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		// AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// 헬스 체크 엔드포인트
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "EREA API 서버가 정상적으로 동작중입니다",
		})
	})

	// API v1 그룹
	v1 := r.Group("/api/v1")
	{
		// 사용자 관련 엔드포인트
		users := v1.Group("/users")
		{
			users.POST("/", handlers.CreateUser)       // 사용자 생성
			users.GET("/", handlers.GetAllUsers)       // 모든 사용자 조회
			users.GET("/:id", handlers.GetUser)        // 특정 사용자 조회
			users.PUT("/:id", handlers.UpdateUser)     // 사용자 정보 업데이트
			users.DELETE("/:id", handlers.DeleteUser)  // 사용자 삭제
			users.GET("/:id/bids", handlers.GetUserBids)  // 사용자 입찰 내역
			users.GET("/:id/stats", handlers.GetUserStats) // 사용자 통계
		}

		// 부동산 속성 관련 엔드포인트
		properties := v1.Group("/properties")
		{
			properties.POST("/", handlers.CreateProperty)         // 부동산 생성
			properties.GET("/", handlers.GetAllProperties)        // 모든 부동산 조회
			properties.GET("/status", handlers.GetPropertiesByStatus) // 상태별 부동산 조회
			properties.GET("/:id", handlers.GetProperty)          // 특정 부동산 조회
			properties.PUT("/:id", handlers.UpdateProperty)       // 부동산 정보 업데이트
			properties.DELETE("/:id", handlers.DeleteProperty)    // 부동산 삭제
			properties.GET("/:id/auction", handlers.GetPropertyAuction) // 부동산 경매 정보
			properties.GET("/:id/bids", handlers.GetBidHistory)         // 부동산 입찰 내역
			properties.GET("/:id/stats", handlers.GetPropertyStats)     // 부동산 통계
		}

		// 입찰 관련 엔드포인트
		bids := v1.Group("/bids")
		{
			bids.POST("/", handlers.PlaceBid)           // 입찰하기
			bids.GET("/", handlers.GetTopBids)          // 상위 입찰 조회
			bids.GET("/:id", handlers.GetBid)           // 특정 입찰 조회
			bids.PUT("/:id/status", handlers.UpdateBidStatus) // 입찰 상태 업데이트
		}

		// 보증금 관련 엔드포인트
		deposits := v1.Group("/deposits")
		{
			deposits.POST("/", handlers.CreateDeposit)           // 보증금 납부
			deposits.GET("/", handlers.GetAllDeposits)           // 모든 보증금 조회
			deposits.GET("/user/:userId", handlers.GetUserDeposits) // 사용자별 보증금 조회
			deposits.GET("/:id", handlers.GetDeposit)            // 특정 보증금 조회
			deposits.PUT("/:id/status", handlers.UpdateDepositStatus) // 보증금 상태 업데이트
		}

		// 경매 관련 엔드포인트
		auctions := v1.Group("/auctions")
		{
			auctions.POST("/", handlers.CreateAuction)     // 경매 생성
			auctions.GET("/", handlers.GetActiveAuctions)  // 활성 경매 조회
			auctions.GET("/:id", handlers.GetAuction)      // 특정 경매 조회
			auctions.PUT("/:id/close", handlers.CloseAuction) // 경매 종료
			auctions.GET("/stats", handlers.GetAuctionStats)  // 경매 통계
		}

		// 통계 및 대시보드 엔드포인트
		stats := v1.Group("/stats")
		{
			stats.GET("/dashboard", handlers.GetDashboardStats) // 대시보드 통계
			stats.GET("/realtime", handlers.GetRealtimeStats)   // 실시간 통계
		}

		// WebSocket 엔드포인트
		ws := v1.Group("/ws")
		{
			ws.GET("/auction", handlers.HandleWebSocket)         // WebSocket 연결
			ws.GET("/clients", handlers.GetConnectedClients)     // 연결된 클라이언트 수
		}

		// 데모 데이터 엔드포인트
		demo := v1.Group("/demo")
		{
			demo.POST("/create", handlers.CreateDemoData)    // 데모 데이터 생성
			demo.DELETE("/clear", handlers.ClearDemoData)    // 데모 데이터 삭제
			demo.GET("/status", handlers.GetDemoStatus)      // 데모 데이터 상태 확인
		}
	}

	return r
}
