package main

import (
	"encoding/json"
	"erea-api/config"
	"erea-api/routes"
	"log"
	"time"
)

// insertDummyData Redis에 더미 데이터를 삽입합니다
func insertDummyData() {
	ctx := config.GetContext()
	client := config.GetRedisClient()
	
	// 더미 사용자 데이터
	dummyUser := map[string]interface{}{
		"id":         "dummy_user_001",
		"name":       "테스트 사용자",
		"email":      "test@example.com",
		"created_at": time.Now().Format(time.RFC3339),
		"status":     "active",
	}
	
	// JSON으로 변환
	userData, err := json.Marshal(dummyUser)
	if err != nil {
		log.Printf("더미 데이터 JSON 변환 실패: %v", err)
		return
	}
	
	// Redis에 저장
	key := "user:dummy_user_001"
	err = client.Set(ctx, key, userData, 0).Err()
	if err != nil {
		log.Printf("더미 데이터 저장 실패: %v", err)
		return
	}
	
	log.Printf("✅ 더미 데이터 저장 성공: %s", key)
	
	// 저장된 데이터 확인
	savedData, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("더미 데이터 조회 실패: %v", err)
		return
	}
	
	log.Printf("📁 저장된 더미 데이터: %s", savedData)
	
	// 추가로 간단한 카운터도 설정
	counterKey := "test:counter"
	err = client.Set(ctx, counterKey, "1", 0).Err()
	if err != nil {
		log.Printf("카운터 저장 실패: %v", err)
	} else {
		log.Printf("✅ 테스트 카운터 저장 성공: %s = 1", counterKey)
	}
}

func main() {
	// Redis 연결 초기화
	log.Println("Redis 연결을 초기화하는 중...")
	config.InitRedis()

	// 더미 데이터 삽입
	// log.Println("더미 데이터를 삽입하는 중...")
	// insertDummyData()

	// 라우터 설정
	log.Println("라우터를 설정하는 중...")
	router := routes.SetupRoutes()

	// 서버 시작
	port := ":8000"
	log.Printf("서버가 포트 %s에서 시작됩니다...", port)
	log.Printf("헬스 체크: http://localhost%s/health", port)
	log.Printf("API 문서: http://localhost%s/api/v1/users", port)

	if err := router.Run(port); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
