package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID            uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"type:varchar(255);not null" validate:"required,min=1,max=255"`
	Description   string    `json:"description" gorm:"type:text" validate:"max=1000"`
	OwnerID       uuid.UUID `json:"owner_id" gorm:"type:uuid;not null"`
	Owner         User      `json:"owner" gorm:"foreignKey:OwnerID"`
	Collaborators []User    `json:"collaborators,omitempty" gorm:"many2many:project_collaborators;"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime;type:timestamp with time zone"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime;type:timestamp with time zone"`
}
