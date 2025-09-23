package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateRelationshipRequest struct {
	SourceTableID uuid.UUID `json:"source_table_id" validate:"required"`
	SourceFieldID uuid.UUID `json:"source_field_id" validate:"required"`
	TargetTableID uuid.UUID `json:"target_table_id" validate:"required"`
	TargetFieldID uuid.UUID `json:"target_field_id" validate:"required"`
	RelationType  string    `json:"relation_type" validate:"oneof=one_to_one one_to_many many_to_many"`
}

type UpdateRelationshipRequest struct {
	SourceTableID *uuid.UUID `json:"source_table_id,omitempty"`
	SourceFieldID *uuid.UUID `json:"source_field_id,omitempty"`
	TargetTableID *uuid.UUID `json:"target_table_id,omitempty"`
	TargetFieldID *uuid.UUID `json:"target_field_id,omitempty"`
	RelationType  *string    `json:"relation_type,omitempty" validate:"omitempty,oneof=one_to_one one_to_many many_to_many"`
}

type RelationshipResponse struct {
	ID            uuid.UUID `json:"id"`
	ProjectID     uuid.UUID `json:"project_id"`
	SourceTableID uuid.UUID `json:"source_table_id"`
	SourceFieldID uuid.UUID `json:"source_field_id"`
	TargetTableID uuid.UUID `json:"target_table_id"`
	TargetFieldID uuid.UUID `json:"target_field_id"`
	RelationType  string    `json:"relation_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}