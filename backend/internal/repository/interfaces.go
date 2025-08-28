package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	Create(user *models.User) (uuid.UUID, error)
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
}

type ProjectRepositoryInterface interface {
	Create(project *models.Project) (uuid.UUID, error)
	GetByID(id uuid.UUID) (*models.Project, error)
	GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error)
	GetByCollaboratorID(collaboratorID uuid.UUID) ([]*models.Project, error)
	GetAll() ([]*models.Project, error)
	Update(project *models.Project) error
	Delete(id uuid.UUID) error
	AddCollaborator(projectID, userID uuid.UUID) error
	RemoveCollaborator(projectID, userID uuid.UUID) error
}
