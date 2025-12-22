// пакет аутентификации пользователей
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// User представляет пользователя в системе
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // не экспортируем в JSON
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest структура запроса на логин
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest структура запроса на регистрацию
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse структура ответа при успешном логине
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// UserStore интерфейс для работы с пользователями
type UserStore interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
}

// DBUserStore реализация хранилища пользователей в PostgreSQL
type DBUserStore struct {
	DB *pgxpool.Pool
}

// NewDBUserStore создает новое хранилище пользователей
func NewDBUserStore(db *pgxpool.Pool) (*DBUserStore, error) {
	store := &DBUserStore{DB: db}
	ctx := context.Background()
	
	// Создаем таблицу пользователей если не существует
	err := store.createUsersTable(ctx)
	if err != nil {
		return nil, err
	}
	
	// Создаем таблицу токенов если не существует
	err = store.createTokensTable(ctx)
	if err != nil {
		return nil, err
	}
	
	return store, nil
}

// createUsersTable создает таблицу пользователей
func (s *DBUserStore) createUsersTable(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err := s.DB.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}
	return nil
}

// createTokensTable создает таблицу токенов
func (s *DBUserStore) createTokensTable(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS tokens (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(100) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL
	);`
	
	_, err := s.DB.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("error creating tokens table: %w", err)
	}
	return nil
}

// CreateUser создает нового пользователя
func (s *DBUserStore) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	user := &User{}
	query := `INSERT INTO users (username, password_hash) 
	          VALUES ($1, $2) 
	          RETURNING id, username, password_hash, created_at`
	
	err := s.DB.QueryRow(ctx, query, username, passwordHash).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	
	return user, nil
}

// GetUserByUsername получает пользователя по имени
func (s *DBUserStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`
	
	err := s.DB.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	return user, nil
}

// CreateToken создает новый токен для пользователя
func (s *DBUserStore) CreateToken(ctx context.Context, userID int) (string, error) {
	// Генерируем случайный токен
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	
	// Токен действителен 24 часа
	expiresAt := time.Now().Add(24 * time.Hour)
	
	query := `INSERT INTO tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err = s.DB.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("error saving token: %w", err)
	}
	
	return token, nil
}

// ValidateToken проверяет токен и возвращает пользователя
func (s *DBUserStore) ValidateToken(ctx context.Context, token string) (*User, error) {
	user := &User{}
	query := `SELECT u.id, u.username, u.password_hash, u.created_at 
	          FROM users u 
	          JOIN tokens t ON u.id = t.user_id 
	          WHERE t.token = $1 AND t.expires_at > NOW()`
	
	err := s.DB.QueryRow(ctx, query, token).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}
	
	return user, nil
}

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword проверяет соответствие пароля хешу
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
