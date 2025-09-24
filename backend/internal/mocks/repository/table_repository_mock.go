package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTableRepository struct {
	mock.Mock
}

func (m *MockTableRepository) Create(table *models.Table) (uuid.UUID, error) {
	args := m.Called(table)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockTableRepository) GetByID(id uuid.UUID) (*models.Table, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Table), args.Error(1)
}

func (m *MockTableRepository) GetByProjectID(projectID uuid.UUID) ([]*models.Table, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Table), args.Error(1)
}

func (m *MockTableRepository) Update(table *models.Table) error {
	args := m.Called(table)
	return args.Error(0)
}

func (m *MockTableRepository) UpdatePosition(id uuid.UUID, posX, posY float64) error {
	args := m.Called(id, posX, posY)
	return args.Error(0)
}

func (m *MockTableRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
