# EREA API - Redis CRUD Backend

Redis를 데이터베이스로 사용하는 Go Gin 기반의 RESTful API 서버입니다.

## 기능

- 사용자 CRUD 작업 (생성, 조회, 수정, 삭제)
- Redis를 데이터 저장소로 사용
- RESTful API 엔드포인트 제공
- JSON 기반 데이터 교환

## 사전 요구사항

1. Go 1.21 이상
2. Redis 서버 (localhost:6379에서 실행)

## 설치 및 실행

1. 의존성 설치:
```bash
go mod tidy
```

2. Redis 서버 시작:
```bash
redis-server
```

3. 애플리케이션 실행:
```bash
go run main.go
```

서버는 `http://localhost:8080`에서 실행됩니다.

## API 엔드포인트

### 헬스 체크
- `GET /health` - 서버 상태 확인

### 사용자 관리

#### 1. 사용자 생성
- **URL**: `POST /api/v1/users`
- **요청 본문**:
```json
{
    "name": "홍길동",
    "email": "hong@example.com",
    "age": 25
}
```
- **응답**:
```json
{
    "success": true,
    "message": "사용자가 성공적으로 생성되었습니다",
    "data": {
        "id": "generated-uuid",
        "name": "홍길동",
        "email": "hong@example.com",
        "age": 25,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

#### 2. 모든 사용자 조회
- **URL**: `GET /api/v1/users`
- **응답**:
```json
{
    "success": true,
    "message": "2명의 사용자를 조회했습니다",
    "data": [
        {
            "id": "user-id-1",
            "name": "홍길동",
            "email": "hong@example.com",
            "age": 25,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        }
    ]
}
```

#### 3. 특정 사용자 조회
- **URL**: `GET /api/v1/users/{id}`
- **응답**:
```json
{
    "success": true,
    "message": "사용자 조회 성공",
    "data": {
        "id": "user-id",
        "name": "홍길동",
        "email": "hong@example.com",
        "age": 25,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

#### 4. 사용자 정보 업데이트
- **URL**: `PUT /api/v1/users/{id}`
- **요청 본문** (모든 필드 선택사항):
```json
{
    "name": "김철수",
    "email": "kim@example.com",
    "age": 30
}
```
- **응답**:
```json
{
    "success": true,
    "message": "사용자 정보가 성공적으로 업데이트되었습니다",
    "data": {
        "id": "user-id",
        "name": "김철수",
        "email": "kim@example.com",
        "age": 30,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T01:00:00Z"
    }
}
```

#### 5. 사용자 삭제
- **URL**: `DELETE /api/v1/users/{id}`
- **응답**:
```json
{
    "success": true,
    "message": "사용자가 성공적으로 삭제되었습니다"
}
```

## 사용 예시 (curl)

### 사용자 생성
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "홍길동",
    "email": "hong@example.com",
    "age": 25
  }'
```

### 모든 사용자 조회
```bash
curl http://localhost:8080/api/v1/users
```

### 특정 사용자 조회
```bash
curl http://localhost:8080/api/v1/users/{user-id}
```

### 사용자 정보 업데이트
```bash
curl -X PUT http://localhost:8080/api/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "김철수",
    "age": 30
  }'
```

### 사용자 삭제
```bash
curl -X DELETE http://localhost:8080/api/v1/users/{user-id}
```

## 프로젝트 구조

```
erea-api/
├── main.go                 # 메인 애플리케이션 엔트리포인트
├── go.mod                  # Go 모듈 파일
├── config/
│   └── redis.go           # Redis 연결 설정
├── models/
│   └── user.go            # 사용자 모델 및 구조체
├── handlers/
│   └── user_handler.go    # HTTP 핸들러 함수
└── routes/
    └── routes.go          # API 라우트 설정
```

## Redis 데이터 구조

- `user:{id}`: 사용자 데이터 (JSON 형태)
- `users`: 모든 사용자 ID를 저장하는 Set

## 에러 처리

모든 API 응답은 다음과 같은 형태를 가집니다:

### 성공 응답
```json
{
    "success": true,
    "message": "작업 성공 메시지",
    "data": { /* 응답 데이터 */ }
}
```

### 에러 응답
```json
{
    "success": false,
    "message": "에러 메시지",
    "error": "상세 에러 정보"
}
```

## 개발자 정보

EREA API는 Redis를 백엔드 데이터베이스로 사용하는 간단한 CRUD API 서버입니다.
# EREA-Back
