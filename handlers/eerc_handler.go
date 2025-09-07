package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// ZK Service 설정 (환경변수로 설정 가능)
var ZK_SERVICE_URL = getEnvOrDefault("ZK_SERVICE_URL", "http://localhost:3001")

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ZK Proof 요청/응답 구조체들
type MintProofRequest struct {
	Amount           int64     `json:"amount"`
	UserPublicKey    []string  `json:"userPublicKey"`
	AuditorPublicKey []string  `json:"auditorPublicKey"`
}

type TransferProofRequest struct {
	Amount                int64     `json:"amount"`
	SenderPublicKey       []string  `json:"senderPublicKey"`
	SenderPrivateKey      string    `json:"senderPrivateKey"`
	SenderBalance         int64     `json:"senderBalance"`
	SenderEncryptedBalance []string `json:"senderEncryptedBalance"`
	ReceiverPublicKey     []string  `json:"receiverPublicKey"`
	AuditorPublicKey      []string  `json:"auditorPublicKey"`
}

type ZKProofResponse struct {
	Success bool        `json:"success"`
	Proof   interface{} `json:"proof"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
}

type ZKMintResponse struct {
	Success bool   `json:"success"`
	TxHash  string `json:"txHash"`
	Amount  int64  `json:"amount"`
	Output  string `json:"output"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type TransferProofResponse struct {
	Success         bool        `json:"success"`
	Proof           interface{} `json:"proof"`
	SenderBalancePCT []string   `json:"senderBalancePCT"`
	Message         string      `json:"message"`
	Error           string      `json:"error,omitempty"`
}

// HTTP 클라이언트 (재사용을 위해 전역으로 선언)
var httpClient = &http.Client{
	Timeout: 60 * time.Second, // ZK proof 생성은 시간이 오래 걸릴 수 있음
}

// EERC 토큰 민팅 (ZK proof 생성 포함)
func MintEERCTokens(c *gin.Context) {
	var req MintProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// 입력값 검증
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Amount must be greater than 0",
		})
		return
	}

	if len(req.UserPublicKey) != 2 || len(req.AuditorPublicKey) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Public keys must have exactly 2 elements [x, y]",
		})
		return
	}

	fmt.Printf("🪙 Processing EERC mint request for amount: %d\n", req.Amount)
	
	// ZK Service에 실제 mint 요청 (05_mint.ts 실행)
	mintRequest := map[string]interface{}{
		"userAddress": "0xE762Abfd0920B06050f5ee0bf3736F933cE6837b", // 실제 user address (config에서 가져와야 함)
		"amount":      req.Amount,
	}
	
	zkResponse, err := callZKService("/api/zk/mint", mintRequest)
	if err != nil {
		fmt.Printf("❌ ZK mint service call failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to execute ZK mint: " + err.Error(),
		})
		return
	}

	// ZK mint 응답 파싱
	var mintResponse ZKMintResponse
	if err := json.Unmarshal(zkResponse, &mintResponse); err != nil {
		fmt.Printf("❌ Failed to parse ZK mint response: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse ZK mint response",
		})
		return
	}

	if !mintResponse.Success {
		fmt.Printf("❌ ZK mint failed: %s\n", mintResponse.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "ZK mint failed: " + mintResponse.Error,
		})
		return
	}

	fmt.Printf("✅ ZK mint completed successfully. TxHash: %s\n", mintResponse.TxHash)

	// 성공 응답 (실제 ZK mint 결과 사용)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"txHash":  mintResponse.TxHash,
		"amount":  req.Amount,
		"message": "EERC tokens minted successfully via ZK service",
		"output":  mintResponse.Output, // 디버깅용 출력
	})
}

// EERC 토큰 전송 (ZK proof 생성 포함)
func TransferEERCTokens(c *gin.Context) {
	var req TransferProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// 입력값 검증
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Amount must be greater than 0",
		})
		return
	}

	if req.SenderBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Insufficient balance",
		})
		return
	}

	fmt.Printf("🔄 Processing EERC transfer request for amount: %d\n", req.Amount)
	
	// ZK Service에 실제 transfer 요청 (07_transfer.ts 실행)
	transferRequest := map[string]interface{}{
		"fromAddress": "0xE762Abfd0920B06050f5ee0bf3736F933cE6837b", // User address
		"toAddress":   "0x1061538525312768d0da8b9E7a44a5757291fB5E", // Admin address  
		"amount":      req.Amount,
	}
	
	zkResponse, err := callZKService("/api/zk/transfer", transferRequest)
	if err != nil {
		fmt.Printf("❌ ZK transfer service call failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to execute ZK transfer: " + err.Error(),
		})
		return
	}

	// ZK transfer 응답 파싱
	var transferResponse ZKMintResponse // mint와 같은 구조 사용
	if err := json.Unmarshal(zkResponse, &transferResponse); err != nil {
		fmt.Printf("❌ Failed to parse ZK transfer response: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse ZK transfer response",
		})
		return
	}

	if !transferResponse.Success {
		fmt.Printf("❌ ZK transfer failed: %s\n", transferResponse.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "ZK transfer failed: " + transferResponse.Error,
		})
		return
	}

	fmt.Printf("✅ ZK transfer completed successfully. TxHash: %s\n", transferResponse.TxHash)

	// 성공 응답 (실제 ZK transfer 결과 사용)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"txHash":  transferResponse.TxHash,
		"amount":  req.Amount,
		"message": "EERC tokens transferred successfully via ZK service",
		"output":  transferResponse.Output, // 디버깅용 출력
	})
}

// ZK Service 호출 헬퍼 함수
func callZKService(endpoint string, payload interface{}) ([]byte, error) {
	// 요청 데이터를 JSON으로 변환
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// HTTP 요청 생성
	url := ZK_SERVICE_URL + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("🔗 Calling ZK Service: %s\n", url)
	
	// 요청 전송
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ZK service: %v", err)
	}
	defer resp.Body.Close()

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ZK service returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Mock 트랜잭션 해시 생성 (실제로는 블록체인에서 받아와야 함)
func generateMockTxHash() string {
	return fmt.Sprintf("0x%032x", time.Now().UnixNano())
}
