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

type ProjectHandlerTestSuite struct {
	suite.Suite
	mockService *mockService.MockProjectService
	handler     *ProjectHandler
	userID      uuid.UUID
}

func (suite *ProjectHandlerTestSuite) SetupTest() {
	suite.mockService = new(mockService.MockProjectService)
	suite.handler = NewProjectHandler(suite.mockService)
	suite.userID = uuid.New()
}

func TestProjectHandlerSuite(t *testing.T) {
	suite.Run(t, new(ProjectHandlerTestSuite))
}

// Test Create Project - Success
func (suite *ProjectHandlerTestSuite) TestCreateProject_Success() {
	// Setup
	requestBody := testutil.CreateValidProjectRequest()
	expectedProject := testutil.CreateTestProject(suite.userID)
	expectedProject.Name = requestBody.Name
	expectedProject.Description = requestBody.Description

	suite.mockService.On("CreateProject", requestBody.Name, requestBody.Description, suite.userID).
		Return(expectedProject, nil)

	// Make request with user context
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects", requestBody)
	req = testutil.WithUserContext(req, suite.userID)
	w := httptest.NewRecorder()

	// Execute
	suite.handler.Create()(w, req)

	// Assert
	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusCreated, "Project created successfully")

	projectResponse, ok := response.Data.(map[string]any)
	suite.True(ok, "Response data should be a project object")
	suite.Equal(expectedProject.Name, projectResponse["name"])
	suite.Equal(expectedProject.Description, projectResponse["description"])
	suite.Equal(expectedProject.OwnerID.String(), projectResponse["owner_id"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Create Project - Missing User Context
func (suite *ProjectHandlerTestSuite) TestCreateProject_MissingUserContext() {
	requestBody := testutil.CreateValidProjectRequest()

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects", requestBody)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "User context not found")
}

// Test Create Project - Invalid JSON
func (suite *ProjectHandlerTestSuite) TestCreateProject_InvalidJSON() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects", "invalid json")
	req = testutil.WithUserContext(req, suite.userID)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid request body")
}

// Test Create Project - Validation Error
func (suite *ProjectHandlerTestSuite) TestCreateProject_ValidationError() {
	invalidRequest := dto.CreateProjectRequest{
		Name:        "", // Required field empty
		Description: "Valid description",
	}

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects", invalidRequest)
	req = testutil.WithUserContext(req, suite.userID)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	response := testutil.AssertJSONResponse(suite.T(), w, http.StatusBadRequest)
	suite.False(response.Success)
	suite.NotNil(response.Errors)
}

// Test Create Project - Service Error
func (suite *ProjectHandlerTestSuite) TestCreateProject_ServiceError() {
	requestBody := testutil.CreateValidProjectRequest()

	suite.mockService.On("CreateProject", requestBody.Name, requestBody.Description, suite.userID).
		Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects", requestBody)
	req = testutil.WithUserContext(req, suite.userID)
	w := httptest.NewRecorder()

	suite.handler.Create()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Failed to create project")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Get Project By ID - Success
func (suite *ProjectHandlerTestSuite) TestGetProjectByID_Success() {
	projectID := uuid.New()
	expectedProject := testutil.CreateTestProject(suite.userID)
	expectedProject.ID = projectID

	suite.mockService.On("GetProjectByID", projectID).Return(expectedProject, nil)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()

	// Setup chi context with URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Project retrieved successfully")

	projectResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(expectedProject.Name, projectResponse["name"])
	suite.Equal(expectedProject.Description, projectResponse["description"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Get Project By ID - Invalid UUID
func (suite *ProjectHandlerTestSuite) TestGetProjectByID_InvalidUUID() {
	req := httptest.NewRequest(http.MethodGet, "/projects/invalid-uuid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid project ID")
}

// Test Get Project By ID - Not Found
func (suite *ProjectHandlerTestSuite) TestGetProjectByID_NotFound() {
	projectID := uuid.New()

	suite.mockService.On("GetProjectByID", projectID).Return(nil, services.ErrProjectNotFound)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.GetByID()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Project not found")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Get My Projects - Success
func (suite *ProjectHandlerTestSuite) TestGetMyProjects_Success() {
	ownedProjects := []*models.Project{
		testutil.CreateTestProject(suite.userID),
		testutil.CreateTestProject(suite.userID),
	}
	collaboratedProjects := []*models.Project{
		testutil.CreateTestProject(uuid.New()), // Different owner
	}

	suite.mockService.On("GetProjectsByOwnerID", suite.userID).Return(ownedProjects, nil)
	suite.mockService.On("GetProjectsByCollaboratorID", suite.userID).Return(collaboratedProjects, nil)

	req := httptest.NewRequest(http.MethodGet, "/projects/my", nil)
	req = testutil.WithUserContext(req, suite.userID)
	w := httptest.NewRecorder()

	suite.handler.GetMyProjects()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "My projects retrieved successfully")

	projectsResponse, ok := response.Data.([]any)
	suite.True(ok)
	suite.Len(projectsResponse, 3) // 2 owned + 1 collaborated

	suite.mockService.AssertExpectations(suite.T())
}

// Test Update Project - Success
func (suite *ProjectHandlerTestSuite) TestUpdateProject_Success() {
	projectID := uuid.New()
	newName := "Updated Project"
	newDescription := "Updated description"
	updateRequest := dto.UpdateProjectRequest{
		Name:        &newName,
		Description: &newDescription,
	}

	updatedProject := testutil.CreateTestProject(suite.userID)
	updatedProject.ID = projectID
	updatedProject.Name = newName
	updatedProject.Description = newDescription

	suite.mockService.On("UpdateProject", projectID, &updateRequest).Return(updatedProject, nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPut, "/projects/"+projectID.String(), updateRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Update()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Project updated successfully")

	projectResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(updatedProject.Name, projectResponse["name"])
	suite.Equal(updatedProject.Description, projectResponse["description"])

	suite.mockService.AssertExpectations(suite.T())
}

// Test Delete Project - Success
func (suite *ProjectHandlerTestSuite) TestDeleteProject_Success() {
	projectID := uuid.New()

	suite.mockService.On("DeleteProject", projectID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.Delete()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Project deleted successfully")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Add Collaborator - Success
func (suite *ProjectHandlerTestSuite) TestAddCollaborator_Success() {
	projectID := uuid.New()
	collaboratorID := uuid.New()
	addRequest := dto.AddCollaboratorRequest{
		CollaboratorID: collaboratorID,
	}

	suite.mockService.On("AddCollaborator", projectID, collaboratorID).Return(nil)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/projects/"+projectID.String()+"/collaborators", addRequest)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.AddCollaborator()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Collaborator added successfully")
	suite.mockService.AssertExpectations(suite.T())
}

// Test Remove Collaborator - Success
func (suite *ProjectHandlerTestSuite) TestRemoveCollaborator_Success() {
	projectID := uuid.New()
	collaboratorID := uuid.New()

	suite.mockService.On("RemoveCollaborator", projectID, collaboratorID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/projects/"+projectID.String()+"/collaborators/"+collaboratorID.String(), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", projectID.String())
	rctx.URLParams.Add("user_id", collaboratorID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	suite.handler.RemoveCollaborator()(w, req)

	testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Collaborator removed successfully")
	suite.mockService.AssertExpectations(suite.T())
}
