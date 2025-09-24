package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// CreateTestUser creates a test user with default values
func CreateTestUser() *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword123",
	}
}

// CreateTestUserWithData creates a test user with custom data
func CreateTestUserWithData(email, username string) *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: "hashedpassword123",
	}
}

// CreateTestProject creates a test project with default values
func CreateTestProject(ownerID uuid.UUID) *models.Project {
	return &models.Project{
		ID:           uuid.New(),
		Name:         "Test Project",
		Description:  "A test project",
		OwnerID:      ownerID,
		DatabaseType: "postgresql",
		CanvasData:   "{}",
	}
}

// MakeJSONRequest creates a test HTTP request with JSON body
func MakeJSONRequest(t *testing.T, method, url string, body any) *http.Request {
	var reqBody *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(jsonBody)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ParseJSONResponse parses the JSON response from a ResponseRecorder
func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, v any) {
	err := json.Unmarshal(w.Body.Bytes(), v)
	require.NoError(t, err, "Failed to parse JSON response: %s", w.Body.String())
}

// AssertJSONResponse asserts the response status and parses JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) dto.APIResponse {
	require.Equal(t, expectedStatus, w.Code, "Response body: %s", w.Body.String())

	var response dto.APIResponse
	ParseJSONResponse(t, w, &response)
	return response
}

// AssertSuccessResponse asserts a successful API response
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) dto.APIResponse {
	response := AssertJSONResponse(t, w, expectedStatus)
	require.True(t, response.Success, "Expected success response")
	if expectedMessage != "" {
		require.Equal(t, expectedMessage, response.Message)
	}
	return response
}

// AssertErrorResponse asserts an error API response
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) dto.APIResponse {
	response := AssertJSONResponse(t, w, expectedStatus)
	require.False(t, response.Success, "Expected error response")
	if expectedMessage != "" {
		require.Equal(t, expectedMessage, response.Message)
	}
	return response
}

// StringPtr returns a pointer to a string (useful for optional fields)
func StringPtr(s string) *string {
	return &s
}

// CreateValidUserRequest creates a valid user creation request
func CreateValidUserRequest() dto.CreateUserRequest {
	return dto.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
}

// CreateValidUpdateUserRequest creates a valid user update request
func CreateValidUpdateUserRequest() dto.UpdateUserRequest {
	return dto.UpdateUserRequest{
		Username: StringPtr("updateduser"),
		Email:    StringPtr("updated@example.com"),
	}
}

// WithUserContext adds user ID to request context (for auth middleware simulation)
func WithUserContext(req *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), "userID", userID.String())
	return req.WithContext(ctx)
}

// CreateValidProjectRequest creates a valid project creation request
func CreateValidProjectRequest() dto.CreateProjectRequest {
	return dto.CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
	}
}

// CreateTestTable creates a test table with default values
func CreateTestTable(projectID uuid.UUID) *models.Table {
	return &models.Table{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      "Test Table",
		PosX:      100.0,
		PosY:      200.0,
	}
}

// CreateValidTableRequest creates a valid table creation request
func CreateValidTableRequest() dto.CreateTableRequest {
	return dto.CreateTableRequest{
		Name: "Test Table",
		PosX: 100.0,
		PosY: 200.0,
	}
}