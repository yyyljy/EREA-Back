package handlers

import (
	"erea-api/config"
	"erea-api/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const userKeyPrefix = "user:"

// CreateUser 새로운 사용자를 생성합니다
func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "잘못된 요청 데이터",
			Error:   err.Error(),
		})
		return
	}

	// 새로운 사용자 생성
	user := models.User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Redis에 사용자 저장
	userJSON, err := user.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 데이터 변환 실패",
			Error:   err.Error(),
		})
		return
	}

	key := userKeyPrefix + user.ID
	err = config.GetRedisClient().Set(config.GetContext(), key, userJSON, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 저장 실패",
			Error:   err.Error(),
		})
		return
	}

	// 사용자 목록에 ID 추가 (검색을 위해)
	err = config.GetRedisClient().SAdd(config.GetContext(), "users", user.ID).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 목록 업데이트 실패",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		Success: true,
		Message: "사용자가 성공적으로 생성되었습니다",
		Data:    user,
	})
}

// GetUser ID로 사용자를 조회합니다
func GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "사용자 ID가 필요합니다",
		})
		return
	}

	key := userKeyPrefix + userID
	userJSON, err := config.GetRedisClient().Get(config.GetContext(), key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "사용자를 찾을 수 없습니다",
			Error:   err.Error(),
		})
		return
	}

	var user models.User
	err = user.FromJSON(userJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 데이터 변환 실패",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "사용자 조회 성공",
		Data:    user,
	})
}

// GetAllUsers 모든 사용자를 조회합니다
func GetAllUsers(c *gin.Context) {
	// 사용자 ID 목록 가져오기
	userIDs, err := config.GetRedisClient().SMembers(config.GetContext(), "users").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 목록 조회 실패",
			Error:   err.Error(),
		})
		return
	}

	if len(userIDs) == 0 {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: true,
			Message: "등록된 사용자가 없습니다",
			Data:    []models.User{},
		})
		return
	}

	// 각 사용자 데이터 가져오기
	var users []models.User
	for _, userID := range userIDs {
		key := userKeyPrefix + userID
		userJSON, err := config.GetRedisClient().Get(config.GetContext(), key).Result()
		if err != nil {
			continue // 해당 사용자 데이터가 없으면 건너뛰기
		}

		var user models.User
		err = user.FromJSON(userJSON)
		if err != nil {
			continue // JSON 변환 실패시 건너뛰기
		}

		users = append(users, user)
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: fmt.Sprintf("%d명의 사용자를 조회했습니다", len(users)),
		Data:    users,
	})
}

// UpdateUser 사용자 정보를 업데이트합니다
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "사용자 ID가 필요합니다",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "잘못된 요청 데이터",
			Error:   err.Error(),
		})
		return
	}

	// 기존 사용자 데이터 조회
	key := userKeyPrefix + userID
	userJSON, err := config.GetRedisClient().Get(config.GetContext(), key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "사용자를 찾을 수 없습니다",
			Error:   err.Error(),
		})
		return
	}

	var user models.User
	err = user.FromJSON(userJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 데이터 변환 실패",
			Error:   err.Error(),
		})
		return
	}

	// 업데이트할 필드만 변경
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Age != 0 {
		user.Age = req.Age
	}
	user.UpdatedAt = time.Now()

	// 업데이트된 사용자 저장
	updatedUserJSON, err := user.ToJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 데이터 변환 실패",
			Error:   err.Error(),
		})
		return
	}

	err = config.GetRedisClient().Set(config.GetContext(), key, updatedUserJSON, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 업데이트 실패",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "사용자 정보가 성공적으로 업데이트되었습니다",
		Data:    user,
	})
}

// DeleteUser 사용자를 삭제합니다
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "사용자 ID가 필요합니다",
		})
		return
	}

	key := userKeyPrefix + userID
	
	// 사용자 존재 확인
	exists, err := config.GetRedisClient().Exists(config.GetContext(), key).Result()
	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "사용자를 찾을 수 없습니다",
		})
		return
	}

	// 사용자 데이터 삭제
	err = config.GetRedisClient().Del(config.GetContext(), key).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 삭제 실패",
			Error:   err.Error(),
		})
		return
	}

	// 사용자 목록에서 ID 제거
	err = config.GetRedisClient().SRem(config.GetContext(), "users", userID).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UserResponse{
			Success: false,
			Message: "사용자 목록 업데이트 실패",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "사용자가 성공적으로 삭제되었습니다",
	})
}
