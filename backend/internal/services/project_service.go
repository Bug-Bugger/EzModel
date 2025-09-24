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

type ProjectService struct {
	projectRepo repository.ProjectRepositoryInterface
	userRepo    repository.UserRepositoryInterface
}

func NewProjectService(projectRepo repository.ProjectRepositoryInterface, userRepo repository.UserRepositoryInterface) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
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
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
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

func (s *ProjectService) UpdateProject(id uuid.UUID, req *dto.UpdateProjectRequest) (*models.Project, error) {
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
