package migrations

import (
	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Post{},
	)
}
