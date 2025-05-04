package repository

import "github.com/Bug-Bugger/ezmodel/internal/models"

type UserRepositoryInterface interface {
	Create(user *models.User) (int64, error)
	GetByID(id int64) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id int64) error
}
