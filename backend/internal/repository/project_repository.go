package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepositoryInterface {
	return &ProjectRepository{db: db}
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
	err := r.db.Preload("Owner").Preload("Collaborators").Preload("Tables.Fields").Preload("Relationships").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Preload("Owner").Preload("Collaborators").Where("owner_id = ?", ownerID).Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) GetByCollaboratorID(collaboratorID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Preload("Owner").Preload("Collaborators").
		Joins("JOIN project_collaborators ON projects.id = project_collaborators.project_id").
		Where("project_collaborators.user_id = ?", collaboratorID).
		Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) GetAll() ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Preload("Owner").Preload("Collaborators").Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) Delete(id uuid.UUID) error {
	// This will also delete the many-to-many relationships due to foreign key constraints
	return r.db.Delete(&models.Project{}, "id = ?", id).Error
}

func (r *ProjectRepository) AddCollaborator(projectID, userID uuid.UUID) error {
	var project models.Project
	if err := r.db.First(&project, "id = ?", projectID).Error; err != nil {
		return err
	}

	var user models.User
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	return r.db.Model(&project).Association("Collaborators").Append(&user)
}

func (r *ProjectRepository) RemoveCollaborator(projectID, userID uuid.UUID) error {
	var project models.Project
	if err := r.db.First(&project, "id = ?", projectID).Error; err != nil {
		return err
	}

	var user models.User
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	return r.db.Model(&project).Association("Collaborators").Delete(&user)
}
