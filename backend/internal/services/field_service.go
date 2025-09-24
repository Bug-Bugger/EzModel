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

type FieldService struct {
	fieldRepo repository.FieldRepositoryInterface
	tableRepo repository.TableRepositoryInterface
	authService AuthorizationServiceInterface
}

func NewFieldService(fieldRepo repository.FieldRepositoryInterface, tableRepo repository.TableRepositoryInterface, authService AuthorizationServiceInterface) *FieldService {
	return &FieldService{
		fieldRepo: fieldRepo,
		tableRepo: tableRepo,
		authService: authService,
	}
}

func (s *FieldService) CreateField(tableID uuid.UUID, req *dto.CreateFieldRequest) (*models.Field, error) {
	name := strings.TrimSpace(req.Name)
	dataType := strings.TrimSpace(req.DataType)

	if len(name) < 1 || len(name) > 255 {
		return nil, ErrInvalidInput
	}

	if len(dataType) < 1 {
		return nil, ErrInvalidInput
	}

	// Verify table exists
	_, err := s.tableRepo.GetByID(tableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}

	field := &models.Field{
		TableID:      tableID,
		Name:         name,
		DataType:     dataType,
		IsPrimaryKey: req.IsPrimaryKey,
		IsNullable:   req.IsNullable,
		DefaultValue: req.DefaultValue,
		Position:     req.Position,
	}

	id, err := s.fieldRepo.Create(field)
	if err != nil {
		return nil, err
	}

	field.ID = id
	return field, nil
}

func (s *FieldService) GetFieldByID(id uuid.UUID) (*models.Field, error) {
	field, err := s.fieldRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFieldNotFound
		}
		return nil, err
	}
	return field, nil
}

func (s *FieldService) GetFieldsByTableID(tableID uuid.UUID) ([]*models.Field, error) {
	return s.fieldRepo.GetByTableID(tableID)
}

func (s *FieldService) UpdateField(id uuid.UUID, req *dto.UpdateFieldRequest) (*models.Field, error) {
	field, err := s.fieldRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFieldNotFound
		}
		return nil, err
	}

	// Only update fields that were provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 255 {
			return nil, ErrInvalidInput
		}
		field.Name = name
	}

	if req.DataType != nil {
		dataType := strings.TrimSpace(*req.DataType)
		if len(dataType) < 1 {
			return nil, ErrInvalidInput
		}
		field.DataType = dataType
	}

	if req.IsPrimaryKey != nil {
		field.IsPrimaryKey = *req.IsPrimaryKey
	}

	if req.IsNullable != nil {
		field.IsNullable = *req.IsNullable
	}

	if req.DefaultValue != nil {
		field.DefaultValue = *req.DefaultValue
	}

	if req.Position != nil {
		field.Position = *req.Position
	}

	if err := s.fieldRepo.Update(field); err != nil {
		return nil, err
	}

	return field, nil
}

func (s *FieldService) DeleteField(id uuid.UUID, userID uuid.UUID) error {
	// Get project ID from field
	projectID, err := s.authService.GetProjectIDFromField(id)
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

	// Verify field exists
	_, err = s.fieldRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFieldNotFound
		}
		return err
	}

	return s.fieldRepo.Delete(id)
}

func (s *FieldService) ReorderFields(tableID uuid.UUID, fieldPositions map[uuid.UUID]int) error {
	// Verify table exists
	_, err := s.tableRepo.GetByID(tableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTableNotFound
		}
		return err
	}

	// Verify all fields belong to the table
	for fieldID := range fieldPositions {
		field, err := s.fieldRepo.GetByID(fieldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFieldNotFound
			}
			return err
		}
		if field.TableID != tableID {
			return ErrInvalidInput
		}
	}

	return s.fieldRepo.ReorderFields(tableID, fieldPositions)
}