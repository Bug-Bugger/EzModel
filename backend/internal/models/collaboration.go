package models

import (
	"time"

	"github.com/google/uuid"
)

// CollaborationSession tracks active real-time collaboration sessions
type CollaborationSession struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID  `gorm:"type:uuid;not null" json:"project_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CursorX    *float64   `json:"cursor_x"`
	CursorY    *float64   `json:"cursor_y"`
	UserColor  string     `json:"user_color"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	LastPingAt time.Time  `json:"last_ping_at"`
	JoinedAt   time.Time  `json:"joined_at"`
	LeftAt     *time.Time `json:"left_at"`
}
