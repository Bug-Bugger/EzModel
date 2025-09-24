package service

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockFieldService struct {
	mock.Mock
}

func (m *MockFieldService) CreateField(tableID uuid.UUID, req *dto.CreateFieldRequest) (*models.Field, error) {
	args := m.Called(tableID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *MockFieldService) GetFieldByID(id uuid.UUID) (*models.Field, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *MockFieldService) GetFieldsByTableID(tableID uuid.UUID) ([]*models.Field, error) {
	args := m.Called(tableID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Field), args.Error(1)
}

func (m *MockFieldService) UpdateField(id uuid.UUID, req *dto.UpdateFieldRequest) (*models.Field, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *MockFieldService) DeleteField(id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockFieldService) ReorderFields(tableID uuid.UUID, fieldPositions map[uuid.UUID]int) error {
	args := m.Called(tableID, fieldPositions)
	return args.Error(0)
}
