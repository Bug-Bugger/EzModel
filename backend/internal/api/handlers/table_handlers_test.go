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
	"github.com/stretchr/testify/suite"
)

type TableHandlerTestSuite struct {
	suite.Suite
	mockService *mockService.MockTableService
	handler     *TableHandler
}

func (suite *TableHandlerTestSuite) SetupTest() {
	suite.mockService = new(mockService.MockTableService)
	suite.handler = NewTableHandler(suite.mockService)
}

func TestTableHandlerSuite(t *testing.T) {
	suite.Run(t, new(TableHandlerTestSuite))
}

// Test Create Table - Success
func (suite *TableHandlerTestSuite) TestCreateTable_Success() {
	projectID := uuid.New()
	requestBody := testutil.CreateValidTableRequest()
	expectedTable := testutil.CreateTestTable(projectID)

	suite.mockService.On("CreateTable", projectID, requestBody.Name, requestBody.PosX, requestBody.PosY).
		Return(expectedTable, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/tables", requestBody)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Create()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusCreated, "Table created successfully")

	tableResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(expectedTable.Name, tableResponse["name"])
	suite.Equal(expectedTable.PosX, tableResponse["pos_x"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Create Table - Invalid Project ID
func (suite *TableHandlerTestSuite) TestCreateTable_InvalidProjectID() {
	requestBody := testutil.CreateValidTableRequest()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/invalid-uuid/tables", requestBody)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid project ID format")
}

// Test Create Table - Validation Error
func (suite *TableHandlerTestSuite) TestCreateTable_ValidationError() {
	projectID := uuid.New()
	invalidRequest := dto.CreateTableRequest{
		Name: "", // Required field empty
		PosX: 100.0,
		PosY: 200.0,
	}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/tables", invalidRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Create()(w, req)

	response := testutil.AssertJSONResponse(suite.T(), w, http.StatusBadRequest)
	suite.False(response.Success)
}

// Test Get Table By ID - Success
func (suite *TableHandlerTestSuite) TestGetTableByID_Success() {
	tableID := uuid.New()
	expectedTable := testutil.CreateTestTable(uuid.New())
	expectedTable.ID = tableID

	suite.mockService.On("GetTableByID", tableID).Return(expectedTable, nil)

	req := httptest.NewRequest(http.MethodGet, "/tables/"+tableID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Table retrieved successfully")

	tableResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(expectedTable.Name, tableResponse["name"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Get Tables By Project ID - Success
func (suite *TableHandlerTestSuite) TestGetTablesByProjectID_Success() {
	projectID := uuid.New()
	tables := []*models.Table{
		testutil.CreateTestTable(projectID),
		testutil.CreateTestTable(projectID),
	}

	suite.mockService.On("GetTablesByProjectID", projectID).Return(tables, nil)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/tables", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByProjectID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Tables retrieved successfully")

	tablesResponse, ok := response.Data.([]any)
	suite.True(ok)
	suite.Len(tablesResponse, 2)

	suite.mockService.AssertExpectations(suite.T())
}

// Test Update Table - Success
func (suite *TableHandlerTestSuite) TestUpdateTable_Success() {
	tableID := uuid.New()
	newName := "Updated Table"
	updateRequest := dto.UpdateTableRequest{
		Name: &newName,
	}

	updatedTable := testutil.CreateTestTable(uuid.New())
	updatedTable.ID = tableID
	updatedTable.Name = newName

	suite.mockService.On("UpdateTable", tableID, &updateRequest).Return(updatedTable, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/tables/"+tableID.String(), updateRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Update()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Table updated successfully")

	tableResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(updatedTable.Name, tableResponse["name"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Update Table Position - Success
func (suite *TableHandlerTestSuite) TestUpdateTablePosition_Success() {
	tableID := uuid.New()
	positionRequest := dto.UpdateTablePositionRequest{
		PosX: 300.0,
		PosY: 400.0,
	}

	suite.mockService.On("UpdateTablePosition", tableID, positionRequest.PosX, positionRequest.PosY).Return(nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/tables/"+tableID.String()+"/position", positionRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.UpdatePosition()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Table position updated successfully")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Delete Table - Success
func (suite *TableHandlerTestSuite) TestDeleteTable_Success() {
	tableID := uuid.New()
	userID := uuid.New()

	suite.mockService.On("DeleteTable", tableID, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/tables/"+tableID.String(), nil)
	req = testutil.WithUserContext(req, userID)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Delete()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Table deleted successfully")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Delete Table - Not Found
func (suite *TableHandlerTestSuite) TestDeleteTable_NotFound() {
	tableID := uuid.New()
	userID := uuid.New()

	suite.mockService.On("DeleteTable", tableID, userID).Return(services.ErrTableNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/tables/"+tableID.String(), nil)
	req = testutil.WithUserContext(req, userID)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Table not found")
	suite.mockService.AssertExpectations(suite.T())
}
