package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(project *models.Project) (uuid.UUID, error) {
	args := m.Called(project)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockProjectRepository) GetByID(id uuid.UUID) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByCollaboratorID(collaboratorID uuid.UUID) ([]*models.Project, error) {
	args := m.Called(collaboratorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectRepository) GetAll() ([]*models.Project, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectRepository) AddCollaborator(projectID, collaboratorID uuid.UUID) error {
	args := m.Called(projectID, collaboratorID)
	return args.Error(0)
}

func (m *MockProjectRepository) RemoveCollaborator(projectID, collaboratorID uuid.UUID) error {
	args := m.Called(projectID, collaboratorID)
	return args.Error(0)
}