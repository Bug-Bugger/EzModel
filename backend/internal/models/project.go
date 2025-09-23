package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a database schema design project
type Project struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Description  string    `json:"description"`
	OwnerID      uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	DatabaseType string    `gorm:"default:'postgresql'" json:"database_type"` // postgresql, mysql, sqlite, sqlserver
	CanvasData   string    `gorm:"type:jsonb" json:"canvas_data"`             // Visual layout/positioning data
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Owner         User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Collaborators []User         `gorm:"many2many:project_collaborators;" json:"collaborators,omitempty"`
	Tables        []Table        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"tables,omitempty"`
	Relationships []Relationship `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"relationships,omitempty"`
}
