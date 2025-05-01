package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(user *models.User) (int64, error) {
	result := r.db.Create(user)
	if result.Error != nil {
		return 0, result.Error
	}
	return user.ID, nil
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserRepository) Update(user *models.User) error {
	result := r.db.Model(user).Select("name").Updates(user)
	return result.Error
}

func (r *UserRepository) Delete(id int64) error {
	result := r.db.Delete(&models.User{}, id)
	return result.Error
}
