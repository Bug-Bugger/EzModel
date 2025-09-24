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
type mockTableAuthService struct {
	mock.Mock
}

func (m *mockTableAuthService) CanUserAccessProject(userID, projectID uuid.UUID) (bool, error) {
	args := m.Called(userID, projectID)
	return args.Bool(0), args.Error(1)
}

func (m *mockTableAuthService) CanUserModifyProject(userID, projectID uuid.UUID) (bool, error) {
	args := m.Called(userID, projectID)
	return args.Bool(0), args.Error(1)
}

func (m *mockTableAuthService) CanUserDeleteCollaborationSession(userID, sessionID uuid.UUID) (bool, error) {
	args := m.Called(userID, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *mockTableAuthService) GetProjectIDFromTable(tableID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(tableID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockTableAuthService) GetProjectIDFromRelationship(relationshipID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(relationshipID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockTableAuthService) GetProjectIDFromField(fieldID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(fieldID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// Test helper functions
func createTestTable(projectID uuid.UUID) *models.Table {
	return &models.Table{
		ID:        uuid.New(),
		Name:      "Test Table",
		ProjectID: projectID,
		PosX:      100.0,
		PosY:      200.0,
	}
}

func tableStringPtr(s string) *string {
	return &s
}

type TableServiceTestSuite struct {
	suite.Suite
	mockTableRepo   *mockRepo.MockTableRepository
	mockProjectRepo *mockRepo.MockProjectRepository
	mockAuthService *mockTableAuthService
	service         *TableService
}

func (suite *TableServiceTestSuite) SetupTest() {
	suite.mockTableRepo = new(mockRepo.MockTableRepository)
	suite.mockProjectRepo = new(mockRepo.MockProjectRepository)
	suite.mockAuthService = new(mockTableAuthService)
	suite.service = NewTableService(suite.mockTableRepo, suite.mockProjectRepo, suite.mockAuthService)
}

func TestTableServiceSuite(t *testing.T) {
	suite.Run(t, new(TableServiceTestSuite))
}

// Test CreateTable - Success
func (suite *TableServiceTestSuite) TestCreateTable_Success() {
	projectID := uuid.New()
	name := "Test Table"
	posX := 100.0
	posY := 200.0
	tableID := uuid.New()

	project := &models.Project{
		ID:     projectID,
		Name:   "Test Project",
		OwnerID: uuid.New(),
	}

	suite.mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	suite.mockTableRepo.On("Create", mock.MatchedBy(func(table *models.Table) bool {
		return table.Name == name && table.ProjectID == projectID && table.PosX == posX && table.PosY == posY
	})).Return(tableID, nil)

	result, err := suite.service.CreateTable(projectID, name, posX, posY)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(tableID, result.ID)
	suite.Equal(name, result.Name)
	suite.Equal(projectID, result.ProjectID)
	suite.Equal(posX, result.PosX)
	suite.Equal(posY, result.PosY)

	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test CreateTable - Invalid Input (empty name)
func (suite *TableServiceTestSuite) TestCreateTable_InvalidName() {
	projectID := uuid.New()

	result, err := suite.service.CreateTable(projectID, "", 100.0, 200.0)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateTable - Name Too Long
func (suite *TableServiceTestSuite) TestCreateTable_NameTooLong() {
	projectID := uuid.New()
	longName := string(make([]byte, 256))

	result, err := suite.service.CreateTable(projectID, longName, 100.0, 200.0)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateTable - Project Not Found
func (suite *TableServiceTestSuite) TestCreateTable_ProjectNotFound() {
	projectID := uuid.New()
	name := "Test Table"

	suite.mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateTable(projectID, name, 100.0, 200.0)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrProjectNotFound, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
}

// Test CreateTable - Repository Error on Create
func (suite *TableServiceTestSuite) TestCreateTable_RepositoryError() {
	projectID := uuid.New()
	name := "Test Table"

	project := &models.Project{
		ID:     projectID,
		Name:   "Test Project",
		OwnerID: uuid.New(),
	}

	suite.mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	suite.mockTableRepo.On("Create", mock.AnythingOfType("*models.Table")).Return(uuid.Nil, assert.AnError)

	result, err := suite.service.CreateTable(projectID, name, 100.0, 200.0)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockProjectRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test GetTableByID - Success
func (suite *TableServiceTestSuite) TestGetTableByID_Success() {
	tableID := uuid.New()
	expectedTable := createTestTable(uuid.New())
	expectedTable.ID = tableID

	suite.mockTableRepo.On("GetByID", tableID).Return(expectedTable, nil)

	result, err := suite.service.GetTableByID(tableID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedTable.ID, result.ID)
	suite.Equal(expectedTable.Name, result.Name)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test GetTableByID - Not Found
func (suite *TableServiceTestSuite) TestGetTableByID_NotFound() {
	tableID := uuid.New()

	suite.mockTableRepo.On("GetByID", tableID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetTableByID(tableID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrTableNotFound, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test GetTablesByProjectID - Success
func (suite *TableServiceTestSuite) TestGetTablesByProjectID_Success() {
	projectID := uuid.New()
	tables := []*models.Table{
		createTestTable(projectID),
		createTestTable(projectID),
	}

	suite.mockTableRepo.On("GetByProjectID", projectID).Return(tables, nil)

	result, err := suite.service.GetTablesByProjectID(projectID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test UpdateTable - Success
func (suite *TableServiceTestSuite) TestUpdateTable_Success() {
	tableID := uuid.New()
	existingTable := createTestTable(uuid.New())
	existingTable.ID = tableID

	newName := "Updated Table"
	updateRequest := &dto.UpdateTableRequest{
		Name: &newName,
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(existingTable, nil)
	suite.mockTableRepo.On("Update", mock.MatchedBy(func(table *models.Table) bool {
		return table.ID == tableID && table.Name == newName
	})).Return(nil)

	result, err := suite.service.UpdateTable(tableID, updateRequest)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(tableID, result.ID)
	suite.Equal(newName, result.Name)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test UpdateTable - Not Found
func (suite *TableServiceTestSuite) TestUpdateTable_NotFound() {
	tableID := uuid.New()
	updateRequest := &dto.UpdateTableRequest{
		Name: tableStringPtr("New name"),
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateTable(tableID, updateRequest)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrTableNotFound, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test UpdateTable - Invalid Name
func (suite *TableServiceTestSuite) TestUpdateTable_InvalidName() {
	tableID := uuid.New()
	existingTable := createTestTable(uuid.New())
	existingTable.ID = tableID

	invalidName := ""
	updateRequest := &dto.UpdateTableRequest{
		Name: &invalidName,
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(existingTable, nil)

	result, err := suite.service.UpdateTable(tableID, updateRequest)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test UpdateTablePosition - Success
func (suite *TableServiceTestSuite) TestUpdateTablePosition_Success() {
	tableID := uuid.New()
	newPosX := 300.0
	newPosY := 400.0

	existingTable := createTestTable(uuid.New())
	existingTable.ID = tableID

	suite.mockTableRepo.On("GetByID", tableID).Return(existingTable, nil)
	suite.mockTableRepo.On("UpdatePosition", tableID, newPosX, newPosY).Return(nil)

	err := suite.service.UpdateTablePosition(tableID, newPosX, newPosY)

	suite.NoError(err)
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test UpdateTablePosition - Table Not Found
func (suite *TableServiceTestSuite) TestUpdateTablePosition_NotFound() {
	tableID := uuid.New()
	newPosX := 300.0
	newPosY := 400.0

	suite.mockTableRepo.On("GetByID", tableID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.UpdateTablePosition(tableID, newPosX, newPosY)

	suite.Error(err)
	suite.Equal(ErrTableNotFound, err)
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test UpdateTablePosition - Repository Error
func (suite *TableServiceTestSuite) TestUpdateTablePosition_RepositoryError() {
	tableID := uuid.New()
	newPosX := 300.0
	newPosY := 400.0

	existingTable := createTestTable(uuid.New())
	existingTable.ID = tableID

	suite.mockTableRepo.On("GetByID", tableID).Return(existingTable, nil)
	suite.mockTableRepo.On("UpdatePosition", tableID, newPosX, newPosY).Return(assert.AnError)

	err := suite.service.UpdateTablePosition(tableID, newPosX, newPosY)

	suite.Error(err)
	suite.Equal(assert.AnError, err)
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test DeleteTable - Success
func (suite *TableServiceTestSuite) TestDeleteTable_Success() {
	tableID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	existingTable := createTestTable(projectID)
	existingTable.ID = tableID

	suite.mockAuthService.On("GetProjectIDFromTable", tableID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(true, nil)
	suite.mockTableRepo.On("GetByID", tableID).Return(existingTable, nil)
	suite.mockTableRepo.On("Delete", tableID).Return(nil)

	err := suite.service.DeleteTable(tableID, userID)

	suite.NoError(err)
	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test DeleteTable - Not Found
func (suite *TableServiceTestSuite) TestDeleteTable_NotFound() {
	tableID := uuid.New()
	userID := uuid.New()

	suite.mockAuthService.On("GetProjectIDFromTable", tableID).Return(uuid.Nil, ErrTableNotFound)

	err := suite.service.DeleteTable(tableID, userID)

	suite.Error(err)
	suite.Equal(ErrTableNotFound, err)

	suite.mockAuthService.AssertExpectations(suite.T())
}

// Test DeleteTable - Forbidden
func (suite *TableServiceTestSuite) TestDeleteTable_Forbidden() {
	tableID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()

	suite.mockAuthService.On("GetProjectIDFromTable", tableID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(false, nil)

	err := suite.service.DeleteTable(tableID, userID)

	suite.Error(err)
	suite.Equal(ErrForbidden, err)

	suite.mockAuthService.AssertExpectations(suite.T())
}

// Test DeleteTable - Repository Error
func (suite *TableServiceTestSuite) TestDeleteTable_RepositoryError() {
	tableID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	existingTable := createTestTable(projectID)
	existingTable.ID = tableID

	suite.mockAuthService.On("GetProjectIDFromTable", tableID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(true, nil)
	suite.mockTableRepo.On("GetByID", tableID).Return(existingTable, nil)
	suite.mockTableRepo.On("Delete", tableID).Return(assert.AnError)

	err := suite.service.DeleteTable(tableID, userID)

	suite.Error(err)
	suite.Equal(assert.AnError, err)

	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
}