package services

import (
	"errors"
	"strings"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TableService struct {
	tableRepo           repository.TableRepositoryInterface
	projectRepo         repository.ProjectRepositoryInterface
	authService         AuthorizationServiceInterface
	collaborationService CollaborationSessionServiceInterface
}

func NewTableService(tableRepo repository.TableRepositoryInterface, projectRepo repository.ProjectRepositoryInterface, authService AuthorizationServiceInterface, collaborationService CollaborationSessionServiceInterface) *TableService {
	return &TableService{
		tableRepo:           tableRepo,
		projectRepo:         projectRepo,
		authService:         authService,
		collaborationService: collaborationService,
	}
}

func (s *TableService) CreateTable(projectID uuid.UUID, name string, posX, posY float64, userID uuid.UUID) (*models.Table, error) {
	name = strings.TrimSpace(name)

	if len(name) < 1 || len(name) > 255 {
		return nil, ErrInvalidInput
	}

	// Verify project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Check authorization
	canModify, err := s.authService.CanUserModifyProject(userID, projectID)
	if err != nil {
		return nil, err
	}
	if !canModify {
		return nil, ErrForbidden
	}

	table := &models.Table{
		ProjectID: projectID,
		Name:      name,
		PosX:      posX,
		PosY:      posY,
	}

	id, err := s.tableRepo.Create(table)
	if err != nil {
		return nil, err
	}

	table.ID = id

	// Broadcast table creation to collaborators
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyTableCreated(projectID, table, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	return table, nil
}

func (s *TableService) GetTableByID(id uuid.UUID) (*models.Table, error) {
	table, err := s.tableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}
	return table, nil
}

func (s *TableService) GetTablesByProjectID(projectID uuid.UUID) ([]*models.Table, error) {
	return s.tableRepo.GetByProjectID(projectID)
}

func (s *TableService) UpdateTable(id uuid.UUID, req *dto.UpdateTableRequest, userID uuid.UUID) (*models.Table, error) {
	table, err := s.tableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}

	// Only update fields that were provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 255 {
			return nil, ErrInvalidInput
		}
		table.Name = name
	}

	if req.PosX != nil {
		table.PosX = *req.PosX
	}

	if req.PosY != nil {
		table.PosY = *req.PosY
	}

	if err := s.tableRepo.Update(table); err != nil {
		return nil, err
	}

	// Broadcast table update to collaborators
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyTableUpdated(table.ProjectID, table, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	return table, nil
}

func (s *TableService) UpdateTablePosition(id uuid.UUID, posX, posY float64, userID uuid.UUID) error {
	// Verify table exists and get table data
	table, err := s.tableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTableNotFound
		}
		return err
	}

	err = s.tableRepo.UpdatePosition(id, posX, posY)
	if err != nil {
		return err
	}

	// Update table position for broadcasting
	table.PosX = posX
	table.PosY = posY

	// Broadcast table position update to collaborators
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyTableUpdated(table.ProjectID, table, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	return nil
}

func (s *TableService) DeleteTable(id uuid.UUID, userID uuid.UUID) error {
	// Get project ID from table
	projectID, err := s.authService.GetProjectIDFromTable(id)
	if err != nil {
		return err
	}

	// Check authorization
	canModify, err := s.authService.CanUserModifyProject(userID, projectID)
	if err != nil {
		return err
	}
	if !canModify {
		return ErrForbidden
	}

	// Verify table exists
	_, err = s.tableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTableNotFound
		}
		return err
	}

	// Delete the table
	if err := s.tableRepo.Delete(id); err != nil {
		return err
	}

	// Notify collaborators about table deletion
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyTableDeleted(projectID, id, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	return nil
}
