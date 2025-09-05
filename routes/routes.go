package routes

import (
	"erea-api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes API 라우트를 설정합니다
func SetupRoutes() *gin.Engine {
	// Gin 엔진 생성 (릴리즈 모드)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS 미들웨어 설정
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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
		}
	}

	return r
}
