package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	websocketPkg "github.com/Bug-Bugger/ezmodel/internal/websocket"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CollaborationSessionService struct {
	sessionRepo repository.CollaborationSessionRepositoryInterface
	projectRepo repository.ProjectRepositoryInterface
	userRepo    repository.UserRepositoryInterface
	authService AuthorizationServiceInterface
	hub         *websocketPkg.Hub
}

func NewCollaborationSessionService(
	sessionRepo repository.CollaborationSessionRepositoryInterface,
	projectRepo repository.ProjectRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	authService AuthorizationServiceInterface,
	hub *websocketPkg.Hub,
) *CollaborationSessionService {
	return &CollaborationSessionService{
		sessionRepo: sessionRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
		authService: authService,
		hub:         hub,
	}
}

func (s *CollaborationSessionService) CreateSession(projectID, userID uuid.UUID, userColor string) (*models.CollaborationSession, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Verify user exists
	_, err = s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if userColor == "" {
		userColor = "#3b82f6" // Default blue color
	}

	now := time.Now()
	session := &models.CollaborationSession{
		ProjectID:  projectID,
		UserID:     userID,
		UserColor:  userColor,
		IsActive:   true,
		LastPingAt: now,
		JoinedAt:   now,
	}

	id, err := s.sessionRepo.Create(session)
	if err != nil {
		return nil, err
	}

	session.ID = id
	return session, nil
}

func (s *CollaborationSessionService) GetSessionByID(id uuid.UUID) (*models.CollaborationSession, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return session, nil
}

func (s *CollaborationSessionService) GetSessionsByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error) {
	return s.sessionRepo.GetByProjectID(projectID)
}

func (s *CollaborationSessionService) GetActiveSessionsByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error) {
	return s.sessionRepo.GetActiveByProjectID(projectID)
}

func (s *CollaborationSessionService) UpdateCursor(sessionID uuid.UUID, cursorX, cursorY *float64) error {
	// Verify session exists
	_, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return s.sessionRepo.UpdateCursor(sessionID, cursorX, cursorY)
}

func (s *CollaborationSessionService) UpdateSession(id uuid.UUID, req *dto.UpdateSessionRequest) (*models.CollaborationSession, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	// Only update fields that were provided
	if req.CursorX != nil {
		session.CursorX = req.CursorX
	}

	if req.CursorY != nil {
		session.CursorY = req.CursorY
	}

	if req.UserColor != nil {
		session.UserColor = *req.UserColor
	}

	if req.IsActive != nil {
		session.IsActive = *req.IsActive
		if !*req.IsActive {
			now := time.Now()
			session.LeftAt = &now
		}
	}

	// Update last ping time
	session.LastPingAt = time.Now()

	if err := s.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *CollaborationSessionService) SetSessionInactive(sessionID uuid.UUID) error {
	// Verify session exists
	_, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return s.sessionRepo.SetInactive(sessionID)
}

func (s *CollaborationSessionService) DeleteSession(sessionID uuid.UUID, userID uuid.UUID) error {
	// Check authorization first
	canDelete, err := s.authService.CanUserDeleteCollaborationSession(userID, sessionID)
	if err != nil {
		return err
	}
	if !canDelete {
		return ErrForbidden
	}

	// Verify session exists
	_, err = s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return s.sessionRepo.Delete(sessionID)
}

// WebSocket integration methods

// BroadcastSchemaChange broadcasts schema changes to all collaborators
func (s *CollaborationSessionService) BroadcastSchemaChange(projectID uuid.UUID, messageType websocketPkg.MessageType, payload interface{}, senderUserID uuid.UUID) error {
	if s.hub == nil {
		return fmt.Errorf("WebSocket hub not initialized")
	}

	message, err := websocketPkg.NewWebSocketMessage(messageType, payload, senderUserID, projectID)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket message: %w", err)
	}

	// Broadcast to all clients in the project
	s.hub.BroadcastToProject(projectID, message, nil)
	return nil
}

// BroadcastCanvasUpdate broadcasts canvas updates to all collaborators
func (s *CollaborationSessionService) BroadcastCanvasUpdate(projectID uuid.UUID, canvasData string, senderUserID uuid.UUID) error {
	payload := websocketPkg.CanvasUpdatedPayload{
		CanvasData: canvasData,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeCanvasUpdated, payload, senderUserID)
}

// NotifyTableCreated notifies collaborators about a new table
func (s *CollaborationSessionService) NotifyTableCreated(projectID uuid.UUID, table *models.Table, senderUserID uuid.UUID) error {
	payload := websocketPkg.TablePayload{
		TableID: table.ID,
		Name:    table.Name,
		X:       table.PosX,
		Y:       table.PosY,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeTableCreated, payload, senderUserID)
}

// NotifyTableUpdated notifies collaborators about a table update
func (s *CollaborationSessionService) NotifyTableUpdated(projectID uuid.UUID, table *models.Table, senderUserID uuid.UUID) error {
	payload := websocketPkg.TablePayload{
		TableID: table.ID,
		Name:    table.Name,
		X:       table.PosX,
		Y:       table.PosY,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeTableUpdated, payload, senderUserID)
}

// NotifyTableDeleted notifies collaborators about a table deletion
func (s *CollaborationSessionService) NotifyTableDeleted(projectID, tableID uuid.UUID, senderUserID uuid.UUID) error {
	payload := websocketPkg.TablePayload{
		TableID: tableID,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeTableDeleted, payload, senderUserID)
}

// NotifyFieldCreated notifies collaborators about a new field
func (s *CollaborationSessionService) NotifyFieldCreated(projectID uuid.UUID, field *models.Field, senderUserID uuid.UUID) error {
	payload := websocketPkg.FieldPayload{
		FieldID:    field.ID,
		TableID:    field.TableID,
		Name:       field.Name,
		Type:       field.DataType,
		IsPrimary:  field.IsPrimaryKey,
		IsNullable: field.IsNullable,
		Default:    &field.DefaultValue,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeFieldCreated, payload, senderUserID)
}

// NotifyFieldUpdated notifies collaborators about a field update
func (s *CollaborationSessionService) NotifyFieldUpdated(projectID uuid.UUID, field *models.Field, senderUserID uuid.UUID) error {
	payload := websocketPkg.FieldPayload{
		FieldID:    field.ID,
		TableID:    field.TableID,
		Name:       field.Name,
		Type:       field.DataType,
		IsPrimary:  field.IsPrimaryKey,
		IsNullable: field.IsNullable,
		Default:    &field.DefaultValue,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeFieldUpdated, payload, senderUserID)
}

// NotifyFieldDeleted notifies collaborators about a field deletion
func (s *CollaborationSessionService) NotifyFieldDeleted(projectID, tableID, fieldID uuid.UUID, senderUserID uuid.UUID) error {
	payload := websocketPkg.FieldPayload{
		FieldID: fieldID,
		TableID: tableID,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeFieldDeleted, payload, senderUserID)
}

// NotifyRelationshipCreated notifies collaborators about a new relationship
func (s *CollaborationSessionService) NotifyRelationshipCreated(projectID uuid.UUID, relationship *models.Relationship, senderUserID uuid.UUID) error {
	payload := websocketPkg.RelationshipPayload{
		RelationshipID: relationship.ID,
		SourceTableID:  relationship.SourceTableID,
		TargetTableID:  relationship.TargetTableID,
		SourceFieldID:  relationship.SourceFieldID,
		TargetFieldID:  relationship.TargetFieldID,
		Type:           relationship.RelationType,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeRelationshipCreated, payload, senderUserID)
}

// NotifyRelationshipUpdated notifies collaborators about a relationship update
func (s *CollaborationSessionService) NotifyRelationshipUpdated(projectID uuid.UUID, relationship *models.Relationship, senderUserID uuid.UUID) error {
	payload := websocketPkg.RelationshipPayload{
		RelationshipID: relationship.ID,
		SourceTableID:  relationship.SourceTableID,
		TargetTableID:  relationship.TargetTableID,
		SourceFieldID:  relationship.SourceFieldID,
		TargetFieldID:  relationship.TargetFieldID,
		Type:           relationship.RelationType,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeRelationshipUpdated, payload, senderUserID)
}

// NotifyRelationshipDeleted notifies collaborators about a relationship deletion
func (s *CollaborationSessionService) NotifyRelationshipDeleted(projectID, relationshipID uuid.UUID, senderUserID uuid.UUID) error {
	payload := websocketPkg.RelationshipPayload{
		RelationshipID: relationshipID,
	}

	return s.BroadcastSchemaChange(projectID, websocketPkg.MessageTypeRelationshipDeleted, payload, senderUserID)
}

// GetActiveClientCount returns the number of active clients for a project
func (s *CollaborationSessionService) GetActiveClientCount(projectID uuid.UUID) int {
	if s.hub == nil {
		return 0
	}
	return s.hub.GetActiveClients(projectID)
}

// GetActiveUsers returns active users for a project from the WebSocket hub
func (s *CollaborationSessionService) GetActiveUsers(projectID uuid.UUID) []websocketPkg.ActiveUser {
	if s.hub == nil {
		return []websocketPkg.ActiveUser{}
	}
	return s.hub.GetActiveUsers(projectID)
}
