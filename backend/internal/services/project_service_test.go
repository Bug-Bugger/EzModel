package services

import (
	"testing"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	mockRepo "github.com/Bug-Bugger/ezmodel/internal/mocks/repository"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)


// Helper functions
func createTestProjectUser() *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword123",
	}
}

func createTestProject(ownerID uuid.UUID) *models.Project {
	return &models.Project{
		ID:           uuid.New(),
		Name:         "Test Project",
		Description:  "A test project",
		OwnerID:      ownerID,
		DatabaseType: "postgresql",
		CanvasData:   "{}",
	}
}

func projectStringPtr(s string) *string {
	return &s
}

type ProjectServiceTestSuite struct {
	suite.Suite
	mockProjectRepo *mockRepo.MockProjectRepository
	mockUserRepo    *mockRepo.MockUserRepository
	service         *ProjectService
}

func (suite *ProjectServiceTestSuite) SetupTest() {
	suite.mockProjectRepo = new(mockRepo.MockProjectRepository)
	suite.mockUserRepo = new(mockRepo.MockUserRepository)
	suite.service = NewProjectService(suite.mockProjectRepo, suite.mockUserRepo)
}

func TestProjectServiceSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}

// Test CreateProject - Success
func (suite *ProjectServiceTestSuite) TestCreateProject_Success() {
	name := "Test Project"
	description := "A test project"
	ownerID := uuid.New()
	projectID := uuid.New()

	owner := createTestProjectUser()
	owner.ID = ownerID

	// Mock that owner exists
	suite.mockUserRepo.On("GetByID", ownerID).Return(owner, nil)

	// Mock successful creation
	suite.mockProjectRepo.On("Create", mock.MatchedBy(func(project *models.Project) bool {
		return project.Name == name && project.Description == description && project.OwnerID == ownerID
	})).Return(projectID, nil)

	// Execute
	result, err := suite.service.CreateProject(name, description, ownerID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(projectID, result.ID)
	suite.Equal(name, result.Name)
	suite.Equal(description, result.Description)
	suite.Equal(ownerID, result.OwnerID)

	suite.mockUserRepo.AssertExpectations(suite.T())
	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test CreateProject - Invalid Input (empty name)
func (suite *ProjectServiceTestSuite) TestCreateProject_InvalidName() {
	result, err := suite.service.CreateProject("", "Valid description", uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateProject - Invalid Input (name too long)
func (suite *ProjectServiceTestSuite) TestCreateProject_NameTooLong() {
	longName := string(make([]byte, 256)) // 256 characters, exceeds limit
	result, err := suite.service.CreateProject(longName, "Valid description", uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateProject - Invalid Input (description too long)
func (suite *ProjectServiceTestSuite) TestCreateProject_DescriptionTooLong() {
	longDescription := string(make([]byte, 1001)) // 1001 characters, exceeds limit
	result, err := suite.service.CreateProject("Valid name", longDescription, uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateProject - Owner Not Found
func (suite *ProjectServiceTestSuite) TestCreateProject_OwnerNotFound() {
	name := "Test Project"
	description := "A test project"
	ownerID := uuid.New()

	suite.mockUserRepo.On("GetByID", ownerID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateProject(name, description, ownerID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrUserNotFound, err)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

// Test CreateProject - Repository Error on User Check
func (suite *ProjectServiceTestSuite) TestCreateProject_UserRepoError() {
	name := "Test Project"
	description := "A test project"
	ownerID := uuid.New()

	suite.mockUserRepo.On("GetByID", ownerID).Return(nil, assert.AnError)

	result, err := suite.service.CreateProject(name, description, ownerID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

// Test CreateProject - Repository Error on Create
func (suite *ProjectServiceTestSuite) TestCreateProject_ProjectRepoError() {
	name := "Test Project"
	description := "A test project"
	ownerID := uuid.New()

	owner := createTestProjectUser()
	owner.ID = ownerID

	suite.mockUserRepo.On("GetByID", ownerID).Return(owner, nil)
	suite.mockProjectRepo.On("Create", mock.AnythingOfType("*models.Project")).Return(uuid.Nil, assert.AnError)

	result, err := suite.service.CreateProject(name, description, ownerID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockUserRepo.AssertExpectations(suite.T())
	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test GetProjectByID - Success
func (suite *ProjectServiceTestSuite) TestGetProjectByID_Success() {
	projectID := uuid.New()
	expectedProject := createTestProject(uuid.New())
	expectedProject.ID = projectID

	suite.mockProjectRepo.On("GetByID", projectID).Return(expectedProject, nil)

	result, err := suite.service.GetProjectByID(projectID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedProject.ID, result.ID)
	suite.Equal(expectedProject.Name, result.Name)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test GetProjectByID - Not Found
func (suite *ProjectServiceTestSuite) TestGetProjectByID_NotFound() {
	projectID := uuid.New()

	suite.mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetProjectByID(projectID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrProjectNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test GetProjectsByOwnerID - Success
func (suite *ProjectServiceTestSuite) TestGetProjectsByOwnerID_Success() {
	ownerID := uuid.New()
	expectedProjects := []*models.Project{
		createTestProject(ownerID),
		createTestProject(ownerID),
	}

	suite.mockProjectRepo.On("GetByOwnerID", ownerID).Return(expectedProjects, nil)

	result, err := suite.service.GetProjectsByOwnerID(ownerID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test UpdateProject - Success
func (suite *ProjectServiceTestSuite) TestUpdateProject_Success() {
	projectID := uuid.New()
	existingProject := createTestProject(uuid.New())
	existingProject.ID = projectID

	newName := "Updated Project"
	newDescription := "Updated description"
	updateRequest := &dto.UpdateProjectRequest{
		Name:        &newName,
		Description: &newDescription,
	}

	suite.mockProjectRepo.On("GetByID", projectID).Return(existingProject, nil)
	suite.mockProjectRepo.On("Update", mock.MatchedBy(func(project *models.Project) bool {
		return project.ID == projectID && project.Name == newName && project.Description == newDescription
	})).Return(nil)

	result, err := suite.service.UpdateProject(projectID, updateRequest)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(projectID, result.ID)
	suite.Equal(newName, result.Name)
	suite.Equal(newDescription, result.Description)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test UpdateProject - Project Not Found
func (suite *ProjectServiceTestSuite) TestUpdateProject_ProjectNotFound() {
	projectID := uuid.New()
	updateRequest := &dto.UpdateProjectRequest{
		Name: projectStringPtr("New name"),
	}

	suite.mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateProject(projectID, updateRequest)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrProjectNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test UpdateProject - Invalid Name
func (suite *ProjectServiceTestSuite) TestUpdateProject_InvalidName() {
	projectID := uuid.New()
	existingProject := createTestProject(uuid.New())
	existingProject.ID = projectID

	invalidName := "" // Empty name
	updateRequest := &dto.UpdateProjectRequest{
		Name: &invalidName,
	}

	suite.mockProjectRepo.On("GetByID", projectID).Return(existingProject, nil)

	result, err := suite.service.UpdateProject(projectID, updateRequest)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test DeleteProject - Success
func (suite *ProjectServiceTestSuite) TestDeleteProject_Success() {
	projectID := uuid.New()
	existingProject := createTestProject(uuid.New())
	existingProject.ID = projectID

	suite.mockProjectRepo.On("GetByID", projectID).Return(existingProject, nil)
	suite.mockProjectRepo.On("Delete", projectID).Return(nil)

	err := suite.service.DeleteProject(projectID)

	suite.NoError(err)
	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test DeleteProject - Project Not Found
func (suite *ProjectServiceTestSuite) TestDeleteProject_ProjectNotFound() {
	projectID := uuid.New()

	suite.mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.DeleteProject(projectID)

	suite.Error(err)
	suite.Equal(ErrProjectNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test AddCollaborator - Success
func (suite *ProjectServiceTestSuite) TestAddCollaborator_Success() {
	projectID := uuid.New()
	collaboratorID := uuid.New()

	existingProject := createTestProject(uuid.New())
	existingProject.ID = projectID

	collaborator := createTestProjectUser()
	collaborator.ID = collaboratorID

	suite.mockProjectRepo.On("GetByID", projectID).Return(existingProject, nil)
	suite.mockUserRepo.On("GetByID", collaboratorID).Return(collaborator, nil)
	suite.mockProjectRepo.On("AddCollaborator", projectID, collaboratorID).Return(nil)

	err := suite.service.AddCollaborator(projectID, collaboratorID)

	suite.NoError(err)
	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockUserRepo.AssertExpectations(suite.T())
}

// Test RemoveCollaborator - Success
func (suite *ProjectServiceTestSuite) TestRemoveCollaborator_Success() {
	projectID := uuid.New()
	collaboratorID := uuid.New()

	existingProject := createTestProject(uuid.New())
	existingProject.ID = projectID

	collaborator := createTestProjectUser()
	collaborator.ID = collaboratorID

	suite.mockProjectRepo.On("GetByID", projectID).Return(existingProject, nil)
	suite.mockUserRepo.On("GetByID", collaboratorID).Return(collaborator, nil)
	suite.mockProjectRepo.On("RemoveCollaborator", projectID, collaboratorID).Return(nil)

	err := suite.service.RemoveCollaborator(projectID, collaboratorID)

	suite.NoError(err)
	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockUserRepo.AssertExpectations(suite.T())
}