//go:build ignore

package main

import (
	"fmt"
	"log"

	"github.com/baodhtv01/CRUD-go-lang/config"
	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(&models.User{}, &models.Post{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// Seed admin user
	adminPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}
	admin := models.User{
		Name:     "Admin User",
		Email:    "admin@example.com",
		Password: adminPassword,
		Role:     "admin",
	}

	if err := db.FirstOrCreate(&admin, models.User{Email: admin.Email}).Error; err != nil {
		log.Printf("failed to create admin user: %v", err)
	} else {
		fmt.Printf("Admin user: %s (ID: %d)\n", admin.Email, admin.ID)
	}

	// Seed regular users
	users := []models.User{
		{Name: "Alice Smith", Email: "alice@example.com", Role: "user"},
		{Name: "Bob Jones", Email: "bob@example.com", Role: "user"},
	}

	for i := range users {
		password, err := utils.HashPassword("password123")
		if err != nil {
			log.Fatalf("failed to hash password for user %s: %v", users[i].Email, err)
		}
		users[i].Password = password
		if err := db.FirstOrCreate(&users[i], models.User{Email: users[i].Email}).Error; err != nil {
			log.Printf("failed to create user %s: %v", users[i].Email, err)
		} else {
			fmt.Printf("User: %s (ID: %d)\n", users[i].Email, users[i].ID)
		}
	}

	// Seed posts
	posts := []models.Post{
		{Title: "Admin's First Post", Content: "Hello from admin!", UserID: admin.ID},
		{Title: "Alice's Post", Content: "Hello from Alice!", UserID: users[0].ID},
		{Title: "Bob's Post", Content: "Hello from Bob!", UserID: users[1].ID},
	}

	for _, post := range posts {
		if err := db.FirstOrCreate(&post, models.Post{Title: post.Title, UserID: post.UserID}).Error; err != nil {
			log.Printf("failed to create post %s: %v", post.Title, err)
		} else {
			fmt.Printf("Post: %s (ID: %d)\n", post.Title, post.ID)
		}
	}

	fmt.Println("Seeding completed successfully!")
}
