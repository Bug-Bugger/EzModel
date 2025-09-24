package testutil

import (
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(email, username, password string) (*models.User, error) {
	args := m.Called(email, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(id uuid.UUID, req *dto.UpdateUserRequest) (*models.User, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdatePassword(id uuid.UUID, password string) error {
	args := m.Called(id, password)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) AuthenticateUser(email, password string) (*models.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockProjectService is a mock implementation of ProjectServiceInterface
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

// MockJWTService is a mock implementation of JWTServiceInterface
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(user *models.User) (*services.TokenPair, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenPair), args.Error(1)
}

func (m *MockJWTService) RefreshTokens(refreshToken string) (*services.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenPair), args.Error(1)
}

func (m *MockJWTService) GetAccessTokenExpiration() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// MockTableService is a mock implementation of TableServiceInterface
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

func (m *MockTableService) DeleteTable(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}