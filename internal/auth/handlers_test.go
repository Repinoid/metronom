package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"gorono/internal/basis"
	"gorono/internal/models"
)

func setupTestHandler(t *testing.T) *Handler {
	// Инициализируем логгер для тестов
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	models.Sugar = *logger.Sugar()

	// Подключаемся к тестовой базе данных
	ctx := context.Background()
	models.DBEndPoint = "postgres://postgres:passwordas@localhost:5432/forgo"
	
	dbStorage, err := basis.InitDBStorage(ctx, models.DBEndPoint)
	if err != nil {
		t.Skip("Database not available, skipping test")
		return nil
	}

	// Создаем хранилище пользователей
	userStore, err := NewDBUserStore(dbStorage.DB)
	require.NoError(t, err)

	// Очищаем таблицы перед тестами
	_, err = dbStorage.DB.Exec(ctx, "DELETE FROM tokens")
	require.NoError(t, err)
	_, err = dbStorage.DB.Exec(ctx, "DELETE FROM users")
	require.NoError(t, err)

	return NewHandler(userStore)
}

func TestRegisterHandler(t *testing.T) {
	handler := setupTestHandler(t)
	if handler == nil {
		return
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Username: "newuser",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "User created successfully", resp["message"])
				assert.Equal(t, "newuser", resp["username"])
			},
		},
		{
			name: "username too short",
			requestBody: RegisterRequest{
				Username: "ab",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "at least 3 characters")
			},
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Username: "validuser",
				Password: "12345",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "at least 6 characters")
			},
		},
		{
			name: "missing username",
			requestBody: RegisterRequest{
				Username: "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "required")
			},
		},
		{
			name: "missing password",
			requestBody: RegisterRequest{
				Username: "validuser",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestRegisterDuplicateUser(t *testing.T) {
	handler := setupTestHandler(t)
	if handler == nil {
		return
	}

	username := "duplicateuser"
	regReq := RegisterRequest{
		Username: username,
		Password: "password123",
	}

	// Первая регистрация должна быть успешной
	body, _ := json.Marshal(regReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Вторая регистрация с тем же именем должна быть отклонена
	req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.Register(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "already exists")
}

func TestLoginHandler(t *testing.T) {
	handler := setupTestHandler(t)
	if handler == nil {
		return
	}

	// Сначала регистрируем пользователя
	username := "loginuser"
	password := "password123"
	regReq := RegisterRequest{
		Username: username,
		Password: password,
	}
	
	body, _ := json.Marshal(regReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Register(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	tests := []struct {
		name           string
		requestBody    LoginRequest
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Username: username,
				Password: password,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.NotEmpty(t, resp["token"])
				assert.Equal(t, username, resp["username"])
			},
		},
		{
			name: "wrong password",
			requestBody: LoginRequest{
				Username: username,
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "Invalid username or password")
			},
		},
		{
			name: "nonexistent user",
			requestBody: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "Invalid username or password")
			},
		},
		{
			name: "missing username",
			requestBody: LoginRequest{
				Username: "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp["error"], "required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	handler := setupTestHandler(t)
	if handler == nil {
		return
	}

	// Регистрируем и логинимся
	username := "middlewareuser"
	password := "password123"
	
	regReq := RegisterRequest{Username: username, Password: password}
	body, _ := json.Marshal(regReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Register(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	loginReq := LoginRequest{Username: username, Password: password}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	handler.Login(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var loginResp LoginResponse
	json.Unmarshal(rr.Body.Bytes(), &loginResp)
	token := loginResp.Token

	// Создаем тестовый handler который будет вызван после middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected resource"))
	})

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.AuthMiddleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "Protected resource", rr.Body.String())
	})

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rr := httptest.NewRecorder()

		handler.AuthMiddleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid token format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rr := httptest.NewRecorder()

		handler.AuthMiddleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalidtoken123")
		rr := httptest.NewRecorder()

		handler.AuthMiddleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
