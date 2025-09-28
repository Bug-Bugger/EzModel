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
	tableRepo   repository.TableRepositoryInterface
	projectRepo repository.ProjectRepositoryInterface
	authService AuthorizationServiceInterface
}

func NewTableService(tableRepo repository.TableRepositoryInterface, projectRepo repository.ProjectRepositoryInterface, authService AuthorizationServiceInterface) *TableService {
	return &TableService{
		tableRepo:   tableRepo,
		projectRepo: projectRepo,
		authService: authService,
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

func (s *TableService) UpdateTable(id uuid.UUID, req *dto.UpdateTableRequest) (*models.Table, error) {
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

	return table, nil
}

func (s *TableService) UpdateTablePosition(id uuid.UUID, posX, posY float64) error {
	// Verify table exists
	_, err := s.tableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTableNotFound
		}
		return err
	}

	return s.tableRepo.UpdatePosition(id, posX, posY)
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

	return s.tableRepo.Delete(id)
}
