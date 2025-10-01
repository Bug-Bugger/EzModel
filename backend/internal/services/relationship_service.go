package services

import (
	"errors"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RelationshipService struct {
	relationshipRepo     repository.RelationshipRepositoryInterface
	projectRepo          repository.ProjectRepositoryInterface
	tableRepo            repository.TableRepositoryInterface
	fieldRepo            repository.FieldRepositoryInterface
	authService          AuthorizationServiceInterface
	collaborationService CollaborationSessionServiceInterface
}

func NewRelationshipService(
	relationshipRepo repository.RelationshipRepositoryInterface,
	projectRepo repository.ProjectRepositoryInterface,
	tableRepo repository.TableRepositoryInterface,
	fieldRepo repository.FieldRepositoryInterface,
	authService AuthorizationServiceInterface,
	collaborationService CollaborationSessionServiceInterface,
) *RelationshipService {
	return &RelationshipService{
		relationshipRepo:     relationshipRepo,
		projectRepo:          projectRepo,
		tableRepo:            tableRepo,
		fieldRepo:            fieldRepo,
		authService:          authService,
		collaborationService: collaborationService,
	}
}

func (s *RelationshipService) CreateRelationship(projectID uuid.UUID, req *dto.CreateRelationshipRequest, userID uuid.UUID) (*models.Relationship, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Verify source table exists
	_, err = s.tableRepo.GetByID(req.SourceTableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}

	// Verify target table exists
	_, err = s.tableRepo.GetByID(req.TargetTableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}

	// Verify source field exists
	_, err = s.fieldRepo.GetByID(req.SourceFieldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFieldNotFound
		}
		return nil, err
	}

	// Verify target field exists
	_, err = s.fieldRepo.GetByID(req.TargetFieldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFieldNotFound
		}
		return nil, err
	}

	relationType := req.RelationType
	if relationType == "" {
		relationType = "one_to_many"
	}

	relationship := &models.Relationship{
		ProjectID:     projectID,
		SourceTableID: req.SourceTableID,
		SourceFieldID: req.SourceFieldID,
		TargetTableID: req.TargetTableID,
		TargetFieldID: req.TargetFieldID,
		RelationType:  relationType,
	}

	// Generate UUID for the relationship before broadcasting
	relationship.ID = uuid.New()

	// Broadcast relationship creation to collaborators FIRST
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyRelationshipCreated(projectID, relationship, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	// Then persist to database
	id, err := s.relationshipRepo.Create(relationship)
	if err != nil {
		return nil, err
	}

	relationship.ID = id

	return relationship, nil
}

func (s *RelationshipService) GetRelationshipByID(id uuid.UUID) (*models.Relationship, error) {
	relationship, err := s.relationshipRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRelationshipNotFound
		}
		return nil, err
	}
	return relationship, nil
}

func (s *RelationshipService) GetRelationshipsByProjectID(projectID uuid.UUID) ([]*models.Relationship, error) {
	return s.relationshipRepo.GetByProjectID(projectID)
}

func (s *RelationshipService) GetRelationshipsByTableID(tableID uuid.UUID) ([]*models.Relationship, error) {
	return s.relationshipRepo.GetByTableID(tableID)
}

func (s *RelationshipService) UpdateRelationship(id uuid.UUID, req *dto.UpdateRelationshipRequest, userID uuid.UUID) (*models.Relationship, error) {
	relationship, err := s.relationshipRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRelationshipNotFound
		}
		return nil, err
	}

	// Only update fields that were provided
	if req.SourceTableID != nil {
		// Verify source table exists
		_, err = s.tableRepo.GetByID(*req.SourceTableID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrTableNotFound
			}
			return nil, err
		}
		relationship.SourceTableID = *req.SourceTableID
	}

	if req.TargetTableID != nil {
		// Verify target table exists
		_, err = s.tableRepo.GetByID(*req.TargetTableID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrTableNotFound
			}
			return nil, err
		}
		relationship.TargetTableID = *req.TargetTableID
	}

	if req.SourceFieldID != nil {
		// Verify source field exists
		_, err = s.fieldRepo.GetByID(*req.SourceFieldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFieldNotFound
			}
			return nil, err
		}
		relationship.SourceFieldID = *req.SourceFieldID
	}

	if req.TargetFieldID != nil {
		// Verify target field exists
		_, err = s.fieldRepo.GetByID(*req.TargetFieldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFieldNotFound
			}
			return nil, err
		}
		relationship.TargetFieldID = *req.TargetFieldID
	}

	if req.RelationType != nil {
		relationship.RelationType = *req.RelationType
	}

	// Broadcast relationship update to collaborators FIRST
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyRelationshipUpdated(relationship.ProjectID, relationship, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	// Then persist to database
	if err := s.relationshipRepo.Update(relationship); err != nil {
		return nil, err
	}

	return relationship, nil
}

func (s *RelationshipService) DeleteRelationship(id uuid.UUID, userID uuid.UUID) error {
	// Get project ID from relationship
	projectID, err := s.authService.GetProjectIDFromRelationship(id)
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

	// Verify relationship exists
	_, err = s.relationshipRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRelationshipNotFound
		}
		return err
	}

	// Notify collaborators about relationship deletion FIRST
	if s.collaborationService != nil {
		if err := s.collaborationService.NotifyRelationshipDeleted(projectID, id, userID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	// Then delete from database
	if err := s.relationshipRepo.Delete(id); err != nil {
		return err
	}

	return nil
}
