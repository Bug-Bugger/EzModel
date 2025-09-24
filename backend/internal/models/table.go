package models

import (
	"time"

	"github.com/google/uuid"
)

// Table represents a database table in the schema
type Table struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null" json:"project_id"`
	Name      string    `gorm:"not null" json:"name"`
	PosX      float64   `json:"pos_x"` // Canvas position
	PosY      float64   `json:"pos_y"` // Canvas position
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Fields []Field `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE" json:"fields,omitempty"`
}
