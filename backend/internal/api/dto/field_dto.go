package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateFieldRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255"`
	DataType     string `json:"data_type" validate:"required"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsNullable   bool   `json:"is_nullable"`
	DefaultValue string `json:"default_value"`
	Position     int    `json:"position"`
}

type UpdateFieldRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	DataType     *string `json:"data_type,omitempty"`
	IsPrimaryKey *bool   `json:"is_primary_key,omitempty"`
	IsNullable   *bool   `json:"is_nullable,omitempty"`
	DefaultValue *string `json:"default_value,omitempty"`
	Position     *int    `json:"position,omitempty"`
}

type ReorderFieldsRequest struct {
	FieldPositions map[uuid.UUID]int `json:"field_positions" validate:"required"`
}

type FieldResponse struct {
	ID           uuid.UUID `json:"id"`
	TableID      uuid.UUID `json:"table_id"`
	Name         string    `json:"name"`
	DataType     string    `json:"data_type"`
	IsPrimaryKey bool      `json:"is_primary_key"`
	IsNullable   bool      `json:"is_nullable"`
	DefaultValue string    `json:"default_value"`
	Position     int       `json:"position"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}