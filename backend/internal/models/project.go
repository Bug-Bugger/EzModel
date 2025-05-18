package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null" validate:"required,min=1,max=255"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	OwnerID     uuid.UUID `json:"owner_id" gorm:"type:uuid;not null" validate:"required"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;type:timestamp with time zone"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime;type:timestamp with time zone"`

	Owner *User `json:"owner,omitempty" gorm:"foreignKey:OwnerID;references:ID"`
}
