package services

import (
	"errors"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CollaborationSessionService struct {
	sessionRepo repository.CollaborationSessionRepositoryInterface
	projectRepo repository.ProjectRepositoryInterface
	userRepo    repository.UserRepositoryInterface
}

func NewCollaborationSessionService(
	sessionRepo repository.CollaborationSessionRepositoryInterface,
	projectRepo repository.ProjectRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
) *CollaborationSessionService {
	return &CollaborationSessionService{
		sessionRepo: sessionRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
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

func (s *CollaborationSessionService) DeleteSession(sessionID uuid.UUID) error {
	_, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return s.sessionRepo.Delete(sessionID)
}