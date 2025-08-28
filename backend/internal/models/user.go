package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email              string    `json:"email" gorm:"type:varchar(255);unique;not null" validate:"required,email"`
	PasswordHash       string    `json:"-" gorm:"type:varchar(255);not null" validate:"required"`
	Username           string    `json:"username" gorm:"type:varchar(100);not null" validate:"required,min=3,max=100"`
	OwnedProjects      []Project `json:"owned_projects,omitempty" gorm:"foreignKey:OwnerID"`
	CollaboratedProjects []Project `json:"collaborated_projects,omitempty" gorm:"many2many:project_collaborators;"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime;type:timestamp with time zone"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime;type:timestamp with time zone"`
}
