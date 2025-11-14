package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// User presence events
	MessageTypeUserJoined   MessageType = "user_joined"
	MessageTypeUserLeft     MessageType = "user_left"
	MessageTypeUserCursor   MessageType = "user_cursor"
	MessageTypeUserPresence MessageType = "user_presence"

	// Schema modification events
	MessageTypeTableCreated MessageType = "table_created"
	MessageTypeTableUpdated MessageType = "table_updated"
	MessageTypeTableMoved   MessageType = "table_moved"
	MessageTypeTableDeleted MessageType = "table_deleted"
	MessageTypeFieldCreated MessageType = "field_created"
	MessageTypeFieldUpdated MessageType = "field_updated"
	MessageTypeFieldDeleted MessageType = "field_deleted"

	// Relationship events
	MessageTypeRelationshipCreated MessageType = "relationship_create"
	MessageTypeRelationshipUpdated MessageType = "relationship_update"
	MessageTypeRelationshipDeleted MessageType = "relationship_delete"

	// Canvas events
	MessageTypeCanvasUpdated MessageType = "canvas_updated"

	// System events
	MessageTypeAuth  MessageType = "auth"
	MessageTypeError MessageType = "error"
	MessageTypePing  MessageType = "ping"
	MessageTypePong  MessageType = "pong"
)

// WebSocketMessage represents a WebSocket message structure
type WebSocketMessage struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	UserID    uuid.UUID       `json:"user_id"`
	ProjectID uuid.UUID       `json:"project_id"`
	Timestamp time.Time       `json:"timestamp"`
}

// User presence payloads
type UserJoinedPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	UserColor string    `json:"user_color"`
}

type UserLeftPayload struct {
	UserID uuid.UUID `json:"user_id"`
}

type UserCursorPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	UserColor string    `json:"user_color"`
	CursorX   float64   `json:"cursor_x"` // Global coordinates in SvelteFlow space
	CursorY   float64   `json:"cursor_y"` // Global coordinates in SvelteFlow space
}

type UserPresencePayload struct {
	ActiveUsers []ActiveUser `json:"active_users"`
}

type ActiveUser struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	UserColor string    `json:"user_color"`
	CursorX   *float64  `json:"cursor_x,omitempty"` // Global coordinates in SvelteFlow space
	CursorY   *float64  `json:"cursor_y,omitempty"` // Global coordinates in SvelteFlow space
	LastSeen  time.Time `json:"last_seen"`
}

// Schema modification payloads
type TablePayload struct {
	TableID uuid.UUID `json:"table_id"`
	Name    string    `json:"name"`
	X       float64   `json:"x"`
	Y       float64   `json:"y"`
}

type FieldPayload struct {
	FieldID      uuid.UUID `json:"field_id"`
	TableID      uuid.UUID `json:"table_id"`
	Name         string    `json:"name"`
	DataType     string    `json:"data_type"`
	IsPrimaryKey bool      `json:"is_primary_key"`
	IsNullable   bool      `json:"is_nullable"`
	DefaultValue *string   `json:"default_value,omitempty"`
	Position     int       `json:"position"`
}

type RelationshipPayload struct {
	RelationshipID uuid.UUID `json:"relationship_id"`
	SourceTableID  uuid.UUID `json:"source_table_id"`
	TargetTableID  uuid.UUID `json:"target_table_id"`
	SourceFieldID  uuid.UUID `json:"source_field_id"`
	TargetFieldID  uuid.UUID `json:"target_field_id"`
	Type           string    `json:"relation_type"`
	FromTableName  string    `json:"from_table"`
	ToTableName    string    `json:"to_table"`
}

// Canvas payload
type CanvasUpdatedPayload struct {
	CanvasData string `json:"canvas_data"`
}

// System payloads
type AuthPayload struct {
	Token string `json:"token"`
}

type AuthSuccessPayload struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type PingPayload struct {
	Timestamp time.Time `json:"timestamp"`
}

type PongPayload struct {
	Timestamp time.Time `json:"timestamp"`
}

// Helper functions to create messages
func NewWebSocketMessage(msgType MessageType, data interface{}, userID, projectID uuid.UUID) (*WebSocketMessage, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &WebSocketMessage{
		Type:      msgType,
		Data:      dataBytes,
		UserID:    userID,
		ProjectID: projectID,
		Timestamp: time.Now(),
	}, nil
}

func (m *WebSocketMessage) UnmarshalData(target interface{}) error {
	return json.Unmarshal(m.Data, target)
}
