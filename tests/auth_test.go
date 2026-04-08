package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baodhtv01/CRUD-go-lang/config"
	"github.com/baodhtv01/CRUD-go-lang/internal/handlers"
	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/repositories"
	"github.com/baodhtv01/CRUD-go-lang/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.User{}, &models.Post{})
	require.NoError(t, err)
	return db
}

func setupAuthRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret",
			ExpiresIn: 24,
		},
	}

	userRepo := repositories.NewUserRepository(db)
	authSvc := services.NewAuthService(userRepo, cfg)
	authHandler := handlers.NewAuthHandler(authSvc)

	r := gin.New()
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	return r
}

func TestRegister_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	r := setupAuthRouter(db)

	body := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestRegister_DuplicateEmail(t *testing.T) {
	db := setupAuthTestDB(t)
	r := setupAuthRouter(db)

	body := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	// First registration
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Duplicate registration
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestRegister_InvalidInput(t *testing.T) {
	db := setupAuthTestDB(t)
	r := setupAuthRouter(db)

	body := map[string]string{
		"name":  "A", // too short
		"email": "not-an-email",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	r := setupAuthRouter(db)

	// Register first
	regBody := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "login@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(regBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginBody := models.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}
	loginJSON, _ := json.Marshal(loginBody)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginJSON))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
}

func TestLogin_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	r := setupAuthRouter(db)

	// Register first
	regBody := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "wrongpwd@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(regBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Login with wrong password
	loginBody := models.LoginRequest{
		Email:    "wrongpwd@example.com",
		Password: "wrongpassword",
	}
	loginJSON, _ := json.Marshal(loginBody)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginJSON))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestLogin_NonExistentUser(t *testing.T) {
	db := setupAuthTestDB(t)
	r := setupAuthRouter(db)

	loginBody := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	loginJSON, _ := json.Marshal(loginBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
