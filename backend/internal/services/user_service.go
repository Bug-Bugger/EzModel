package services

import (
	"errors"
	"strings"

	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Implement UserServiceInterface
func (s *UserService) CreateUser(name string) (*models.User, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, ErrInvalidInput
	}

	// Business logic: Check for duplicate names. Might be dropped in the future.
	existingUsers, err := s.userRepo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, user := range existingUsers {
		if strings.EqualFold(user.Name, name) {
			return nil, ErrUserAlreadyExists
		}
	}

	user := &models.User{
		Name: name,
	}

	id, err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (s *UserService) GetUserByID(id int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) UpdateUser(id int64, name string) (*models.User, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.Name = name
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(id int64) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return s.userRepo.Delete(id)
}
