package services

import "github.com/Bug-Bugger/ezmodel/internal/models"

type UserServiceInterface interface {
	CreateUser(name string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(id int64, name string) (*models.User, error)
	DeleteUser(id int64) error
}
