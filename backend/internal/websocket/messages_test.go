package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MessageTestSuite struct {
	suite.Suite
}

func TestMessageSuite(t *testing.T) {
	suite.Run(t, new(MessageTestSuite))
}

// Test NewWebSocketMessage Creation
func (suite *MessageTestSuite) TestNewWebSocketMessage() {
	userID := uuid.New()
	projectID := uuid.New()
	payload := UserCursorPayload{
		UserID:    userID,
		Username:  "testuser",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}

	message, err := NewWebSocketMessage(MessageTypeUserCursor, payload, userID, projectID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), message)
	assert.Equal(suite.T(), MessageTypeUserCursor, message.Type)
	assert.Equal(suite.T(), userID, message.UserID)
	assert.Equal(suite.T(), projectID, message.ProjectID)
	assert.NotEmpty(suite.T(), message.Data)
	assert.WithinDuration(suite.T(), time.Now(), message.Timestamp, time.Second)
}

// Test Message Marshaling and Unmarshaling
func (suite *MessageTestSuite) TestMessageMarshaling() {
	userID := uuid.New()
	projectID := uuid.New()
	payload := UserCursorPayload{
		UserID:    userID,
		Username:  "testuser",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}

	// Create message
	message, err := NewWebSocketMessage(MessageTypeUserCursor, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	// Marshal to JSON
	jsonData, err := json.Marshal(message)
	assert.NoError(suite.T(), err)

	// Unmarshal back
	var unmarshaledMessage WebSocketMessage
	err = json.Unmarshal(jsonData, &unmarshaledMessage)
	assert.NoError(suite.T(), err)

	// Verify fields
	assert.Equal(suite.T(), message.Type, unmarshaledMessage.Type)
	assert.Equal(suite.T(), message.UserID, unmarshaledMessage.UserID)
	assert.Equal(suite.T(), message.ProjectID, unmarshaledMessage.ProjectID)
	assert.Equal(suite.T(), message.Data, unmarshaledMessage.Data)
}

// Test UnmarshalData Method
func (suite *MessageTestSuite) TestUnmarshalData() {
	originalPayload := UserCursorPayload{
		UserID:    uuid.New(),
		Username:  "testuser",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}

	message, err := NewWebSocketMessage(MessageTypeUserCursor, originalPayload, uuid.New(), uuid.New())
	assert.NoError(suite.T(), err)

	// Unmarshal data back to payload
	var unmarshaledPayload UserCursorPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	// Verify payload fields
	assert.Equal(suite.T(), originalPayload.UserID, unmarshaledPayload.UserID)
	assert.Equal(suite.T(), originalPayload.Username, unmarshaledPayload.Username)
	assert.Equal(suite.T(), originalPayload.UserColor, unmarshaledPayload.UserColor)
	assert.Equal(suite.T(), originalPayload.CursorX, unmarshaledPayload.CursorX)
	assert.Equal(suite.T(), originalPayload.CursorY, unmarshaledPayload.CursorY)
}

// Test UserJoinedPayload
func (suite *MessageTestSuite) TestUserJoinedPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	payload := UserJoinedPayload{
		UserID:    userID,
		Username:  "newuser",
		UserColor: "#4ECDC4",
	}

	message, err := NewWebSocketMessage(MessageTypeUserJoined, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload UserJoinedPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.UserID, unmarshaledPayload.UserID)
	assert.Equal(suite.T(), payload.Username, unmarshaledPayload.Username)
	assert.Equal(suite.T(), payload.UserColor, unmarshaledPayload.UserColor)
}

// Test UserLeftPayload
func (suite *MessageTestSuite) TestUserLeftPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	payload := UserLeftPayload{
		UserID: userID,
	}

	message, err := NewWebSocketMessage(MessageTypeUserLeft, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload UserLeftPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.UserID, unmarshaledPayload.UserID)
}

// Test TablePayload
func (suite *MessageTestSuite) TestTablePayload() {
	userID := uuid.New()
	projectID := uuid.New()
	tableID := uuid.New()
	payload := TablePayload{
		TableID:     tableID,
		Name:        "Users Table",
		Description: "User data table",
		X:           150.5,
		Y:           250.5,
	}

	message, err := NewWebSocketMessage(MessageTypeTableCreated, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload TablePayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.TableID, unmarshaledPayload.TableID)
	assert.Equal(suite.T(), payload.Name, unmarshaledPayload.Name)
	assert.Equal(suite.T(), payload.Description, unmarshaledPayload.Description)
	assert.Equal(suite.T(), payload.X, unmarshaledPayload.X)
	assert.Equal(suite.T(), payload.Y, unmarshaledPayload.Y)
}

// Test FieldPayload
func (suite *MessageTestSuite) TestFieldPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	fieldID := uuid.New()
	tableID := uuid.New()
	defaultValue := "default_value"
	payload := FieldPayload{
		FieldID:    fieldID,
		TableID:    tableID,
		Name:       "email",
		Type:       "VARCHAR(255)",
		IsPrimary:  false,
		IsNullable: false,
		IsUnique:   true,
		Default:    &defaultValue,
	}

	message, err := NewWebSocketMessage(MessageTypeFieldCreated, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload FieldPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.FieldID, unmarshaledPayload.FieldID)
	assert.Equal(suite.T(), payload.TableID, unmarshaledPayload.TableID)
	assert.Equal(suite.T(), payload.Name, unmarshaledPayload.Name)
	assert.Equal(suite.T(), payload.Type, unmarshaledPayload.Type)
	assert.Equal(suite.T(), payload.IsPrimary, unmarshaledPayload.IsPrimary)
	assert.Equal(suite.T(), payload.IsNullable, unmarshaledPayload.IsNullable)
	assert.Equal(suite.T(), payload.IsUnique, unmarshaledPayload.IsUnique)
	assert.Equal(suite.T(), *payload.Default, *unmarshaledPayload.Default)
}

// Test RelationshipPayload
func (suite *MessageTestSuite) TestRelationshipPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	relationshipID := uuid.New()
	sourceTableID := uuid.New()
	targetTableID := uuid.New()
	sourceFieldID := uuid.New()
	targetFieldID := uuid.New()

	payload := RelationshipPayload{
		RelationshipID: relationshipID,
		SourceTableID:  sourceTableID,
		TargetTableID:  targetTableID,
		SourceFieldID:  sourceFieldID,
		TargetFieldID:  targetFieldID,
		Type:           "one_to_many",
	}

	message, err := NewWebSocketMessage(MessageTypeRelationshipCreated, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload RelationshipPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.RelationshipID, unmarshaledPayload.RelationshipID)
	assert.Equal(suite.T(), payload.SourceTableID, unmarshaledPayload.SourceTableID)
	assert.Equal(suite.T(), payload.TargetTableID, unmarshaledPayload.TargetTableID)
	assert.Equal(suite.T(), payload.SourceFieldID, unmarshaledPayload.SourceFieldID)
	assert.Equal(suite.T(), payload.TargetFieldID, unmarshaledPayload.TargetFieldID)
	assert.Equal(suite.T(), payload.Type, unmarshaledPayload.Type)
}

// Test CanvasUpdatedPayload
func (suite *MessageTestSuite) TestCanvasUpdatedPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	canvasData := `{"tables": [{"id": "123", "x": 100, "y": 200}], "relationships": []}`

	payload := CanvasUpdatedPayload{
		CanvasData: canvasData,
	}

	message, err := NewWebSocketMessage(MessageTypeCanvasUpdated, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload CanvasUpdatedPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.CanvasData, unmarshaledPayload.CanvasData)
}

// Test UserPresencePayload
func (suite *MessageTestSuite) TestUserPresencePayload() {
	userID := uuid.New()
	projectID := uuid.New()
	user1ID := uuid.New()
	user2ID := uuid.New()
	cursorX := 150.0
	cursorY := 250.0

	payload := UserPresencePayload{
		ActiveUsers: []ActiveUser{
			{
				UserID:    user1ID,
				Username:  "user1",
				UserColor: "#FF6B6B",
				CursorX:   &cursorX,
				CursorY:   &cursorY,
				LastSeen:  time.Now(),
			},
			{
				UserID:    user2ID,
				Username:  "user2",
				UserColor: "#4ECDC4",
				LastSeen:  time.Now(),
			},
		},
	}

	message, err := NewWebSocketMessage(MessageTypeUserPresence, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload UserPresencePayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), unmarshaledPayload.ActiveUsers, 2)

	// Check first user
	user1 := unmarshaledPayload.ActiveUsers[0]
	assert.Equal(suite.T(), user1ID, user1.UserID)
	assert.Equal(suite.T(), "user1", user1.Username)
	assert.Equal(suite.T(), "#FF6B6B", user1.UserColor)
	assert.NotNil(suite.T(), user1.CursorX)
	assert.NotNil(suite.T(), user1.CursorY)
	assert.Equal(suite.T(), cursorX, *user1.CursorX)
	assert.Equal(suite.T(), cursorY, *user1.CursorY)

	// Check second user
	user2 := unmarshaledPayload.ActiveUsers[1]
	assert.Equal(suite.T(), user2ID, user2.UserID)
	assert.Equal(suite.T(), "user2", user2.Username)
	assert.Equal(suite.T(), "#4ECDC4", user2.UserColor)
	assert.Nil(suite.T(), user2.CursorX)
	assert.Nil(suite.T(), user2.CursorY)
}

// Test ErrorPayload
func (suite *MessageTestSuite) TestErrorPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	payload := ErrorPayload{
		Message: "Authentication failed",
		Code:    "AUTH_ERROR",
	}

	message, err := NewWebSocketMessage(MessageTypeError, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload ErrorPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), payload.Message, unmarshaledPayload.Message)
	assert.Equal(suite.T(), payload.Code, unmarshaledPayload.Code)
}

// Test PingPayload
func (suite *MessageTestSuite) TestPingPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	now := time.Now()
	payload := PingPayload{
		Timestamp: now,
	}

	message, err := NewWebSocketMessage(MessageTypePing, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload PingPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	// Time comparison with some tolerance for JSON marshaling precision
	assert.WithinDuration(suite.T(), payload.Timestamp, unmarshaledPayload.Timestamp, time.Millisecond)
}

// Test PongPayload
func (suite *MessageTestSuite) TestPongPayload() {
	userID := uuid.New()
	projectID := uuid.New()
	now := time.Now()
	payload := PongPayload{
		Timestamp: now,
	}

	message, err := NewWebSocketMessage(MessageTypePong, payload, userID, projectID)
	assert.NoError(suite.T(), err)

	var unmarshaledPayload PongPayload
	err = message.UnmarshalData(&unmarshaledPayload)
	assert.NoError(suite.T(), err)

	// Time comparison with some tolerance for JSON marshaling precision
	assert.WithinDuration(suite.T(), payload.Timestamp, unmarshaledPayload.Timestamp, time.Millisecond)
}

// Test Invalid Message Type
func (suite *MessageTestSuite) TestInvalidMessageType() {
	userID := uuid.New()
	projectID := uuid.New()
	payload := "invalid payload"

	message, err := NewWebSocketMessage("invalid_type", payload, userID, projectID)

	// Should still create message but with string payload
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), message)
	assert.Equal(suite.T(), MessageType("invalid_type"), message.Type)
}

// Test UnmarshalData with Wrong Type
func (suite *MessageTestSuite) TestUnmarshalDataWrongType() {
	// Create message with UserCursorPayload
	payload := UserCursorPayload{
		UserID:    uuid.New(),
		Username:  "testuser",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}

	message, err := NewWebSocketMessage(MessageTypeUserCursor, payload, uuid.New(), uuid.New())
	assert.NoError(suite.T(), err)

	// Try to unmarshal to wrong type
	var wrongPayload TablePayload
	err = message.UnmarshalData(&wrongPayload)

	// Should return error or have unexpected values
	assert.Error(suite.T(), err)
}

// Benchmark tests
func BenchmarkNewWebSocketMessage(b *testing.B) {
	userID := uuid.New()
	projectID := uuid.New()
	payload := UserCursorPayload{
		UserID:    userID,
		Username:  "testuser",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewWebSocketMessage(MessageTypeUserCursor, payload, userID, projectID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageUnmarshalData(b *testing.B) {
	payload := UserCursorPayload{
		UserID:    uuid.New(),
		Username:  "testuser",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}

	message, _ := NewWebSocketMessage(MessageTypeUserCursor, payload, uuid.New(), uuid.New())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var unmarshaledPayload UserCursorPayload
		err := message.UnmarshalData(&unmarshaledPayload)
		if err != nil {
			b.Fatal(err)
		}
	}
}