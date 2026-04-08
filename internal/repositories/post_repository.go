package repositories

import (
	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *models.Post) error
	FindByID(id uint) (*models.Post, error)
	FindAll(page, limit int) ([]models.Post, int64, error)
	FindByUserID(userID uint, page, limit int) ([]models.Post, int64, error)
	Update(post *models.Post) error
	Delete(id uint) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *postRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) FindAll(page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("User").Offset(offset).Limit(limit).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) FindByUserID(userID uint, page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&models.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("user_id = ?", userID).Preload("User").Offset(offset).Limit(limit).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *postRepository) Delete(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}
