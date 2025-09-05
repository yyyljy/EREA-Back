package main

import (
	"erea-api/config"
	"erea-api/routes"
	"log"
)

func main() {
	// Redis 연결 초기화
	log.Println("Redis 연결을 초기화하는 중...")
	config.InitRedis()

	// 라우터 설정
	log.Println("라우터를 설정하는 중...")
	router := routes.SetupRoutes()

	// 서버 시작
	port := ":8080"
	log.Printf("서버가 포트 %s에서 시작됩니다...", port)
	log.Printf("헬스 체크: http://localhost%s/health", port)
	log.Printf("API 문서: http://localhost%s/api/v1/users", port)

	if err := router.Run(port); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
