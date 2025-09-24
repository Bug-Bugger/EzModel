package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockFieldRepository struct {
	mock.Mock
}

func (m *MockFieldRepository) Create(field *models.Field) (uuid.UUID, error) {
	args := m.Called(field)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockFieldRepository) GetByID(id uuid.UUID) (*models.Field, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *MockFieldRepository) GetByTableID(tableID uuid.UUID) ([]*models.Field, error) {
	args := m.Called(tableID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Field), args.Error(1)
}

func (m *MockFieldRepository) Update(field *models.Field) error {
	args := m.Called(field)
	return args.Error(0)
}

func (m *MockFieldRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFieldRepository) ReorderFields(tableID uuid.UUID, fieldPositions map[uuid.UUID]int) error {
	args := m.Called(tableID, fieldPositions)
	return args.Error(0)
}
