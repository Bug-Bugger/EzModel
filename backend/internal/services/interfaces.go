package services

import (
	"time"

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
	UpdatePassword(id uuid.UUID, currentPassword, newPassword string) error
	DeleteUser(id uuid.UUID) error
	AuthenticateUser(email, password string) (*models.User, error)
}

type ProjectServiceInterface interface {
	CreateProject(name, description string, ownerID uuid.UUID) (*models.Project, error)
	GetProjectByID(id uuid.UUID) (*models.Project, error)
	GetProjectsByOwnerID(ownerID uuid.UUID) ([]*models.Project, error)
	GetProjectsByCollaboratorID(collaboratorID uuid.UUID) ([]*models.Project, error)
	GetAllProjects() ([]*models.Project, error)
	UpdateProject(id uuid.UUID, req *dto.UpdateProjectRequest, userID uuid.UUID) (*models.Project, error)
	DeleteProject(id uuid.UUID) error
	AddCollaborator(projectID, collaboratorID uuid.UUID) error
	RemoveCollaborator(projectID, collaboratorID uuid.UUID) error
}

type TableServiceInterface interface {
	CreateTable(projectID uuid.UUID, name string, posX, posY float64, userID uuid.UUID) (*models.Table, error)
	GetTableByID(id uuid.UUID) (*models.Table, error)
	GetTablesByProjectID(projectID uuid.UUID) ([]*models.Table, error)
	UpdateTable(id uuid.UUID, req *dto.UpdateTableRequest, userID uuid.UUID) (*models.Table, error)
	UpdateTablePosition(id uuid.UUID, posX, posY float64, userID uuid.UUID) error
	DeleteTable(id uuid.UUID, userID uuid.UUID) error
}

type FieldServiceInterface interface {
	CreateField(tableID uuid.UUID, req *dto.CreateFieldRequest, userID uuid.UUID) (*models.Field, error)
	GetFieldByID(id uuid.UUID) (*models.Field, error)
	GetFieldsByTableID(tableID uuid.UUID) ([]*models.Field, error)
	UpdateField(id uuid.UUID, req *dto.UpdateFieldRequest, userID uuid.UUID) (*models.Field, error)
	DeleteField(id uuid.UUID, userID uuid.UUID) error
	ReorderFields(tableID uuid.UUID, fieldPositions map[uuid.UUID]int) error
}

type RelationshipServiceInterface interface {
	CreateRelationship(projectID uuid.UUID, req *dto.CreateRelationshipRequest, userID uuid.UUID) (*models.Relationship, error)
	GetRelationshipByID(id uuid.UUID) (*models.Relationship, error)
	GetRelationshipsByProjectID(projectID uuid.UUID) ([]*models.Relationship, error)
	GetRelationshipsByTableID(tableID uuid.UUID) ([]*models.Relationship, error)
	UpdateRelationship(id uuid.UUID, req *dto.UpdateRelationshipRequest, userID uuid.UUID) (*models.Relationship, error)
	DeleteRelationship(id uuid.UUID, userID uuid.UUID) error
}

type CollaborationSessionServiceInterface interface {
	CreateSession(projectID, userID uuid.UUID, userColor string) (*models.CollaborationSession, error)
	GetSessionByID(id uuid.UUID) (*models.CollaborationSession, error)
	GetSessionsByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error)
	GetActiveSessionsByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error)
	UpdateCursor(sessionID uuid.UUID, cursorX, cursorY *float64) error
	UpdateSession(id uuid.UUID, req *dto.UpdateSessionRequest) (*models.CollaborationSession, error)
	SetSessionInactive(sessionID uuid.UUID) error
	DeleteSession(sessionID uuid.UUID, userID uuid.UUID) error

	// Field collaboration methods
	NotifyFieldCreated(projectID uuid.UUID, field *models.Field, senderUserID uuid.UUID) error
	NotifyFieldUpdated(projectID uuid.UUID, field *models.Field, senderUserID uuid.UUID) error
	NotifyFieldDeleted(projectID, tableID, fieldID uuid.UUID, fieldName string, senderUserID uuid.UUID) error

	// Table collaboration methods
	NotifyTableCreated(projectID uuid.UUID, table *models.Table, senderUserID uuid.UUID) error
	NotifyTableUpdated(projectID uuid.UUID, table *models.Table, senderUserID uuid.UUID) error
	NotifyTableDeleted(projectID, tableID uuid.UUID, tableName string, senderUserID uuid.UUID) error

	// Relationship collaboration methods
	NotifyRelationshipCreated(projectID uuid.UUID, relationship *models.Relationship, senderUserID uuid.UUID) error
	NotifyRelationshipUpdated(projectID uuid.UUID, relationship *models.Relationship, senderUserID uuid.UUID) error
	NotifyRelationshipDeleted(projectID, relationshipID uuid.UUID, senderUserID uuid.UUID) error

	// Canvas collaboration methods
	BroadcastCanvasUpdate(projectID uuid.UUID, canvasData string, senderUserID uuid.UUID) error
}

type JWTServiceInterface interface {
	GenerateTokenPair(user *models.User) (*TokenPair, error)
	RefreshTokens(refreshToken string) (*TokenPair, error)
	GetAccessTokenExpiration() time.Duration
	ValidateToken(tokenString string) (*CustomClaims, error)
}
