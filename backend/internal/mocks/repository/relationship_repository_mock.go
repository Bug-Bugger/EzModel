package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRelationshipRepository struct {
	mock.Mock
}

func (m *MockRelationshipRepository) Create(relationship *models.Relationship) (uuid.UUID, error) {
	args := m.Called(relationship)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRelationshipRepository) GetByID(id uuid.UUID) (*models.Relationship, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Relationship), args.Error(1)
}

func (m *MockRelationshipRepository) GetByProjectID(projectID uuid.UUID) ([]*models.Relationship, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Relationship), args.Error(1)
}

func (m *MockRelationshipRepository) GetByTableID(tableID uuid.UUID) ([]*models.Relationship, error) {
	args := m.Called(tableID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Relationship), args.Error(1)
}

func (m *MockRelationshipRepository) Update(relationship *models.Relationship) error {
	args := m.Called(relationship)
	return args.Error(0)
}

func (m *MockRelationshipRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
