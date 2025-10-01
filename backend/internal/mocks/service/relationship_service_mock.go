package service

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRelationshipService struct {
	mock.Mock
}

func (m *MockRelationshipService) CreateRelationship(projectID uuid.UUID, req *dto.CreateRelationshipRequest, userID uuid.UUID) (*models.Relationship, error) {
	args := m.Called(projectID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Relationship), args.Error(1)
}

func (m *MockRelationshipService) GetRelationshipByID(id uuid.UUID) (*models.Relationship, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Relationship), args.Error(1)
}

func (m *MockRelationshipService) GetRelationshipsByProjectID(projectID uuid.UUID) ([]*models.Relationship, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Relationship), args.Error(1)
}

func (m *MockRelationshipService) GetRelationshipsByTableID(tableID uuid.UUID) ([]*models.Relationship, error) {
	args := m.Called(tableID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Relationship), args.Error(1)
}

func (m *MockRelationshipService) UpdateRelationship(id uuid.UUID, req *dto.UpdateRelationshipRequest, userID uuid.UUID) (*models.Relationship, error) {
	args := m.Called(id, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Relationship), args.Error(1)
}

func (m *MockRelationshipService) DeleteRelationship(id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
