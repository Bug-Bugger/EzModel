package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) Create(project *models.Project) (uuid.UUID, error) {
	result := r.db.Create(project)
	if result.Error != nil {
		return uuid.Nil, result.Error
	}
	return project.ID, nil
}

func (r *ProjectRepository) GetByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	result := r.db.First(&project, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &project, nil
}

func (r *ProjectRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project
	result := r.db.Where("owner_id = ?", ownerID).Find(&projects)
	if result.Error != nil {
		return nil, result.Error
	}
	return projects, nil
}

func (r *ProjectRepository) GetAll() ([]*models.Project, error) {
	var projects []*models.Project
	result := r.db.Find(&projects)
	if result.Error != nil {
		return nil, result.Error
	}
	return projects, nil
}

func (r *ProjectRepository) Update(project *models.Project) error {
	result := r.db.Save(project)
	return result.Error
}

func (r *ProjectRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.Project{}, "id = ?", id)
	return result.Error
}

func (r *ProjectRepository) GetWithOwner(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	result := r.db.Preload("Owner").First(&project, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &project, nil
}

func (r *ProjectRepository) GetAllWithOwner() ([]*models.Project, error) {
	var projects []*models.Project
	result := r.db.Preload("Owner").Find(&projects)
	if result.Error != nil {
		return nil, result.Error
	}
	return projects, nil
}
