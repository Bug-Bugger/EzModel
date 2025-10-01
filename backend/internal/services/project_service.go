package services

import (
	"errors"
	"log"
	"strings"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectService struct {
	projectRepo          repository.ProjectRepositoryInterface
	userRepo             repository.UserRepositoryInterface
	collaborationService CollaborationSessionServiceInterface
}

func NewProjectService(projectRepo repository.ProjectRepositoryInterface, userRepo repository.UserRepositoryInterface, collaborationService CollaborationSessionServiceInterface) *ProjectService {
	return &ProjectService{
		projectRepo:          projectRepo,
		userRepo:             userRepo,
		collaborationService: collaborationService,
	}
}

func (s *ProjectService) CreateProject(name, description string, ownerID uuid.UUID) (*models.Project, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if len(name) < 1 || len(name) > 255 {
		return nil, ErrInvalidInput
	}

	if len(description) > 1000 {
		return nil, ErrInvalidInput
	}

	// Verify owner exists
	_, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	project := &models.Project{
		Name:         name,
		Description:  description,
		OwnerID:      ownerID,
		DatabaseType: "postgresql", // Default to PostgreSQL
		CanvasData:   "{}",         // Initialize with empty JSON object
	}

	id, err := s.projectRepo.Create(project)
	if err != nil {
		return nil, err
	}

	project.ID = id
	return project, nil
}

func (s *ProjectService) GetProjectByID(id uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) GetProjectsByOwnerID(ownerID uuid.UUID) ([]*models.Project, error) {
	return s.projectRepo.GetByOwnerID(ownerID)
}

func (s *ProjectService) GetProjectsByCollaboratorID(collaboratorID uuid.UUID) ([]*models.Project, error) {
	return s.projectRepo.GetByCollaboratorID(collaboratorID)
}

func (s *ProjectService) GetAllProjects() ([]*models.Project, error) {
	return s.projectRepo.GetAll()
}

func (s *ProjectService) UpdateProject(id uuid.UUID, req *dto.UpdateProjectRequest, userID uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Only update fields that were provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 255 {
			return nil, ErrInvalidInput
		}
		project.Name = name
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if len(description) > 1000 {
			return nil, ErrInvalidInput
		}
		project.Description = description
	}

	if req.CanvasData != nil {
		canvasData := strings.TrimSpace(*req.CanvasData)
		// Validate that it's valid JSON (basic check)
		if canvasData == "" {
			canvasData = "{}"
		}
		// Note: For production, you might want to use json.Valid() for proper validation
		project.CanvasData = canvasData

		// Debug logging for canvas data updates
		log.Printf("CANVAS DEBUG: Updating canvas data for project %s, data length: %d",
			project.ID.String(), len(canvasData))

		// Broadcast canvas update to collaborators FIRST if canvas data was changed
		if s.collaborationService != nil {
			if err := s.collaborationService.BroadcastCanvasUpdate(id, project.CanvasData, userID); err != nil {
				// Log error but don't fail the operation
				// TODO: Add proper logging
			}
		}
	}

	// Then persist to database
	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(id uuid.UUID) error {
	_, err := s.projectRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	return s.projectRepo.Delete(id)
}

func (s *ProjectService) AddCollaborator(projectID, collaboratorID uuid.UUID) error {
	// Verify project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	// Verify collaborator exists
	_, err = s.userRepo.GetByID(collaboratorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaboratorNotFound
		}
		return err
	}

	return s.projectRepo.AddCollaborator(projectID, collaboratorID)
}

func (s *ProjectService) RemoveCollaborator(projectID, collaboratorID uuid.UUID) error {
	// Verify project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	// Verify collaborator exists
	_, err = s.userRepo.GetByID(collaboratorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaboratorNotFound
		}
		return err
	}

	return s.projectRepo.RemoveCollaborator(projectID, collaboratorID)
}
