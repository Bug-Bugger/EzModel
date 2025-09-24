package repository

import (
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CollaborationSessionRepository struct {
	db *gorm.DB
}

func NewCollaborationSessionRepository(db *gorm.DB) CollaborationSessionRepositoryInterface {
	return &CollaborationSessionRepository{db: db}
}

func (r *CollaborationSessionRepository) Create(session *models.CollaborationSession) (uuid.UUID, error) {
	if err := r.db.Create(session).Error; err != nil {
		return uuid.Nil, err
	}
	return session.ID, nil
}

func (r *CollaborationSessionRepository) GetByID(id uuid.UUID) (*models.CollaborationSession, error) {
	var session models.CollaborationSession
	err := r.db.First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *CollaborationSessionRepository) GetByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error) {
	var sessions []*models.CollaborationSession
	err := r.db.Where("project_id = ?", projectID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *CollaborationSessionRepository) GetActiveByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error) {
	var sessions []*models.CollaborationSession
	err := r.db.Where("project_id = ? AND is_active = true", projectID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *CollaborationSessionRepository) GetByUserID(userID uuid.UUID) ([]*models.CollaborationSession, error) {
	var sessions []*models.CollaborationSession
	err := r.db.Where("user_id = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *CollaborationSessionRepository) Update(session *models.CollaborationSession) error {
	return r.db.Save(session).Error
}

func (r *CollaborationSessionRepository) UpdateCursor(id uuid.UUID, cursorX, cursorY *float64) error {
	updates := map[string]interface{}{
		"last_ping_at": time.Now(),
	}

	if cursorX != nil {
		updates["cursor_x"] = cursorX
	}
	if cursorY != nil {
		updates["cursor_y"] = cursorY
	}

	return r.db.Model(&models.CollaborationSession{}).Where("id = ?", id).Updates(updates).Error
}

func (r *CollaborationSessionRepository) SetInactive(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.CollaborationSession{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active": false,
		"left_at":   &now,
	}).Error
}

func (r *CollaborationSessionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.CollaborationSession{}, "id = ?", id).Error
}