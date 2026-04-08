package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baodhtv01/CRUD-go-lang/config"
	"github.com/baodhtv01/CRUD-go-lang/internal/handlers"
	"github.com/baodhtv01/CRUD-go-lang/internal/middleware"
	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/repositories"
	"github.com/baodhtv01/CRUD-go-lang/internal/services"
	"github.com/baodhtv01/CRUD-go-lang/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.User{}, &models.Post{})
	require.NoError(t, err)
	return db
}

func setupUserRouter(db *gorm.DB) (*gin.Engine, *config.Config) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret",
			ExpiresIn: 24,
		},
	}

	userRepo := repositories.NewUserRepository(db)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	r := gin.New()
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		protected.GET("/users", middleware.AdminOnly(), userHandler.GetAll)
		protected.GET("/users/:id", userHandler.GetByID)
		protected.PUT("/users/:id", userHandler.Update)
		protected.DELETE("/users/:id", userHandler.Delete)
	}

	return r, cfg
}

func createTestUser(t *testing.T, db *gorm.DB, name, email, role string) *models.User {
	hashedPwd, err := utils.HashPassword("password123")
	require.NoError(t, err)
	user := &models.User{
		Name:     name,
		Email:    email,
		Password: hashedPwd,
		Role:     role,
	}
	err = db.Create(user).Error
	require.NoError(t, err)
	return user
}

func generateTestToken(t *testing.T, cfg *config.Config, user *models.User) string {
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	require.NoError(t, err)
	return token
}

func TestGetAllUsers_Admin(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	admin := createTestUser(t, db, "Admin", "admin@test.com", "admin")
	createTestUser(t, db, "User1", "user1@test.com", "user")
	token := generateTestToken(t, cfg, admin)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestGetAllUsers_NonAdmin(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	user := createTestUser(t, db, "User1", "user1@test.com", "user")
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetUserByID_OwnProfile(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	user := createTestUser(t, db, "User1", "user1@test.com", "user")
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%d", user.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserByID_OtherProfile_Forbidden(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	user1 := createTestUser(t, db, "User1", "user1@test.com", "user")
	user2 := createTestUser(t, db, "User2", "user2@test.com", "user")
	token := generateTestToken(t, cfg, user1)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%d", user2.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	user := createTestUser(t, db, "User1", "user1@test.com", "user")
	token := generateTestToken(t, cfg, user)

	updateBody := models.UpdateUserRequest{Name: "Updated Name"}
	jsonBody, _ := json.Marshal(updateBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/users/%d", user.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Name", data["name"])
}

func TestDeleteUser_OwnAccount(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	user := createTestUser(t, db, "User1", "user1@test.com", "user")
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/users/%d", user.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteUser_OtherAccount_Forbidden(t *testing.T) {
	db := setupUserTestDB(t)
	r, cfg := setupUserRouter(db)

	user1 := createTestUser(t, db, "User1", "user1@test.com", "user")
	user2 := createTestUser(t, db, "User2", "user2@test.com", "user")
	token := generateTestToken(t, cfg, user1)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/users/%d", user2.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetUser_Unauthorized(t *testing.T) {
	db := setupUserTestDB(t)
	r, _ := setupUserRouter(db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
