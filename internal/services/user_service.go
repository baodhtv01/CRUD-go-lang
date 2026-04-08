package services

import (
	"errors"

	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/repositories"
)

type UserService interface {
	GetAll(page, limit int) ([]models.User, int64, error)
	GetByID(id uint) (*models.User, error)
	Update(id uint, req *models.UpdateUserRequest) (*models.User, error)
	Delete(id uint) error
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetAll(page, limit int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.userRepo.FindAll(page, limit)
}

func (s *userService) GetByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) Update(id uint, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check email uniqueness if changed
		if req.Email != user.Email {
			existing, _ := s.userRepo.FindByEmail(req.Email)
			if existing != nil {
				return nil, errors.New("email already in use")
			}
		}
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Delete(id uint) error {
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	return s.userRepo.Delete(id)
}
