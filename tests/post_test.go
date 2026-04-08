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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupPostTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.User{}, &models.Post{})
	require.NoError(t, err)
	return db
}

func setupPostRouter(db *gorm.DB) (*gin.Engine, *config.Config) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret",
			ExpiresIn: 24,
		},
	}

	postRepo := repositories.NewPostRepository(db)
	postSvc := services.NewPostService(postRepo)
	postHandler := handlers.NewPostHandler(postSvc)

	r := gin.New()
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		protected.GET("/posts", postHandler.GetAll)
		protected.GET("/posts/:id", postHandler.GetByID)
		protected.POST("/posts", postHandler.Create)
		protected.PUT("/posts/:id", postHandler.Update)
		protected.DELETE("/posts/:id", postHandler.Delete)
	}

	return r, cfg
}

func createTestPost(t *testing.T, db *gorm.DB, title, content string, userID uint) *models.Post {
	post := &models.Post{
		Title:   title,
		Content: content,
		UserID:  userID,
	}
	err := db.Create(post).Error
	require.NoError(t, err)
	return post
}

func TestCreatePost_Success(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "poster@test.com", "user")
	token := generateTestToken(t, cfg, user)

	body := models.CreatePostRequest{
		Title:   "My First Post",
		Content: "Hello World!",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "My First Post", data["title"])
}

func TestCreatePost_InvalidInput(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "poster2@test.com", "user")
	token := generateTestToken(t, cfg, user)

	body := map[string]string{"title": ""} // missing content
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllPosts(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "getposts@test.com", "user")
	createTestPost(t, db, "Post 1", "Content 1", user.ID)
	createTestPost(t, db, "Post 2", "Content 2", user.ID)
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(2), resp["total"])
}

func TestGetPostByID(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "getpost@test.com", "user")
	post := createTestPost(t, db, "My Post", "Content", user.ID)
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/posts/%d", post.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPostByID_NotFound(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "notfound@test.com", "user")
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/9999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdatePost_OwnPost(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "updatepost@test.com", "user")
	post := createTestPost(t, db, "Original Title", "Original Content", user.ID)
	token := generateTestToken(t, cfg, user)

	body := models.UpdatePostRequest{Title: "Updated Title"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/posts/%d", post.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", data["title"])
}

func TestUpdatePost_OtherUserPost_Forbidden(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user1 := createTestUser(t, db, "User1", "user1post@test.com", "user")
	user2 := createTestUser(t, db, "User2", "user2post@test.com", "user")
	post := createTestPost(t, db, "User1 Post", "Content", user1.ID)
	token := generateTestToken(t, cfg, user2)

	body := models.UpdatePostRequest{Title: "Hacked Title"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/posts/%d", post.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeletePost_OwnPost(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "delpost@test.com", "user")
	post := createTestPost(t, db, "To Delete", "Content", user.ID)
	token := generateTestToken(t, cfg, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/posts/%d", post.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeletePost_AdminCanDeleteAny(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user := createTestUser(t, db, "User1", "userfordelete@test.com", "user")
	admin := createTestUser(t, db, "Admin", "admindelete@test.com", "admin")
	post := createTestPost(t, db, "User Post", "Content", user.ID)
	token := generateTestToken(t, cfg, admin)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/posts/%d", post.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeletePost_OtherUserPost_Forbidden(t *testing.T) {
	db := setupPostTestDB(t)
	r, cfg := setupPostRouter(db)

	user1 := createTestUser(t, db, "User1", "del1@test.com", "user")
	user2 := createTestUser(t, db, "User2", "del2@test.com", "user")
	post := createTestPost(t, db, "User1 Post", "Content", user1.ID)
	token := generateTestToken(t, cfg, user2)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/posts/%d", post.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPost_Unauthorized(t *testing.T) {
	db := setupPostTestDB(t)
	r, _ := setupPostRouter(db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
