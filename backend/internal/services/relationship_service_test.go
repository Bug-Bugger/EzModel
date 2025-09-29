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

// Mock authorization service for tests
type mockRelationshipAuthService struct {
	mock.Mock
}

func (m *mockRelationshipAuthService) CanUserAccessProject(userID, projectID uuid.UUID) (bool, error) {
	args := m.Called(userID, projectID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRelationshipAuthService) CanUserModifyProject(userID, projectID uuid.UUID) (bool, error) {
	args := m.Called(userID, projectID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRelationshipAuthService) CanUserDeleteCollaborationSession(userID, sessionID uuid.UUID) (bool, error) {
	args := m.Called(userID, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRelationshipAuthService) GetProjectIDFromTable(tableID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(tableID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRelationshipAuthService) GetProjectIDFromRelationship(relationshipID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(relationshipID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRelationshipAuthService) GetProjectIDFromField(fieldID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(fieldID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// Test helper functions
func createTestRelationship(projectID, sourceTableID, targetTableID, sourceFieldID, targetFieldID uuid.UUID) *models.Relationship {
	return &models.Relationship{
		ID:            uuid.New(),
		ProjectID:     projectID,
		SourceTableID: sourceTableID,
		SourceFieldID: sourceFieldID,
		TargetTableID: targetTableID,
		TargetFieldID: targetFieldID,
		RelationType:  "one_to_many",
	}
}

func relationshipStringPtr(s string) *string {
	return &s
}

func relationshipUUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

type RelationshipServiceTestSuite struct {
	suite.Suite
	mockRelationshipRepo *mockRepo.MockRelationshipRepository
	mockProjectRepo      *mockRepo.MockProjectRepository
	mockTableRepo        *mockRepo.MockTableRepository
	mockFieldRepo        *mockRepo.MockFieldRepository
	mockAuthService      *mockRelationshipAuthService
	mockCollaborationService *mockCollaborationService
	service              *RelationshipService
}

func (suite *RelationshipServiceTestSuite) SetupTest() {
	suite.mockRelationshipRepo = new(mockRepo.MockRelationshipRepository)
	suite.mockProjectRepo = new(mockRepo.MockProjectRepository)
	suite.mockTableRepo = new(mockRepo.MockTableRepository)
	suite.mockFieldRepo = new(mockRepo.MockFieldRepository)
	suite.mockAuthService = new(mockRelationshipAuthService)
	suite.mockCollaborationService = new(mockCollaborationService)
	suite.service = NewRelationshipService(
		suite.mockRelationshipRepo,
		suite.mockProjectRepo,
		suite.mockTableRepo,
		suite.mockFieldRepo,
		suite.mockAuthService,
		suite.mockCollaborationService,
	)
}

func TestRelationshipServiceSuite(t *testing.T) {
	suite.Run(t, new(RelationshipServiceTestSuite))
}

// Test CreateRelationship - Success
func (suite *RelationshipServiceTestSuite) TestCreateRelationship_Success() {
	projectID := uuid.New()
	sourceTableID := uuid.New()
	targetTableID := uuid.New()
	sourceFieldID := uuid.New()
	targetFieldID := uuid.New()
	relationshipID := uuid.New()

	req := &dto.CreateRelationshipRequest{
		SourceTableID: sourceTableID,
		SourceFieldID: sourceFieldID,
		TargetTableID: targetTableID,
		TargetFieldID: targetFieldID,
		RelationType:  "one_to_many",
	}

	project := &models.Project{ID: projectID, Name: "Test Project"}
	sourceTable := &models.Table{ID: sourceTableID, Name: "Source Table"}
	targetTable := &models.Table{ID: targetTableID, Name: "Target Table"}
	sourceField := &models.Field{ID: sourceFieldID, Name: "Source Field"}
	targetField := &models.Field{ID: targetFieldID, Name: "Target Field"}

	suite.mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	suite.mockTableRepo.On("GetByID", sourceTableID).Return(sourceTable, nil)
	suite.mockTableRepo.On("GetByID", targetTableID).Return(targetTable, nil)
	suite.mockFieldRepo.On("GetByID", sourceFieldID).Return(sourceField, nil)
	suite.mockFieldRepo.On("GetByID", targetFieldID).Return(targetField, nil)
	suite.mockRelationshipRepo.On("Create", mock.MatchedBy(func(rel *models.Relationship) bool {
		return rel.ProjectID == projectID &&
			rel.SourceTableID == sourceTableID &&
			rel.TargetTableID == targetTableID &&
			rel.SourceFieldID == sourceFieldID &&
			rel.TargetFieldID == targetFieldID &&
			rel.RelationType == "one_to_many"
	})).Return(relationshipID, nil)
	suite.mockCollaborationService.On("NotifyRelationshipCreated", projectID, mock.AnythingOfType("*models.Relationship"), mock.AnythingOfType("uuid.UUID")).Return(nil)

	result, err := suite.service.CreateRelationship(projectID, req, uuid.New())

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(relationshipID, result.ID)
	suite.Equal(projectID, result.ProjectID)
	suite.Equal(sourceTableID, result.SourceTableID)
	suite.Equal(targetTableID, result.TargetTableID)
	suite.Equal(sourceFieldID, result.SourceFieldID)
	suite.Equal(targetFieldID, result.TargetFieldID)
	suite.Equal("one_to_many", result.RelationType)

	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
	suite.mockRelationshipRepo.AssertExpectations(suite.T())
	suite.mockCollaborationService.AssertExpectations(suite.T())
}

// Test CreateRelationship - Default RelationType
func (suite *RelationshipServiceTestSuite) TestCreateRelationship_DefaultRelationType() {
	projectID := uuid.New()
	sourceTableID := uuid.New()
	targetTableID := uuid.New()
	sourceFieldID := uuid.New()
	targetFieldID := uuid.New()
	relationshipID := uuid.New()

	req := &dto.CreateRelationshipRequest{
		SourceTableID: sourceTableID,
		SourceFieldID: sourceFieldID,
		TargetTableID: targetTableID,
		TargetFieldID: targetFieldID,
		RelationType:  "", // Empty, should default to one_to_many
	}

	project := &models.Project{ID: projectID, Name: "Test Project"}
	sourceTable := &models.Table{ID: sourceTableID, Name: "Source Table"}
	targetTable := &models.Table{ID: targetTableID, Name: "Target Table"}
	sourceField := &models.Field{ID: sourceFieldID, Name: "Source Field"}
	targetField := &models.Field{ID: targetFieldID, Name: "Target Field"}

	suite.mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	suite.mockTableRepo.On("GetByID", sourceTableID).Return(sourceTable, nil)
	suite.mockTableRepo.On("GetByID", targetTableID).Return(targetTable, nil)
	suite.mockFieldRepo.On("GetByID", sourceFieldID).Return(sourceField, nil)
	suite.mockFieldRepo.On("GetByID", targetFieldID).Return(targetField, nil)
	suite.mockRelationshipRepo.On("Create", mock.MatchedBy(func(rel *models.Relationship) bool {
		return rel.RelationType == "one_to_many"
	})).Return(relationshipID, nil)
	suite.mockCollaborationService.On("NotifyRelationshipCreated", projectID, mock.AnythingOfType("*models.Relationship"), mock.AnythingOfType("uuid.UUID")).Return(nil)

	result, err := suite.service.CreateRelationship(projectID, req, uuid.New())

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("one_to_many", result.RelationType)

	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
	suite.mockRelationshipRepo.AssertExpectations(suite.T())
	suite.mockCollaborationService.AssertExpectations(suite.T())
}

// Test CreateRelationship - Project Not Found
func (suite *RelationshipServiceTestSuite) TestCreateRelationship_ProjectNotFound() {
	projectID := uuid.New()
	req := &dto.CreateRelationshipRequest{
		SourceTableID: uuid.New(),
		SourceFieldID: uuid.New(),
		TargetTableID: uuid.New(),
		TargetFieldID: uuid.New(),
		RelationType:  "one_to_many",
	}

	suite.mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateRelationship(projectID, req, uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrProjectNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test CreateRelationship - Source Table Not Found
func (suite *RelationshipServiceTestSuite) TestCreateRelationship_SourceTableNotFound() {
	projectID := uuid.New()
	sourceTableID := uuid.New()
	req := &dto.CreateRelationshipRequest{
		SourceTableID: sourceTableID,
		SourceFieldID: uuid.New(),
		TargetTableID: uuid.New(),
		TargetFieldID: uuid.New(),
		RelationType:  "one_to_many",
	}

	project := &models.Project{ID: projectID, Name: "Test Project"}

	suite.mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	suite.mockTableRepo.On("GetByID", sourceTableID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateRelationship(projectID, req, uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrTableNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test CreateRelationship - Source Field Not Found
func (suite *RelationshipServiceTestSuite) TestCreateRelationship_SourceFieldNotFound() {
	projectID := uuid.New()
	sourceTableID := uuid.New()
	targetTableID := uuid.New()
	sourceFieldID := uuid.New()
	req := &dto.CreateRelationshipRequest{
		SourceTableID: sourceTableID,
		SourceFieldID: sourceFieldID,
		TargetTableID: targetTableID,
		TargetFieldID: uuid.New(),
		RelationType:  "one_to_many",
	}

	project := &models.Project{ID: projectID, Name: "Test Project"}
	sourceTable := &models.Table{ID: sourceTableID, Name: "Source Table"}
	targetTable := &models.Table{ID: targetTableID, Name: "Target Table"}

	suite.mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	suite.mockTableRepo.On("GetByID", sourceTableID).Return(sourceTable, nil)
	suite.mockTableRepo.On("GetByID", targetTableID).Return(targetTable, nil)
	suite.mockFieldRepo.On("GetByID", sourceFieldID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateRelationship(projectID, req, uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrFieldNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test GetRelationshipByID - Success
func (suite *RelationshipServiceTestSuite) TestGetRelationshipByID_Success() {
	relationshipID := uuid.New()
	expectedRelationship := createTestRelationship(uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	expectedRelationship.ID = relationshipID

	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(expectedRelationship, nil)

	result, err := suite.service.GetRelationshipByID(relationshipID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedRelationship.ID, result.ID)
	suite.Equal(expectedRelationship.RelationType, result.RelationType)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
}

// Test GetRelationshipByID - Not Found
func (suite *RelationshipServiceTestSuite) TestGetRelationshipByID_NotFound() {
	relationshipID := uuid.New()

	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetRelationshipByID(relationshipID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrRelationshipNotFound, err)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
}

// Test GetRelationshipsByProjectID - Success
func (suite *RelationshipServiceTestSuite) TestGetRelationshipsByProjectID_Success() {
	projectID := uuid.New()
	relationships := []*models.Relationship{
		createTestRelationship(projectID, uuid.New(), uuid.New(), uuid.New(), uuid.New()),
		createTestRelationship(projectID, uuid.New(), uuid.New(), uuid.New(), uuid.New()),
	}

	suite.mockRelationshipRepo.On("GetByProjectID", projectID).Return(relationships, nil)

	result, err := suite.service.GetRelationshipsByProjectID(projectID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
}

// Test GetRelationshipsByTableID - Success
func (suite *RelationshipServiceTestSuite) TestGetRelationshipsByTableID_Success() {
	tableID := uuid.New()
	relationships := []*models.Relationship{
		createTestRelationship(uuid.New(), tableID, uuid.New(), uuid.New(), uuid.New()),
		createTestRelationship(uuid.New(), uuid.New(), tableID, uuid.New(), uuid.New()),
	}

	suite.mockRelationshipRepo.On("GetByTableID", tableID).Return(relationships, nil)

	result, err := suite.service.GetRelationshipsByTableID(tableID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
}

// Test UpdateRelationship - Success
func (suite *RelationshipServiceTestSuite) TestUpdateRelationship_Success() {
	relationshipID := uuid.New()
	existingRelationship := createTestRelationship(uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	existingRelationship.ID = relationshipID

	newSourceTableID := uuid.New()
	newRelationType := "many_to_many"
	updateRequest := &dto.UpdateRelationshipRequest{
		SourceTableID: &newSourceTableID,
		RelationType:  &newRelationType,
	}

	newTable := &models.Table{ID: newSourceTableID, Name: "New Source Table"}

	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(existingRelationship, nil)
	suite.mockTableRepo.On("GetByID", newSourceTableID).Return(newTable, nil)
	suite.mockRelationshipRepo.On("Update", mock.MatchedBy(func(rel *models.Relationship) bool {
		return rel.ID == relationshipID &&
			rel.SourceTableID == newSourceTableID &&
			rel.RelationType == newRelationType
	})).Return(nil)
	suite.mockCollaborationService.On("NotifyRelationshipUpdated", existingRelationship.ProjectID, mock.AnythingOfType("*models.Relationship"), mock.AnythingOfType("uuid.UUID")).Return(nil)

	result, err := suite.service.UpdateRelationship(relationshipID, updateRequest, uuid.New())

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(relationshipID, result.ID)
	suite.Equal(newSourceTableID, result.SourceTableID)
	suite.Equal(newRelationType, result.RelationType)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockCollaborationService.AssertExpectations(suite.T())
}

// Test UpdateRelationship - Not Found
func (suite *RelationshipServiceTestSuite) TestUpdateRelationship_NotFound() {
	relationshipID := uuid.New()
	updateRequest := &dto.UpdateRelationshipRequest{
		RelationType: relationshipStringPtr("many_to_many"),
	}

	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateRelationship(relationshipID, updateRequest, uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrRelationshipNotFound, err)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
}

// Test UpdateRelationship - Source Table Not Found
func (suite *RelationshipServiceTestSuite) TestUpdateRelationship_SourceTableNotFound() {
	relationshipID := uuid.New()
	existingRelationship := createTestRelationship(uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	existingRelationship.ID = relationshipID

	newSourceTableID := uuid.New()
	updateRequest := &dto.UpdateRelationshipRequest{
		SourceTableID: &newSourceTableID,
	}

	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(existingRelationship, nil)
	suite.mockTableRepo.On("GetByID", newSourceTableID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateRelationship(relationshipID, updateRequest, uuid.New())

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrTableNotFound, err)

	suite.mockRelationshipRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test DeleteRelationship - Success
func (suite *RelationshipServiceTestSuite) TestDeleteRelationship_Success() {
	relationshipID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	existingRelationship := createTestRelationship(projectID, uuid.New(), uuid.New(), uuid.New(), uuid.New())
	existingRelationship.ID = relationshipID

	suite.mockAuthService.On("GetProjectIDFromRelationship", relationshipID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(true, nil)
	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(existingRelationship, nil)
	suite.mockRelationshipRepo.On("Delete", relationshipID).Return(nil)
	suite.mockCollaborationService.On("NotifyRelationshipDeleted", projectID, relationshipID, userID).Return(nil)

	err := suite.service.DeleteRelationship(relationshipID, userID)

	suite.NoError(err)
	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockRelationshipRepo.AssertExpectations(suite.T())
	suite.mockCollaborationService.AssertExpectations(suite.T())
}

// Test DeleteRelationship - Not Found
func (suite *RelationshipServiceTestSuite) TestDeleteRelationship_NotFound() {
	relationshipID := uuid.New()
	userID := uuid.New()

	suite.mockAuthService.On("GetProjectIDFromRelationship", relationshipID).Return(uuid.Nil, ErrRelationshipNotFound)

	err := suite.service.DeleteRelationship(relationshipID, userID)

	suite.Error(err)
	suite.Equal(ErrRelationshipNotFound, err)

	suite.mockAuthService.AssertExpectations(suite.T())
}

// Test DeleteRelationship - Forbidden
func (suite *RelationshipServiceTestSuite) TestDeleteRelationship_Forbidden() {
	relationshipID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()

	suite.mockAuthService.On("GetProjectIDFromRelationship", relationshipID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(false, nil)

	err := suite.service.DeleteRelationship(relationshipID, userID)

	suite.Error(err)
	suite.Equal(ErrForbidden, err)

	suite.mockAuthService.AssertExpectations(suite.T())
}

// Test DeleteRelationship - Repository Error
func (suite *RelationshipServiceTestSuite) TestDeleteRelationship_RepositoryError() {
	relationshipID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	existingRelationship := createTestRelationship(projectID, uuid.New(), uuid.New(), uuid.New(), uuid.New())
	existingRelationship.ID = relationshipID

	suite.mockAuthService.On("GetProjectIDFromRelationship", relationshipID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(true, nil)
	suite.mockRelationshipRepo.On("GetByID", relationshipID).Return(existingRelationship, nil)
	suite.mockCollaborationService.On("NotifyRelationshipDeleted", projectID, relationshipID, userID).Return(nil)
	suite.mockRelationshipRepo.On("Delete", relationshipID).Return(assert.AnError)

	err := suite.service.DeleteRelationship(relationshipID, userID)

	suite.Error(err)
	suite.Equal(assert.AnError, err)

	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockCollaborationService.AssertExpectations(suite.T())
	suite.mockRelationshipRepo.AssertExpectations(suite.T())
}
