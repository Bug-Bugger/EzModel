package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTableRequest struct {
	Name string  `json:"name" validate:"required,min=1,max=255"`
	PosX float64 `json:"pos_x"`
	PosY float64 `json:"pos_y"`
}

type UpdateTableRequest struct {
	Name *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	PosX *float64 `json:"pos_x,omitempty"`
	PosY *float64 `json:"pos_y,omitempty"`
}

type UpdateTablePositionRequest struct {
	PosX float64 `json:"pos_x" validate:"required"`
	PosY float64 `json:"pos_y" validate:"required"`
}

type TableResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
	PosX      float64   `json:"pos_x"`
	PosY      float64   `json:"pos_y"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TableWithFieldsResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProjectID uuid.UUID       `json:"project_id"`
	Name      string          `json:"name"`
	PosX      float64         `json:"pos_x"`
	PosY      float64         `json:"pos_y"`
	Fields    []FieldResponse `json:"fields,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
