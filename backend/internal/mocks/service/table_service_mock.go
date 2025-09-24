package service

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTableService struct {
	mock.Mock
}

func (m *MockTableService) CreateTable(projectID uuid.UUID, name string, posX, posY float64) (*models.Table, error) {
	args := m.Called(projectID, name, posX, posY)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Table), args.Error(1)
}

func (m *MockTableService) GetTableByID(id uuid.UUID) (*models.Table, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Table), args.Error(1)
}

func (m *MockTableService) GetTablesByProjectID(projectID uuid.UUID) ([]*models.Table, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Table), args.Error(1)
}

func (m *MockTableService) UpdateTable(id uuid.UUID, req *dto.UpdateTableRequest) (*models.Table, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Table), args.Error(1)
}

func (m *MockTableService) UpdateTablePosition(id uuid.UUID, posX, posY float64) error {
	args := m.Called(id, posX, posY)
	return args.Error(0)
}

func (m *MockTableService) DeleteTable(id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
