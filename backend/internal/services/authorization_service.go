package services

import (
	"errors"

	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthorizationServiceInterface interface {
	CanUserAccessProject(userID, projectID uuid.UUID) (bool, error)
	CanUserModifyProject(userID, projectID uuid.UUID) (bool, error)
	CanUserDeleteCollaborationSession(userID, sessionID uuid.UUID) (bool, error)
	GetProjectIDFromTable(tableID uuid.UUID) (uuid.UUID, error)
	GetProjectIDFromRelationship(relationshipID uuid.UUID) (uuid.UUID, error)
	GetProjectIDFromField(fieldID uuid.UUID) (uuid.UUID, error)
}

type AuthorizationService struct {
	projectRepo      repository.ProjectRepositoryInterface
	tableRepo        repository.TableRepositoryInterface
	fieldRepo        repository.FieldRepositoryInterface
	relationshipRepo repository.RelationshipRepositoryInterface
	sessionRepo      repository.CollaborationSessionRepositoryInterface
}

func NewAuthorizationService(
	projectRepo repository.ProjectRepositoryInterface,
	tableRepo repository.TableRepositoryInterface,
	fieldRepo repository.FieldRepositoryInterface,
	relationshipRepo repository.RelationshipRepositoryInterface,
	sessionRepo repository.CollaborationSessionRepositoryInterface,
) *AuthorizationService {
	return &AuthorizationService{
		projectRepo:      projectRepo,
		tableRepo:        tableRepo,
		fieldRepo:        fieldRepo,
		relationshipRepo: relationshipRepo,
		sessionRepo:      sessionRepo,
	}
}

func (s *AuthorizationService) CanUserAccessProject(userID, projectID uuid.UUID) (bool, error) {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrProjectNotFound
		}
		return false, err
	}

	// Check if user is owner
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user is collaborator
	collaboratorProjects, err := s.projectRepo.GetByCollaboratorID(userID)
	if err != nil {
		return false, err
	}

	for _, collabProject := range collaboratorProjects {
		if collabProject.ID == projectID {
			return true, nil
		}
	}

	return false, nil
}

func (s *AuthorizationService) CanUserModifyProject(userID, projectID uuid.UUID) (bool, error) {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrProjectNotFound
		}
		return false, err
	}

	// Check if user is owner
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user is collaborator
	collaboratorProjects, err := s.projectRepo.GetByCollaboratorID(userID)
	if err != nil {
		return false, err
	}

	for _, collabProject := range collaboratorProjects {
		if collabProject.ID == projectID {
			return true, nil
		}
	}

	return false, nil
}

func (s *AuthorizationService) CanUserDeleteCollaborationSession(userID, sessionID uuid.UUID) (bool, error) {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrSessionNotFound
		}
		return false, err
	}

	// User can delete their own session
	if session.UserID == userID {
		return true, nil
	}

	// Or if user is project owner
	canModify, err := s.CanUserModifyProject(userID, session.ProjectID)
	if err != nil {
		return false, err
	}

	return canModify, nil
}

func (s *AuthorizationService) GetProjectIDFromTable(tableID uuid.UUID) (uuid.UUID, error) {
	table, err := s.tableRepo.GetByID(tableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, ErrTableNotFound
		}
		return uuid.Nil, err
	}
	return table.ProjectID, nil
}

func (s *AuthorizationService) GetProjectIDFromRelationship(relationshipID uuid.UUID) (uuid.UUID, error) {
	relationship, err := s.relationshipRepo.GetByID(relationshipID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, ErrRelationshipNotFound
		}
		return uuid.Nil, err
	}
	return relationship.ProjectID, nil
}

func (s *AuthorizationService) GetProjectIDFromField(fieldID uuid.UUID) (uuid.UUID, error) {
	field, err := s.fieldRepo.GetByID(fieldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, ErrFieldNotFound
		}
		return uuid.Nil, err
	}

	// Get table to find project ID
	return s.GetProjectIDFromTable(field.TableID)
}
