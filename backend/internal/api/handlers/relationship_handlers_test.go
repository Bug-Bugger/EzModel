package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	mockService "github.com/Bug-Bugger/ezmodel/internal/mocks/service"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/Bug-Bugger/ezmodel/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

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

func createValidRelationshipRequest() dto.CreateRelationshipRequest {
	return dto.CreateRelationshipRequest{
		SourceTableID: uuid.New(),
		SourceFieldID: uuid.New(),
		TargetTableID: uuid.New(),
		TargetFieldID: uuid.New(),
		RelationType:  "one_to_many",
	}
}

func createValidUpdateRelationshipRequest() dto.UpdateRelationshipRequest {
	relationType := "many_to_many"
	return dto.UpdateRelationshipRequest{
		RelationType: &relationType,
	}
}

type RelationshipHandlerTestSuite struct {
	suite.Suite
	mockRelationshipService *mockService.MockRelationshipService
	handler                 *RelationshipHandler
}

func (suite *RelationshipHandlerTestSuite) SetupTest() {
	suite.mockRelationshipService = new(mockService.MockRelationshipService)
	suite.handler = NewRelationshipHandler(suite.mockRelationshipService)
}

func TestRelationshipHandlerSuite(t *testing.T) {
	suite.Run(t, new(RelationshipHandlerTestSuite))
}

// Test Create - Success
func (suite *RelationshipHandlerTestSuite) TestCreate_Success() {
	projectID := uuid.New()
	relationshipRequest := createValidRelationshipRequest()
	relationship := createTestRelationship(
		projectID,
		relationshipRequest.SourceTableID,
		relationshipRequest.TargetTableID,
		relationshipRequest.SourceFieldID,
		relationshipRequest.TargetFieldID,
	)

	suite.mockRelationshipService.On("CreateRelationship", projectID, &relationshipRequest, mock.AnythingOfType("uuid.UUID")).Return(relationship, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", relationshipRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusCreated, "Relationship created successfully")

	relationshipResponse, ok := response.Data.(map[string]any)
	suite.True(ok, "Response data should be a relationship object")
	suite.Equal(relationship.ID.String(), relationshipResponse["id"])
	suite.Equal(relationship.RelationType, relationshipResponse["relation_type"])

	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Create - Invalid Project ID
func (suite *RelationshipHandlerTestSuite) TestCreate_InvalidProjectID() {
	relationshipRequest := createValidRelationshipRequest()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/invalid-id/relationships", relationshipRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid project ID format")
}

// Test Create - Invalid JSON
func (suite *RelationshipHandlerTestSuite) TestCreate_InvalidJSON() {
	projectID := uuid.New()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", "invalid json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid request body")
}

// Test Create - Validation Error
func (suite *RelationshipHandlerTestSuite) TestCreate_ValidationError() {
	projectID := uuid.New()
	invalidRequest := dto.CreateRelationshipRequest{
		// Missing required fields
		RelationType: "one_to_many",
	}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", invalidRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	// Should return validation errors (400 status)
	suite.Equal(http.StatusBadRequest, w.Code)
}

// Test Create - Project Not Found
func (suite *RelationshipHandlerTestSuite) TestCreate_ProjectNotFound() {
	projectID := uuid.New()
	relationshipRequest := createValidRelationshipRequest()

	suite.mockRelationshipService.On("CreateRelationship", projectID, &relationshipRequest, mock.AnythingOfType("uuid.UUID")).Return(nil, services.ErrProjectNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", relationshipRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Project not found")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Create - Table Not Found
func (suite *RelationshipHandlerTestSuite) TestCreate_TableNotFound() {
	projectID := uuid.New()
	relationshipRequest := createValidRelationshipRequest()

	suite.mockRelationshipService.On("CreateRelationship", projectID, &relationshipRequest, mock.AnythingOfType("uuid.UUID")).Return(nil, services.ErrTableNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", relationshipRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Table not found")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Create - Field Not Found
func (suite *RelationshipHandlerTestSuite) TestCreate_FieldNotFound() {
	projectID := uuid.New()
	relationshipRequest := createValidRelationshipRequest()

	suite.mockRelationshipService.On("CreateRelationship", projectID, &relationshipRequest, mock.AnythingOfType("uuid.UUID")).Return(nil, services.ErrFieldNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", relationshipRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Field not found")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Create - Service Error
func (suite *RelationshipHandlerTestSuite) TestCreate_ServiceError() {
	projectID := uuid.New()
	relationshipRequest := createValidRelationshipRequest()

	suite.mockRelationshipService.On("CreateRelationship", projectID, &relationshipRequest, mock.AnythingOfType("uuid.UUID")).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/relationships", relationshipRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByID - Success
func (suite *RelationshipHandlerTestSuite) TestGetByID_Success() {
	relationshipID := uuid.New()
	relationship := createTestRelationship(uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	relationship.ID = relationshipID

	suite.mockRelationshipService.On("GetRelationshipByID", relationshipID).Return(relationship, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/relationships/"+relationshipID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Relationship retrieved successfully")

	relationshipResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(relationship.ID.String(), relationshipResponse["id"])
	suite.Equal(relationship.RelationType, relationshipResponse["relation_type"])

	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByID - Invalid Relationship ID
func (suite *RelationshipHandlerTestSuite) TestGetByID_InvalidRelationshipID() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/relationships/invalid-id", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid relationship ID format")
}

// Test GetByID - Relationship Not Found
func (suite *RelationshipHandlerTestSuite) TestGetByID_RelationshipNotFound() {
	relationshipID := uuid.New()

	suite.mockRelationshipService.On("GetRelationshipByID", relationshipID).Return(nil, services.ErrRelationshipNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/relationships/"+relationshipID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Relationship not found")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByID - Service Error
func (suite *RelationshipHandlerTestSuite) TestGetByID_ServiceError() {
	relationshipID := uuid.New()

	suite.mockRelationshipService.On("GetRelationshipByID", relationshipID).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/relationships/"+relationshipID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByProjectID - Success
func (suite *RelationshipHandlerTestSuite) TestGetByProjectID_Success() {
	projectID := uuid.New()
	relationships := []*models.Relationship{
		createTestRelationship(projectID, uuid.New(), uuid.New(), uuid.New(), uuid.New()),
		createTestRelationship(projectID, uuid.New(), uuid.New(), uuid.New(), uuid.New()),
	}

	suite.mockRelationshipService.On("GetRelationshipsByProjectID", projectID).Return(relationships, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/projects/"+projectID.String()+"/relationships", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByProjectID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Relationships retrieved successfully")

	relationshipResponses, ok := response.Data.([]any)
	suite.True(ok)
	suite.Len(relationshipResponses, 2)

	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByProjectID - Invalid Project ID
func (suite *RelationshipHandlerTestSuite) TestGetByProjectID_InvalidProjectID() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/projects/invalid-id/relationships", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByProjectID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid project ID format")
}

// Test GetByProjectID - Service Error
func (suite *RelationshipHandlerTestSuite) TestGetByProjectID_ServiceError() {
	projectID := uuid.New()

	suite.mockRelationshipService.On("GetRelationshipsByProjectID", projectID).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/projects/"+projectID.String()+"/relationships", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByProjectID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByTableID - Success
func (suite *RelationshipHandlerTestSuite) TestGetByTableID_Success() {
	tableID := uuid.New()
	relationships := []*models.Relationship{
		createTestRelationship(uuid.New(), tableID, uuid.New(), uuid.New(), uuid.New()),
		createTestRelationship(uuid.New(), uuid.New(), tableID, uuid.New(), uuid.New()),
	}

	suite.mockRelationshipService.On("GetRelationshipsByTableID", tableID).Return(relationships, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/tables/"+tableID.String()+"/relationships", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByTableID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Relationships retrieved successfully")

	relationshipResponses, ok := response.Data.([]any)
	suite.True(ok)
	suite.Len(relationshipResponses, 2)

	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test GetByTableID - Invalid Table ID
func (suite *RelationshipHandlerTestSuite) TestGetByTableID_InvalidTableID() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/tables/invalid-id/relationships", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByTableID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid table ID format")
}

// Test Update - Success
func (suite *RelationshipHandlerTestSuite) TestUpdate_Success() {
	relationshipID := uuid.New()
	updateRequest := createValidUpdateRelationshipRequest()
	updatedRelationship := createTestRelationship(uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	updatedRelationship.ID = relationshipID
	updatedRelationship.RelationType = *updateRequest.RelationType

	suite.mockRelationshipService.On("UpdateRelationship", relationshipID, &updateRequest, mock.AnythingOfType("uuid.UUID")).Return(updatedRelationship, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/relationships/"+relationshipID.String(), updateRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Update()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Relationship updated successfully")

	relationshipResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(updatedRelationship.ID.String(), relationshipResponse["id"])
	suite.Equal(updatedRelationship.RelationType, relationshipResponse["relation_type"])

	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Update - Invalid Relationship ID
func (suite *RelationshipHandlerTestSuite) TestUpdate_InvalidRelationshipID() {
	updateRequest := createValidUpdateRelationshipRequest()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/relationships/invalid-id", updateRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Update()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid relationship ID format")
}

// Test Update - Relationship Not Found
func (suite *RelationshipHandlerTestSuite) TestUpdate_RelationshipNotFound() {
	relationshipID := uuid.New()
	updateRequest := createValidUpdateRelationshipRequest()

	suite.mockRelationshipService.On("UpdateRelationship", relationshipID, &updateRequest, mock.AnythingOfType("uuid.UUID")).Return(nil, services.ErrRelationshipNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/relationships/"+relationshipID.String(), updateRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Update()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Relationship not found")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Delete - Success
func (suite *RelationshipHandlerTestSuite) TestDelete_Success() {
	relationshipID := uuid.New()
	userID := uuid.New()

	suite.mockRelationshipService.On("DeleteRelationship", relationshipID, userID).Return(nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/relationships/"+relationshipID.String(), nil)
	req = testutil.WithUserContext(req, userID) // Add user context

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Relationship deleted successfully")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Delete - Invalid Relationship ID
func (suite *RelationshipHandlerTestSuite) TestDelete_InvalidRelationshipID() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/relationships/invalid-id", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid relationship ID format")
}

// Test Delete - Relationship Not Found
func (suite *RelationshipHandlerTestSuite) TestDelete_RelationshipNotFound() {
	relationshipID := uuid.New()
	userID := uuid.New()

	suite.mockRelationshipService.On("DeleteRelationship", relationshipID, userID).Return(services.ErrRelationshipNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/relationships/"+relationshipID.String(), nil)
	req = testutil.WithUserContext(req, userID) // Add user context

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Relationship not found")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Delete - Service Error
func (suite *RelationshipHandlerTestSuite) TestDelete_ServiceError() {
	relationshipID := uuid.New()
	userID := uuid.New()

	suite.mockRelationshipService.On("DeleteRelationship", relationshipID, userID).Return(assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/relationships/"+relationshipID.String(), nil)
	req = testutil.WithUserContext(req, userID) // Add user context

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockRelationshipService.AssertExpectations(suite.T())
}

// Test Delete - No User Context
func (suite *RelationshipHandlerTestSuite) TestDelete_NoUserContext() {
	relationshipID := uuid.New()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/relationships/"+relationshipID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("relationship_id", relationshipID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "User context not found")
}
