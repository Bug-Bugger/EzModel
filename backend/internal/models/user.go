package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email         string     `json:"email" gorm:"type:varchar(255);unique;not null" validate:"required,email"`
	PasswordHash  string     `json:"-" gorm:"type:varchar(255);not null" validate:"required"`
	Username      string     `json:"username" gorm:"type:varchar(100);not null" validate:"required,min=3,max=100"`
	AvatarURL     string     `json:"avatar_url,omitempty" gorm:"type:varchar(500)"`
	EmailVerified bool       `json:"email_verified" gorm:"default:false"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" gorm:"type:timestamp with time zone"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime;type:timestamp with time zone"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime;type:timestamp with time zone"`
}
