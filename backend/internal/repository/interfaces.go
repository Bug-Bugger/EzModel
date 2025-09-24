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

type TableRepositoryInterface interface {
	Create(table *models.Table) (uuid.UUID, error)
	GetByID(id uuid.UUID) (*models.Table, error)
	GetByProjectID(projectID uuid.UUID) ([]*models.Table, error)
	Update(table *models.Table) error
	Delete(id uuid.UUID) error
	UpdatePosition(id uuid.UUID, posX, posY float64) error
}

type FieldRepositoryInterface interface {
	Create(field *models.Field) (uuid.UUID, error)
	GetByID(id uuid.UUID) (*models.Field, error)
	GetByTableID(tableID uuid.UUID) ([]*models.Field, error)
	Update(field *models.Field) error
	Delete(id uuid.UUID) error
	ReorderFields(tableID uuid.UUID, fieldPositions map[uuid.UUID]int) error
}

type RelationshipRepositoryInterface interface {
	Create(relationship *models.Relationship) (uuid.UUID, error)
	GetByID(id uuid.UUID) (*models.Relationship, error)
	GetByProjectID(projectID uuid.UUID) ([]*models.Relationship, error)
	GetByTableID(tableID uuid.UUID) ([]*models.Relationship, error)
	Update(relationship *models.Relationship) error
	Delete(id uuid.UUID) error
}

type CollaborationSessionRepositoryInterface interface {
	Create(session *models.CollaborationSession) (uuid.UUID, error)
	GetByID(id uuid.UUID) (*models.CollaborationSession, error)
	GetByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error)
	GetActiveByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error)
	GetByUserID(userID uuid.UUID) ([]*models.CollaborationSession, error)
	Update(session *models.CollaborationSession) error
	UpdateCursor(id uuid.UUID, cursorX, cursorY *float64) error
	SetInactive(id uuid.UUID) error
	Delete(id uuid.UUID) error
}
