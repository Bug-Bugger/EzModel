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

type UserHandlerTestSuite struct {
	suite.Suite
	mockService *mockService.MockUserService
	handler     *UserHandler
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.mockService = new(mockService.MockUserService)
	suite.handler = NewUserHandler(suite.mockService)
}

func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

// Test Create User - Success
func (suite *UserHandlerTestSuite) TestCreateUser_Success() {
	// Setup
	requestBody := testutil.CreateValidUserRequest()
	expectedUser := testutil.CreateTestUserWithData(requestBody.Email, requestBody.Username)

	suite.mockService.On("CreateUser", requestBody.Email, requestBody.Username, requestBody.Password).
		Return(expectedUser, nil)

	// Make request
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/users", requestBody)
	w := httptest.NewRecorder()

	// Execute
	suite.handler.Create()(w, req)

	// Assert
	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusCreated, "User created successfully")

	// Verify response data
	userResponse, ok := response.Data.(map[string]any)
	suite.True(ok, "Response data should be a user object")
	suite.Equal(expectedUser.Email, userResponse["email"])
	suite.Equal(expectedUser.Username, userResponse["username"])
	suite.NotNil(userResponse["id"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Create User - Invalid JSON
func (suite *UserHandlerTestSuite) TestCreateUser_InvalidJSON() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/users", "invalid json")
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid request body")
}

// Test Create User - Validation Error
func (suite *UserHandlerTestSuite) TestCreateUser_ValidationError() {
	invalidRequest := dto.CreateUserRequest{
		Email:    "invalid-email",
		Username: "ab", // too short
		Password: "123", // too short
	}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/users", invalidRequest)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	response := testutil.AssertJSONResponse(suite.T(), w, http.StatusBadRequest)
	suite.False(response.Success)
	suite.NotNil(response.Errors)
}

// Test Create User - User Already Exists
func (suite *UserHandlerTestSuite) TestCreateUser_UserAlreadyExists() {
	requestBody := testutil.CreateValidUserRequest()

	suite.mockService.On("CreateUser", requestBody.Email, requestBody.Username, requestBody.Password).
		Return(nil, services.ErrUserAlreadyExists)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/users", requestBody)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusConflict, "User already exists")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Create User - Service Error
func (suite *UserHandlerTestSuite) TestCreateUser_ServiceError() {
	requestBody := testutil.CreateValidUserRequest()

	suite.mockService.On("CreateUser", requestBody.Email, requestBody.Username, requestBody.Password).
		Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/users", requestBody)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Failed to create user")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Get User By ID - Success
func (suite *UserHandlerTestSuite) TestGetUserByID_Success() {
	userID := uuid.New()
	expectedUser := testutil.CreateTestUser()
	expectedUser.ID = userID

	suite.mockService.On("GetUserByID", userID).Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	// Setup chi context with URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "User retrieved successfully")

	userResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(expectedUser.Email, userResponse["email"])
	suite.Equal(expectedUser.Username, userResponse["username"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Get User By ID - Invalid UUID
func (suite *UserHandlerTestSuite) TestGetUserByID_InvalidUUID() {
	req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid user ID")
}

// Test Get User By ID - Not Found
func (suite *UserHandlerTestSuite) TestGetUserByID_NotFound() {
	userID := uuid.New()

	suite.mockService.On("GetUserByID", userID).Return(nil, services.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "User not found")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Update User - Success
func (suite *UserHandlerTestSuite) TestUpdateUser_Success() {
	userID := uuid.New()
	updateRequest := testutil.CreateValidUpdateUserRequest()
	updatedUser := testutil.CreateTestUser()
	updatedUser.ID = userID
	updatedUser.Username = *updateRequest.Username
	updatedUser.Email = *updateRequest.Email

	suite.mockService.On("UpdateUser", userID, &updateRequest).Return(updatedUser, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/users/"+userID.String(), updateRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Update()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "User updated successfully")

	userResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(updatedUser.Email, userResponse["email"])
	suite.Equal(updatedUser.Username, userResponse["username"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Update User - Empty Request
func (suite *UserHandlerTestSuite) TestUpdateUser_EmptyRequest() {
	userID := uuid.New()
	emptyRequest := dto.UpdateUserRequest{}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/users/"+userID.String(), emptyRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Update()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "No fields to update provided")
}

// Test Update Password - Success
func (suite *UserHandlerTestSuite) TestUpdatePassword_Success() {
	userID := uuid.New()
	passwordRequest := dto.UpdatePasswordRequest{Password: "newpassword123"}

	suite.mockService.On("UpdatePassword", userID, passwordRequest.Password).Return(nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/users/"+userID.String()+"/password", passwordRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.UpdatePassword()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Password updated successfully")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Get All Users - Success
func (suite *UserHandlerTestSuite) TestGetAllUsers_Success() {
	// Create real test users
	realUsers := []*models.User{
		testutil.CreateTestUser(),
		testutil.CreateTestUserWithData("user2@test.com", "user2"),
	}

	suite.mockService.On("GetAllUsers").Return(realUsers, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	suite.handler.GetAll()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Users retrieved successfully")

	usersResponse, ok := response.Data.([]any)
	suite.True(ok)
	suite.Len(usersResponse, 2)

	suite.mockService.AssertExpectations(suite.T())
}

// Test Delete User - Success
func (suite *UserHandlerTestSuite) TestDeleteUser_Success() {
	userID := uuid.New()

	suite.mockService.On("DeleteUser", userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Delete()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "User deleted successfully")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Delete User - Not Found
func (suite *UserHandlerTestSuite) TestDeleteUser_NotFound() {
	userID := uuid.New()

	suite.mockService.On("DeleteUser", userID).Return(services.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Delete()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "User not found")
	suite.mockService.AssertExpectations(suite.T())
}