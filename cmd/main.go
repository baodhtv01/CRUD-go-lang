package main

import (
	"log"

	"github.com/baodhtv01/CRUD-go-lang/config"
	"github.com/baodhtv01/CRUD-go-lang/internal/handlers"
	"github.com/baodhtv01/CRUD-go-lang/internal/middleware"
	"github.com/baodhtv01/CRUD-go-lang/internal/repositories"
	"github.com/baodhtv01/CRUD-go-lang/internal/services"
	"github.com/baodhtv01/CRUD-go-lang/migrations"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := migrations.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	postRepo := repositories.NewPostRepository(db)

	// Services
	authSvc := services.NewAuthService(userRepo, cfg)
	userSvc := services.NewUserService(userRepo)
	postSvc := services.NewPostService(postRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	userHandler := handlers.NewUserHandler(userSvc)
	postHandler := handlers.NewPostHandler(postSvc)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())

	// Auth routes (public)
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("", middleware.AdminOnly(), userHandler.GetAll)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// Post routes
		posts := protected.Group("/posts")
		{
			posts.GET("", postHandler.GetAll)
			posts.GET("/:id", postHandler.GetByID)
			posts.POST("", postHandler.Create)
			posts.PUT("/:id", postHandler.Update)
			posts.DELETE("/:id", postHandler.Delete)
		}
	}

	log.Printf("Server running on port %s in %s mode", cfg.Server.Port, cfg.Server.Mode)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
