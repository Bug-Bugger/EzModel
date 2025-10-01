package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=1000"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	CanvasData  *string `json:"canvas_data,omitempty"`
}

type AddCollaboratorRequest struct {
	CollaboratorID uuid.UUID `json:"collaborator_id" validate:"required"`
}

type ProjectResponse struct {
	ID            uuid.UUID                 `json:"id"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	OwnerID       uuid.UUID                 `json:"owner_id"`
	DatabaseType  string                    `json:"database_type"`
	CanvasData    string                    `json:"canvas_data"`
	Owner         UserResponse              `json:"owner"`
	Collaborators []UserResponse            `json:"collaborators,omitempty"`
	Tables        []TableWithFieldsResponse `json:"tables,omitempty"`
	Relationships []RelationshipResponse    `json:"relationships,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

type ProjectSummaryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
