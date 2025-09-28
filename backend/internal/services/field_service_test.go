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
type mockAuthorizationService struct {
	mock.Mock
}

func (m *mockAuthorizationService) CanUserAccessProject(userID, projectID uuid.UUID) (bool, error) {
	args := m.Called(userID, projectID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAuthorizationService) CanUserModifyProject(userID, projectID uuid.UUID) (bool, error) {
	args := m.Called(userID, projectID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAuthorizationService) CanUserDeleteCollaborationSession(userID, sessionID uuid.UUID) (bool, error) {
	args := m.Called(userID, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAuthorizationService) GetProjectIDFromTable(tableID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(tableID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockAuthorizationService) GetProjectIDFromRelationship(relationshipID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(relationshipID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockAuthorizationService) GetProjectIDFromField(fieldID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(fieldID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// Mock collaboration service for tests
type mockCollaborationService struct {
	mock.Mock
}

func (m *mockCollaborationService) CreateSession(projectID, userID uuid.UUID, userColor string) (*models.CollaborationSession, error) {
	args := m.Called(projectID, userID, userColor)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaborationSession), args.Error(1)
}

func (m *mockCollaborationService) GetSessionByID(id uuid.UUID) (*models.CollaborationSession, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaborationSession), args.Error(1)
}

func (m *mockCollaborationService) GetSessionsByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CollaborationSession), args.Error(1)
}

func (m *mockCollaborationService) GetActiveSessionsByProjectID(projectID uuid.UUID) ([]*models.CollaborationSession, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CollaborationSession), args.Error(1)
}

func (m *mockCollaborationService) UpdateCursor(sessionID uuid.UUID, cursorX, cursorY *float64) error {
	args := m.Called(sessionID, cursorX, cursorY)
	return args.Error(0)
}

func (m *mockCollaborationService) UpdateSession(id uuid.UUID, req *dto.UpdateSessionRequest) (*models.CollaborationSession, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaborationSession), args.Error(1)
}

func (m *mockCollaborationService) SetSessionInactive(sessionID uuid.UUID) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *mockCollaborationService) DeleteSession(sessionID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(sessionID, userID)
	return args.Error(0)
}

func (m *mockCollaborationService) NotifyFieldCreated(projectID uuid.UUID, field *models.Field, senderUserID uuid.UUID) error {
	args := m.Called(projectID, field, senderUserID)
	return args.Error(0)
}

func (m *mockCollaborationService) NotifyFieldUpdated(projectID uuid.UUID, field *models.Field, senderUserID uuid.UUID) error {
	args := m.Called(projectID, field, senderUserID)
	return args.Error(0)
}

func (m *mockCollaborationService) NotifyFieldDeleted(projectID, fieldID uuid.UUID, senderUserID uuid.UUID) error {
	args := m.Called(projectID, fieldID, senderUserID)
	return args.Error(0)
}

// Test helper functions
func createTestField(tableID uuid.UUID) *models.Field {
	return &models.Field{
		ID:           uuid.New(),
		TableID:      tableID,
		Name:         "Test Field",
		DataType:     "VARCHAR(255)",
		IsPrimaryKey: false,
		IsNullable:   true,
		DefaultValue: "",
		Position:     1,
	}
}

func fieldStringPtr(s string) *string {
	return &s
}

type FieldServiceTestSuite struct {
	suite.Suite
	mockFieldRepo         *mockRepo.MockFieldRepository
	mockTableRepo         *mockRepo.MockTableRepository
	mockAuthService       *mockAuthorizationService
	mockCollabService     *mockCollaborationService
	service               *FieldService
}

func (suite *FieldServiceTestSuite) SetupTest() {
	suite.mockFieldRepo = new(mockRepo.MockFieldRepository)
	suite.mockTableRepo = new(mockRepo.MockTableRepository)
	suite.mockAuthService = new(mockAuthorizationService)
	suite.mockCollabService = new(mockCollaborationService)
	suite.service = NewFieldService(suite.mockFieldRepo, suite.mockTableRepo, suite.mockAuthService, suite.mockCollabService)
}

func TestFieldServiceSuite(t *testing.T) {
	suite.Run(t, new(FieldServiceTestSuite))
}

// Test CreateField - Success
func (suite *FieldServiceTestSuite) TestCreateField_Success() {
	tableID := uuid.New()
	fieldID := uuid.New()
	req := &dto.CreateFieldRequest{
		Name:         "test_field",
		DataType:     "VARCHAR(255)",
		IsPrimaryKey: true,
		IsNullable:   false,
		DefaultValue: "default_value",
		Position:     1,
	}

	table := &models.Table{
		ID:        tableID,
		Name:      "Test Table",
		ProjectID: uuid.New(),
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(table, nil)
	suite.mockFieldRepo.On("Create", mock.MatchedBy(func(field *models.Field) bool {
		return field.TableID == tableID &&
			field.Name == "test_field" &&
			field.DataType == "VARCHAR(255)" &&
			field.IsPrimaryKey == true &&
			field.IsNullable == false &&
			field.DefaultValue == "default_value" &&
			field.Position == 1
	})).Return(fieldID, nil)

	userID := uuid.New()
	suite.mockCollabService.On("NotifyFieldCreated", table.ProjectID, mock.AnythingOfType("*models.Field"), userID).Return(nil)
	result, err := suite.service.CreateField(tableID, req, userID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(fieldID, result.ID)
	suite.Equal(tableID, result.TableID)
	suite.Equal("test_field", result.Name)
	suite.Equal("VARCHAR(255)", result.DataType)
	suite.Equal(true, result.IsPrimaryKey)
	suite.Equal(false, result.IsNullable)
	suite.Equal("default_value", result.DefaultValue)
	suite.Equal(1, result.Position)

	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
	suite.mockCollabService.AssertExpectations(suite.T())
}

// Test CreateField - Invalid Name (empty)
func (suite *FieldServiceTestSuite) TestCreateField_InvalidNameEmpty() {
	tableID := uuid.New()
	req := &dto.CreateFieldRequest{
		Name:     "",
		DataType: "VARCHAR(255)",
	}

	userID := uuid.New()
	result, err := suite.service.CreateField(tableID, req, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateField - Invalid Name (too long)
func (suite *FieldServiceTestSuite) TestCreateField_InvalidNameTooLong() {
	tableID := uuid.New()
	longName := string(make([]byte, 256))
	req := &dto.CreateFieldRequest{
		Name:     longName,
		DataType: "VARCHAR(255)",
	}

	userID := uuid.New()
	result, err := suite.service.CreateField(tableID, req, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateField - Invalid DataType (empty)
func (suite *FieldServiceTestSuite) TestCreateField_InvalidDataType() {
	tableID := uuid.New()
	req := &dto.CreateFieldRequest{
		Name:     "test_field",
		DataType: "",
	}

	userID := uuid.New()
	result, err := suite.service.CreateField(tableID, req, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateField - Table Not Found
func (suite *FieldServiceTestSuite) TestCreateField_TableNotFound() {
	tableID := uuid.New()
	req := &dto.CreateFieldRequest{
		Name:     "test_field",
		DataType: "VARCHAR(255)",
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(nil, gorm.ErrRecordNotFound)

	userID := uuid.New()
	result, err := suite.service.CreateField(tableID, req, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrTableNotFound, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test CreateField - Repository Error on Create
func (suite *FieldServiceTestSuite) TestCreateField_RepositoryError() {
	tableID := uuid.New()
	req := &dto.CreateFieldRequest{
		Name:     "test_field",
		DataType: "VARCHAR(255)",
	}

	table := &models.Table{
		ID:        tableID,
		Name:      "Test Table",
		ProjectID: uuid.New(),
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(table, nil)
	suite.mockFieldRepo.On("Create", mock.AnythingOfType("*models.Field")).Return(uuid.Nil, assert.AnError)

	userID := uuid.New()
	result, err := suite.service.CreateField(tableID, req, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test GetFieldByID - Success
func (suite *FieldServiceTestSuite) TestGetFieldByID_Success() {
	fieldID := uuid.New()
	expectedField := createTestField(uuid.New())
	expectedField.ID = fieldID

	suite.mockFieldRepo.On("GetByID", fieldID).Return(expectedField, nil)

	result, err := suite.service.GetFieldByID(fieldID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedField.ID, result.ID)
	suite.Equal(expectedField.Name, result.Name)

	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test GetFieldByID - Not Found
func (suite *FieldServiceTestSuite) TestGetFieldByID_NotFound() {
	fieldID := uuid.New()

	suite.mockFieldRepo.On("GetByID", fieldID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetFieldByID(fieldID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrFieldNotFound, err)

	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test GetFieldsByTableID - Success
func (suite *FieldServiceTestSuite) TestGetFieldsByTableID_Success() {
	tableID := uuid.New()
	fields := []*models.Field{
		createTestField(tableID),
		createTestField(tableID),
	}

	suite.mockFieldRepo.On("GetByTableID", tableID).Return(fields, nil)

	result, err := suite.service.GetFieldsByTableID(tableID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)

	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test UpdateField - Success
func (suite *FieldServiceTestSuite) TestUpdateField_Success() {
	fieldID := uuid.New()
	existingField := createTestField(uuid.New())
	existingField.ID = fieldID

	newName := "updated_field"
	newDataType := "INT"
	isPrimaryKey := true
	updateRequest := &dto.UpdateFieldRequest{
		Name:         &newName,
		DataType:     &newDataType,
		IsPrimaryKey: &isPrimaryKey,
	}

	table := &models.Table{
		ID:        existingField.TableID,
		Name:      "Test Table",
		ProjectID: uuid.New(),
	}

	suite.mockFieldRepo.On("GetByID", fieldID).Return(existingField, nil)
	suite.mockFieldRepo.On("Update", mock.MatchedBy(func(field *models.Field) bool {
		return field.ID == fieldID &&
			field.Name == newName &&
			field.DataType == newDataType &&
			field.IsPrimaryKey == isPrimaryKey
	})).Return(nil)
	suite.mockTableRepo.On("GetByID", existingField.TableID).Return(table, nil)

	userID := uuid.New()
	suite.mockCollabService.On("NotifyFieldUpdated", table.ProjectID, mock.AnythingOfType("*models.Field"), userID).Return(nil)
	result, err := suite.service.UpdateField(fieldID, updateRequest, userID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(fieldID, result.ID)
	suite.Equal(newName, result.Name)
	suite.Equal(newDataType, result.DataType)
	suite.Equal(isPrimaryKey, result.IsPrimaryKey)

	suite.mockFieldRepo.AssertExpectations(suite.T())
	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockCollabService.AssertExpectations(suite.T())
}

// Test UpdateField - Not Found
func (suite *FieldServiceTestSuite) TestUpdateField_NotFound() {
	fieldID := uuid.New()
	updateRequest := &dto.UpdateFieldRequest{
		Name: fieldStringPtr("new_name"),
	}

	suite.mockFieldRepo.On("GetByID", fieldID).Return(nil, gorm.ErrRecordNotFound)

	userID := uuid.New()
	result, err := suite.service.UpdateField(fieldID, updateRequest, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrFieldNotFound, err)

	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test UpdateField - Invalid Name
func (suite *FieldServiceTestSuite) TestUpdateField_InvalidName() {
	fieldID := uuid.New()
	existingField := createTestField(uuid.New())
	existingField.ID = fieldID

	invalidName := ""
	updateRequest := &dto.UpdateFieldRequest{
		Name: &invalidName,
	}

	suite.mockFieldRepo.On("GetByID", fieldID).Return(existingField, nil)

	userID := uuid.New()
	result, err := suite.service.UpdateField(fieldID, updateRequest, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)

	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test UpdateField - Invalid DataType
func (suite *FieldServiceTestSuite) TestUpdateField_InvalidDataType() {
	fieldID := uuid.New()
	existingField := createTestField(uuid.New())
	existingField.ID = fieldID

	invalidDataType := ""
	updateRequest := &dto.UpdateFieldRequest{
		DataType: &invalidDataType,
	}

	suite.mockFieldRepo.On("GetByID", fieldID).Return(existingField, nil)

	userID := uuid.New()
	result, err := suite.service.UpdateField(fieldID, updateRequest, userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)

	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test DeleteField - Success
func (suite *FieldServiceTestSuite) TestDeleteField_Success() {
	fieldID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	existingField := createTestField(uuid.New())
	existingField.ID = fieldID

	suite.mockAuthService.On("GetProjectIDFromField", fieldID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(true, nil)
	suite.mockFieldRepo.On("GetByID", fieldID).Return(existingField, nil)
	suite.mockFieldRepo.On("Delete", fieldID).Return(nil)
	suite.mockCollabService.On("NotifyFieldDeleted", projectID, fieldID, userID).Return(nil)

	err := suite.service.DeleteField(fieldID, userID)

	suite.NoError(err)
	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
	suite.mockCollabService.AssertExpectations(suite.T())
}

// Test DeleteField - Not Found
func (suite *FieldServiceTestSuite) TestDeleteField_NotFound() {
	fieldID := uuid.New()
	userID := uuid.New()

	suite.mockAuthService.On("GetProjectIDFromField", fieldID).Return(uuid.Nil, ErrFieldNotFound)

	err := suite.service.DeleteField(fieldID, userID)

	suite.Error(err)
	suite.Equal(ErrFieldNotFound, err)

	suite.mockAuthService.AssertExpectations(suite.T())
}

// Test DeleteField - Forbidden
func (suite *FieldServiceTestSuite) TestDeleteField_Forbidden() {
	fieldID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()

	suite.mockAuthService.On("GetProjectIDFromField", fieldID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(false, nil)

	err := suite.service.DeleteField(fieldID, userID)

	suite.Error(err)
	suite.Equal(ErrForbidden, err)

	suite.mockAuthService.AssertExpectations(suite.T())
}

// Test DeleteField - Repository Error
func (suite *FieldServiceTestSuite) TestDeleteField_RepositoryError() {
	fieldID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	existingField := createTestField(uuid.New())
	existingField.ID = fieldID

	suite.mockAuthService.On("GetProjectIDFromField", fieldID).Return(projectID, nil)
	suite.mockAuthService.On("CanUserModifyProject", userID, projectID).Return(true, nil)
	suite.mockFieldRepo.On("GetByID", fieldID).Return(existingField, nil)
	suite.mockFieldRepo.On("Delete", fieldID).Return(assert.AnError)

	err := suite.service.DeleteField(fieldID, userID)

	suite.Error(err)
	suite.Equal(assert.AnError, err)

	suite.mockAuthService.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test ReorderFields - Success
func (suite *FieldServiceTestSuite) TestReorderFields_Success() {
	tableID := uuid.New()
	fieldID1 := uuid.New()
	fieldID2 := uuid.New()

	table := &models.Table{
		ID:        tableID,
		Name:      "Test Table",
		ProjectID: uuid.New(),
	}

	field1 := createTestField(tableID)
	field1.ID = fieldID1

	field2 := createTestField(tableID)
	field2.ID = fieldID2

	fieldPositions := map[uuid.UUID]int{
		fieldID1: 1,
		fieldID2: 2,
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(table, nil)
	suite.mockFieldRepo.On("GetByID", fieldID1).Return(field1, nil)
	suite.mockFieldRepo.On("GetByID", fieldID2).Return(field2, nil)
	suite.mockFieldRepo.On("ReorderFields", tableID, fieldPositions).Return(nil)

	err := suite.service.ReorderFields(tableID, fieldPositions)

	suite.NoError(err)

	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test ReorderFields - Table Not Found
func (suite *FieldServiceTestSuite) TestReorderFields_TableNotFound() {
	tableID := uuid.New()
	fieldPositions := map[uuid.UUID]int{
		uuid.New(): 1,
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.ReorderFields(tableID, fieldPositions)

	suite.Error(err)
	suite.Equal(ErrTableNotFound, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
}

// Test ReorderFields - Field Not Found
func (suite *FieldServiceTestSuite) TestReorderFields_FieldNotFound() {
	tableID := uuid.New()
	fieldID := uuid.New()

	table := &models.Table{
		ID:        tableID,
		Name:      "Test Table",
		ProjectID: uuid.New(),
	}

	fieldPositions := map[uuid.UUID]int{
		fieldID: 1,
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(table, nil)
	suite.mockFieldRepo.On("GetByID", fieldID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.ReorderFields(tableID, fieldPositions)

	suite.Error(err)
	suite.Equal(ErrFieldNotFound, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
}

// Test ReorderFields - Field Belongs To Different Table
func (suite *FieldServiceTestSuite) TestReorderFields_FieldBelongsToDifferentTable() {
	tableID := uuid.New()
	fieldID := uuid.New()
	differentTableID := uuid.New()

	table := &models.Table{
		ID:        tableID,
		Name:      "Test Table",
		ProjectID: uuid.New(),
	}

	field := createTestField(differentTableID)
	field.ID = fieldID

	fieldPositions := map[uuid.UUID]int{
		fieldID: 1,
	}

	suite.mockTableRepo.On("GetByID", tableID).Return(table, nil)
	suite.mockFieldRepo.On("GetByID", fieldID).Return(field, nil)

	err := suite.service.ReorderFields(tableID, fieldPositions)

	suite.Error(err)
	suite.Equal(ErrInvalidInput, err)

	suite.mockTableRepo.AssertExpectations(suite.T())
	suite.mockFieldRepo.AssertExpectations(suite.T())
}
