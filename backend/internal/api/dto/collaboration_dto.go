package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateSessionRequest struct {
	UserColor string `json:"user_color"`
}

type UpdateSessionRequest struct {
	CursorX   *float64 `json:"cursor_x,omitempty"`
	CursorY   *float64 `json:"cursor_y,omitempty"`
	UserColor *string  `json:"user_color,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

type UpdateCursorRequest struct {
	CursorX *float64 `json:"cursor_x"`
	CursorY *float64 `json:"cursor_y"`
}

type CollaborationSessionResponse struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	UserID     uuid.UUID  `json:"user_id"`
	CursorX    *float64   `json:"cursor_x"`
	CursorY    *float64   `json:"cursor_y"`
	UserColor  string     `json:"user_color"`
	IsActive   bool       `json:"is_active"`
	LastPingAt time.Time  `json:"last_ping_at"`
	JoinedAt   time.Time  `json:"joined_at"`
	LeftAt     *time.Time `json:"left_at"`
}
