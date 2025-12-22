// пакет хендлеров аутентификации
package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gorono/internal/models"
)

// Handler структура для хендлеров аутентификации
type Handler struct {
	Store UserStore
}

// NewHandler создает новый хендлер аутентификации
func NewHandler(store UserStore) *Handler {
	return &Handler{Store: store}
}

// Register обрабатывает запрос на регистрацию пользователя
// router.HandleFunc("/register", authHandler.Register).Methods("POST")
func (h *Handler) Register(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Invalid request body"})
		return
	}
	defer req.Body.Close()

	var regReq RegisterRequest
	err = json.Unmarshal(body, &regReq)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Валидация
	if regReq.Username == "" || regReq.Password == "" {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Username and password are required"})
		return
	}

	if len(regReq.Username) < 3 {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Username must be at least 3 characters"})
		return
	}

	if len(regReq.Password) < 6 {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Password must be at least 6 characters"})
		return
	}

	// Хешируем пароль
	passwordHash, err := HashPassword(regReq.Password)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Error hashing password"})
		models.Sugar.Debugf("Error hashing password: %v", err)
		return
	}

	// Создаем пользователя
	user, err := h.Store.CreateUser(req.Context(), regReq.Username, passwordHash)
	if err != nil {
		// Проверяем на дублирование username
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			rwr.WriteHeader(http.StatusConflict)
			json.NewEncoder(rwr).Encode(map[string]string{"error": "Username already exists"})
			return
		}
		rwr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Error creating user"})
		models.Sugar.Debugf("Error creating user: %v", err)
		return
	}

	rwr.WriteHeader(http.StatusCreated)
	json.NewEncoder(rwr).Encode(map[string]interface{}{
		"message":  "User created successfully",
		"username": user.Username,
	})
}

// Login обрабатывает запрос на вход пользователя
// router.HandleFunc("/login", authHandler.Login).Methods("POST")
func (h *Handler) Login(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Invalid request body"})
		return
	}
	defer req.Body.Close()

	var loginReq LoginRequest
	err = json.Unmarshal(body, &loginReq)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Валидация
	if loginReq.Username == "" || loginReq.Password == "" {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Username and password are required"})
		return
	}

	// Получаем пользователя
	user, err := h.Store.GetUserByUsername(req.Context(), loginReq.Username)
	if err != nil {
		rwr.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Invalid username or password"})
		return
	}

	// Проверяем пароль
	if !CheckPassword(loginReq.Password, user.PasswordHash) {
		rwr.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Invalid username or password"})
		return
	}

	// Создаем токен
	token, err := h.Store.CreateToken(req.Context(), user.ID)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rwr).Encode(map[string]string{"error": "Error creating token"})
		models.Sugar.Debugf("Error creating token: %v", err)
		return
	}

	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(LoginResponse{
		Token:    token,
		Username: user.Username,
	})
}

// AuthMiddleware middleware для проверки аутентификации
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rwr http.ResponseWriter, req *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(rwr, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Ожидаем формат: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(rwr, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Валидируем токен
		user, err := h.Store.ValidateToken(req.Context(), token)
		if err != nil {
			http.Error(rwr, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Сохраняем информацию о пользователе в контексте для использования в хендлерах
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))
		req.Header.Set("X-Username", user.Username)

		next.ServeHTTP(rwr, req)
	})
}
