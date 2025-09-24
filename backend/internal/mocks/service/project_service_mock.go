package service

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) CreateProject(name, description string, ownerID uuid.UUID) (*models.Project, error) {
	args := m.Called(name, description, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectService) GetProjectByID(id uuid.UUID) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectService) GetProjectsByOwnerID(ownerID uuid.UUID) ([]*models.Project, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectService) GetProjectsByCollaboratorID(collaboratorID uuid.UUID) ([]*models.Project, error) {
	args := m.Called(collaboratorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectService) GetAllProjects() ([]*models.Project, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectService) UpdateProject(id uuid.UUID, req *dto.UpdateProjectRequest) (*models.Project, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectService) DeleteProject(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectService) AddCollaborator(projectID, collaboratorID uuid.UUID) error {
	args := m.Called(projectID, collaboratorID)
	return args.Error(0)
}

func (m *MockProjectService) RemoveCollaborator(projectID, collaboratorID uuid.UUID) error {
	args := m.Called(projectID, collaboratorID)
	return args.Error(0)
}