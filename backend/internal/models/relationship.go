package models

import (
	"time"

	"github.com/google/uuid"
)

// Relationship represents a foreign key relationship between tables
type Relationship struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null" json:"project_id"`
	SourceTableID uuid.UUID `gorm:"type:uuid;not null" json:"source_table_id"`
	SourceFieldID uuid.UUID `gorm:"type:uuid;not null" json:"source_field_id"`
	TargetTableID uuid.UUID `gorm:"type:uuid;not null" json:"target_table_id"`
	TargetFieldID uuid.UUID `gorm:"type:uuid;not null" json:"target_field_id"`
	RelationType  string    `gorm:"default:'one_to_many'" json:"relation_type"` // one_to_one, one_to_many, many_to_many
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
