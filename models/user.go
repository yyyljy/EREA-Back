package models

import (
	"encoding/json"
	"time"
)

// User 사용자 모델 구조체
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Age       int       `json:"age" binding:"required,min=1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToJSON User 구조체를 JSON 문자열로 변환
func (u *User) ToJSON() (string, error) {
	jsonData, err := json.Marshal(u)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON JSON 문자열을 User 구조체로 변환
func (u *User) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), u)
}

// UserResponse API 응답용 구조체
type UserResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreateUserRequest 사용자 생성 요청 구조체
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required,min=1"`
}

// UpdateUserRequest 사용자 업데이트 요청 구조체
type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age,omitempty"`
}
