package models

import (
	"time"

	"github.com/google/uuid"
)

// Field represents a column in a database table
type Field struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TableID      uuid.UUID `gorm:"type:uuid;not null" json:"table_id"`
	Name         string    `gorm:"not null" json:"name"`
	DataType     string    `gorm:"not null" json:"data_type"` // VARCHAR, INT, TEXT, etc.
	IsPrimaryKey bool      `gorm:"default:false" json:"is_primary_key"`
	IsNullable   bool      `gorm:"default:true" json:"is_nullable"`
	DefaultValue string    `json:"default_value"`
	Position     int       `json:"position"` // Field order in table
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}