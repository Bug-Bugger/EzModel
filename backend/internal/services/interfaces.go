package services

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
)

type UserServiceInterface interface {
	CreateUser(email, username, password string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(id uuid.UUID, req *dto.UpdateUserRequest) (*models.User, error)
	UpdatePassword(id uuid.UUID, password string) error
	DeleteUser(id uuid.UUID) error
	AuthenticateUser(email, password string) (*models.User, error)
}
