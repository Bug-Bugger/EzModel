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
	"github.com/stretchr/testify/suite"
)

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

func createValidFieldRequest() dto.CreateFieldRequest {
	return dto.CreateFieldRequest{
		Name:         "test_field",
		DataType:     "VARCHAR(255)",
		IsPrimaryKey: false,
		IsNullable:   true,
		DefaultValue: "",
		Position:     1,
	}
}

func createValidUpdateFieldRequest() dto.UpdateFieldRequest {
	name := "updated_field"
	dataType := "INT"
	isPrimaryKey := true
	return dto.UpdateFieldRequest{
		Name:         &name,
		DataType:     &dataType,
		IsPrimaryKey: &isPrimaryKey,
	}
}

type FieldHandlerTestSuite struct {
	suite.Suite
	mockFieldService *mockService.MockFieldService
	handler          *FieldHandler
}

func (suite *FieldHandlerTestSuite) SetupTest() {
	suite.mockFieldService = new(mockService.MockFieldService)
	suite.handler = NewFieldHandler(suite.mockFieldService)
}

func TestFieldHandlerSuite(t *testing.T) {
	suite.Run(t, new(FieldHandlerTestSuite))
}

// Test Create - Success
func (suite *FieldHandlerTestSuite) TestCreate_Success() {
	tableID := uuid.New()
	fieldRequest := createValidFieldRequest()
	field := createTestField(tableID)

	suite.mockFieldService.On("CreateField", tableID, &fieldRequest).Return(field, nil)

	// Create request with table_id in URL params
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields", fieldRequest)

	// Add URL params using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusCreated, "Field created successfully")

	fieldResponse, ok := response.Data.(map[string]any)
	suite.True(ok, "Response data should be a field object")
	suite.Equal(field.ID.String(), fieldResponse["id"])
	suite.Equal(field.Name, fieldResponse["name"])
	suite.Equal(field.DataType, fieldResponse["data_type"])

	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Create - Invalid Table ID
func (suite *FieldHandlerTestSuite) TestCreate_InvalidTableID() {
	fieldRequest := createValidFieldRequest()

	// Create request with invalid table_id in URL params
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/invalid-id/fields", fieldRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid table ID format")
}

// Test Create - Invalid JSON
func (suite *FieldHandlerTestSuite) TestCreate_InvalidJSON() {
	tableID := uuid.New()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields", "invalid json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid request body")
}

// Test Create - Validation Error
func (suite *FieldHandlerTestSuite) TestCreate_ValidationError() {
	tableID := uuid.New()
	invalidRequest := dto.CreateFieldRequest{
		Name:     "", // Invalid: empty name
		DataType: "VARCHAR(255)",
	}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields", invalidRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	// Should return validation errors (400 status)
	suite.Equal(http.StatusBadRequest, w.Code)
}

// Test Create - Table Not Found
func (suite *FieldHandlerTestSuite) TestCreate_TableNotFound() {
	tableID := uuid.New()
	fieldRequest := createValidFieldRequest()

	suite.mockFieldService.On("CreateField", tableID, &fieldRequest).Return(nil, services.ErrTableNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields", fieldRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Table not found")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Create - Invalid Input
func (suite *FieldHandlerTestSuite) TestCreate_InvalidInput() {
	tableID := uuid.New()
	fieldRequest := createValidFieldRequest()

	suite.mockFieldService.On("CreateField", tableID, &fieldRequest).Return(nil, services.ErrInvalidInput)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields", fieldRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid input")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Create - Service Error
func (suite *FieldHandlerTestSuite) TestCreate_ServiceError() {
	tableID := uuid.New()
	fieldRequest := createValidFieldRequest()

	suite.mockFieldService.On("CreateField", tableID, &fieldRequest).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields", fieldRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test GetByID - Success
func (suite *FieldHandlerTestSuite) TestGetByID_Success() {
	fieldID := uuid.New()
	field := createTestField(uuid.New())
	field.ID = fieldID

	suite.mockFieldService.On("GetFieldByID", fieldID).Return(field, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/fields/"+fieldID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Field retrieved successfully")

	fieldResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(field.ID.String(), fieldResponse["id"])
	suite.Equal(field.Name, fieldResponse["name"])

	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test GetByID - Invalid Field ID
func (suite *FieldHandlerTestSuite) TestGetByID_InvalidFieldID() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/fields/invalid-id", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid field ID format")
}

// Test GetByID - Field Not Found
func (suite *FieldHandlerTestSuite) TestGetByID_FieldNotFound() {
	fieldID := uuid.New()

	suite.mockFieldService.On("GetFieldByID", fieldID).Return(nil, services.ErrFieldNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/fields/"+fieldID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Field not found")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test GetByID - Service Error
func (suite *FieldHandlerTestSuite) TestGetByID_ServiceError() {
	fieldID := uuid.New()

	suite.mockFieldService.On("GetFieldByID", fieldID).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/fields/"+fieldID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test GetByTableID - Success
func (suite *FieldHandlerTestSuite) TestGetByTableID_Success() {
	tableID := uuid.New()
	fields := []*models.Field{
		createTestField(tableID),
		createTestField(tableID),
	}

	suite.mockFieldService.On("GetFieldsByTableID", tableID).Return(fields, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/tables/"+tableID.String()+"/fields", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByTableID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Fields retrieved successfully")

	fieldResponses, ok := response.Data.([]any)
	suite.True(ok)
	suite.Len(fieldResponses, 2)

	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test GetByTableID - Invalid Table ID
func (suite *FieldHandlerTestSuite) TestGetByTableID_InvalidTableID() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/tables/invalid-id/fields", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByTableID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid table ID format")
}

// Test GetByTableID - Service Error
func (suite *FieldHandlerTestSuite) TestGetByTableID_ServiceError() {
	tableID := uuid.New()

	suite.mockFieldService.On("GetFieldsByTableID", tableID).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodGet, "/tables/"+tableID.String()+"/fields", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.GetByTableID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Internal server error")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Update - Success
func (suite *FieldHandlerTestSuite) TestUpdate_Success() {
	fieldID := uuid.New()
	updateRequest := createValidUpdateFieldRequest()
	updatedField := createTestField(uuid.New())
	updatedField.ID = fieldID
	updatedField.Name = *updateRequest.Name
	updatedField.DataType = *updateRequest.DataType
	updatedField.IsPrimaryKey = *updateRequest.IsPrimaryKey

	suite.mockFieldService.On("UpdateField", fieldID, &updateRequest).Return(updatedField, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/fields/"+fieldID.String(), updateRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Update()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Field updated successfully")

	fieldResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(updatedField.ID.String(), fieldResponse["id"])
	suite.Equal(updatedField.Name, fieldResponse["name"])
	suite.Equal(updatedField.DataType, fieldResponse["data_type"])

	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Update - Invalid Field ID
func (suite *FieldHandlerTestSuite) TestUpdate_InvalidFieldID() {
	updateRequest := createValidUpdateFieldRequest()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/fields/invalid-id", updateRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Update()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid field ID format")
}

// Test Update - Field Not Found
func (suite *FieldHandlerTestSuite) TestUpdate_FieldNotFound() {
	fieldID := uuid.New()
	updateRequest := createValidUpdateFieldRequest()

	suite.mockFieldService.On("UpdateField", fieldID, &updateRequest).Return(nil, services.ErrFieldNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/fields/"+fieldID.String(), updateRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Update()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Field not found")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Reorder - Success
func (suite *FieldHandlerTestSuite) TestReorder_Success() {
	tableID := uuid.New()
	fieldPositions := map[uuid.UUID]int{
		uuid.New(): 1,
		uuid.New(): 2,
	}
	reorderRequest := dto.ReorderFieldsRequest{
		FieldPositions: fieldPositions,
	}

	suite.mockFieldService.On("ReorderFields", tableID, fieldPositions).Return(nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields/reorder", reorderRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Reorder()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Fields reordered successfully")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Reorder - Invalid Table ID
func (suite *FieldHandlerTestSuite) TestReorder_InvalidTableID() {
	fieldPositions := map[uuid.UUID]int{
		uuid.New(): 1,
	}
	reorderRequest := dto.ReorderFieldsRequest{
		FieldPositions: fieldPositions,
	}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/invalid-id/fields/reorder", reorderRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Reorder()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid table ID format")
}

// Test Reorder - Table Not Found
func (suite *FieldHandlerTestSuite) TestReorder_TableNotFound() {
	tableID := uuid.New()
	fieldPositions := map[uuid.UUID]int{
		uuid.New(): 1,
	}
	reorderRequest := dto.ReorderFieldsRequest{
		FieldPositions: fieldPositions,
	}

	suite.mockFieldService.On("ReorderFields", tableID, fieldPositions).Return(services.ErrTableNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/tables/"+tableID.String()+"/fields/reorder", reorderRequest)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("table_id", tableID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Reorder()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Table not found")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Delete - Success
func (suite *FieldHandlerTestSuite) TestDelete_Success() {
	fieldID := uuid.New()
	userID := uuid.New()

	suite.mockFieldService.On("DeleteField", fieldID).Return(nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/fields/"+fieldID.String(), nil)
	req = testutil.WithUserContext(req, userID) // Add user context

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Field deleted successfully")
	suite.mockFieldService.AssertExpectations(suite.T())
}

// Test Delete - No User Context
func (suite *FieldHandlerTestSuite) TestDelete_NoUserContext() {
	fieldID := uuid.New()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/fields/"+fieldID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "User context not found")
}

// Test Delete - Invalid Field ID
func (suite *FieldHandlerTestSuite) TestDelete_InvalidFieldID() {
	userID := uuid.New()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/fields/invalid-id", nil)
	req = testutil.WithUserContext(req, userID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid field ID format")
}

// Test Delete - Field Not Found
func (suite *FieldHandlerTestSuite) TestDelete_FieldNotFound() {
	fieldID := uuid.New()
	userID := uuid.New()

	suite.mockFieldService.On("DeleteField", fieldID).Return(services.ErrFieldNotFound)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodDelete, "/fields/"+fieldID.String(), nil)
	req = testutil.WithUserContext(req, userID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("field_id", fieldID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Field not found")
	suite.mockFieldService.AssertExpectations(suite.T())
}