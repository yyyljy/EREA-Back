package config

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	ctx         = context.Background()
)

// InitRedis Redis 클라이언트를 초기화합니다
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 서버 주소
		Password: "",               // 비밀번호 없음
		DB:       0,                // 기본 DB 사용
	})

	// Redis 연결 테스트
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis 연결 실패: %v", err)
	}
	log.Printf("Redis 연결 성공: %s", pong)
}

// GetRedisClient Redis 클라이언트를 반환합니다
func GetRedisClient() *redis.Client {
	return RedisClient
}

// GetContext 컨텍스트를 반환합니다
func GetContext() context.Context {
	return ctx
}
