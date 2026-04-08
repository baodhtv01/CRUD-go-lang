package services

import (
	"errors"

	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/repositories"
)

type PostService interface {
	Create(userID uint, req *models.CreatePostRequest) (*models.Post, error)
	GetAll(page, limit int) ([]models.Post, int64, error)
	GetByID(id uint) (*models.Post, error)
	Update(id uint, userID uint, role string, req *models.UpdatePostRequest) (*models.Post, error)
	Delete(id uint, userID uint, role string) error
}

type postService struct {
	postRepo repositories.PostRepository
}

func NewPostService(postRepo repositories.PostRepository) PostService {
	return &postService{postRepo: postRepo}
}

func (s *postService) Create(userID uint, req *models.CreatePostRequest) (*models.Post, error) {
	post := &models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) GetAll(page, limit int) ([]models.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.postRepo.FindAll(page, limit)
}

func (s *postService) GetByID(id uint) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *postService) Update(id uint, userID uint, role string, req *models.UpdatePostRequest) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	if role != "admin" && post.UserID != userID {
		return nil, errors.New("forbidden: you can only update your own posts")
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) Delete(id uint, userID uint, role string) error {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return errors.New("post not found")
	}

	if role != "admin" && post.UserID != userID {
		return errors.New("forbidden: you can only delete your own posts")
	}

	return s.postRepo.Delete(id)
}
