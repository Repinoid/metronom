package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"gorono/internal/basis"
	"gorono/internal/models"
)

// setupTestDB создает тестовую базу данных для тестирования
func setupTestDB(t *testing.T) *DBUserStore {
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

	return userStore
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	
	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	
	// Проверяем что можем валидировать пароль
	assert.True(t, CheckPassword(password, hash))
	assert.False(t, CheckPassword("wrongpassword", hash))
}

func TestCreateUser(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	
	ctx := context.Background()
	username := "testuser"
	password := "password123"
	
	passwordHash, err := HashPassword(password)
	require.NoError(t, err)
	
	user, err := store.CreateUser(ctx, username, passwordHash)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Greater(t, user.ID, 0)
	assert.NotZero(t, user.CreatedAt)
	
	// Проверяем что нельзя создать пользователя с тем же именем
	_, err = store.CreateUser(ctx, username, passwordHash)
	assert.Error(t, err)
}

func TestGetUserByUsername(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	
	ctx := context.Background()
	username := "testuser2"
	password := "password123"
	
	passwordHash, err := HashPassword(password)
	require.NoError(t, err)
	
	// Создаем пользователя
	createdUser, err := store.CreateUser(ctx, username, passwordHash)
	require.NoError(t, err)
	
	// Получаем пользователя по имени
	user, err := store.GetUserByUsername(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, createdUser.Username, user.Username)
	assert.True(t, CheckPassword(password, user.PasswordHash))
	
	// Проверяем что несуществующий пользователь не найден
	_, err = store.GetUserByUsername(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCreateAndValidateToken(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	
	ctx := context.Background()
	username := "testuser3"
	password := "password123"
	
	passwordHash, err := HashPassword(password)
	require.NoError(t, err)
	
	// Создаем пользователя
	user, err := store.CreateUser(ctx, username, passwordHash)
	require.NoError(t, err)
	
	// Создаем токен
	token, err := store.CreateToken(ctx, user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Валидируем токен
	validatedUser, err := store.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, user.Username, validatedUser.Username)
	
	// Проверяем что неверный токен не валидируется
	_, err = store.ValidateToken(ctx, "invalidtoken")
	assert.Error(t, err)
}

func TestTokenExpiration(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		return
	}
	
	ctx := context.Background()
	username := "testuser4"
	password := "password123"
	
	passwordHash, err := HashPassword(password)
	require.NoError(t, err)
	
	// Создаем пользователя
	user, err := store.CreateUser(ctx, username, passwordHash)
	require.NoError(t, err)
	
	// Создаем токен с истекшим сроком (для теста вставляем напрямую в БД)
	expiredToken := "expiredtoken123"
	expiresAt := time.Now().Add(-1 * time.Hour) // истек час назад
	
	_, err = store.DB.Exec(ctx, 
		"INSERT INTO tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		user.ID, expiredToken, expiresAt,
	)
	require.NoError(t, err)
	
	// Проверяем что истекший токен не валидируется
	_, err = store.ValidateToken(ctx, expiredToken)
	assert.Error(t, err)
}
