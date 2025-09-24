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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Test helper functions
func createTestUser() *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword123",
	}
}

func createTestUserWithData(email, username string) *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: "hashedpassword123",
	}
}

func stringPtr(s string) *string {
	return &s
}

type UserServiceTestSuite struct {
	suite.Suite
	mockRepo *mockRepo.MockUserRepository
	service  *UserService
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mockRepo.MockUserRepository)
	suite.service = NewUserService(suite.mockRepo)
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// Test CreateUser - Success
func (suite *UserServiceTestSuite) TestCreateUser_Success() {
	email := "test@example.com"
	username := "testuser"
	password := "password123"
	userID := uuid.New()

	// Mock that email doesn't exist
	suite.mockRepo.On("GetByEmail", email).Return(nil, gorm.ErrRecordNotFound)

	// Mock successful creation
	suite.mockRepo.On("Create", mock.MatchedBy(func(user *models.User) bool {
		return user.Email == email && user.Username == username && user.PasswordHash != ""
	})).Return(userID, nil)

	// Execute
	result, err := suite.service.CreateUser(email, username, password)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(userID, result.ID)
	suite.Equal(email, result.Email)
	suite.Equal(username, result.Username)
	suite.NotEmpty(result.PasswordHash)

	// Verify password was hashed correctly
	err = bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte(password))
	suite.NoError(err, "Password should be hashed correctly")

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test CreateUser - Invalid Input (short email)
func (suite *UserServiceTestSuite) TestCreateUser_InvalidEmail() {
	result, err := suite.service.CreateUser("a@b", "testuser", "password123")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateUser - Invalid Input (short username)
func (suite *UserServiceTestSuite) TestCreateUser_InvalidUsername() {
	result, err := suite.service.CreateUser("test@example.com", "ab", "password123")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateUser - Invalid Input (short password)
func (suite *UserServiceTestSuite) TestCreateUser_InvalidPassword() {
	result, err := suite.service.CreateUser("test@example.com", "testuser", "12345")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidInput, err)
}

// Test CreateUser - User Already Exists
func (suite *UserServiceTestSuite) TestCreateUser_UserAlreadyExists() {
	email := "test@example.com"
	username := "testuser"
	password := "password123"

	existingUser := createTestUserWithData(email, "existinguser")
	suite.mockRepo.On("GetByEmail", email).Return(existingUser, nil)

	result, err := suite.service.CreateUser(email, username, password)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrUserAlreadyExists, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test CreateUser - Repository Error on Email Check
func (suite *UserServiceTestSuite) TestCreateUser_RepositoryErrorOnEmailCheck() {
	email := "test@example.com"
	username := "testuser"
	password := "password123"

	suite.mockRepo.On("GetByEmail", email).Return(nil, assert.AnError)

	result, err := suite.service.CreateUser(email, username, password)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test CreateUser - Repository Error on Create
func (suite *UserServiceTestSuite) TestCreateUser_RepositoryErrorOnCreate() {
	email := "test@example.com"
	username := "testuser"
	password := "password123"

	suite.mockRepo.On("GetByEmail", email).Return(nil, gorm.ErrRecordNotFound)
	suite.mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(uuid.Nil, assert.AnError)

	result, err := suite.service.CreateUser(email, username, password)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByID - Success
func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
	userID := uuid.New()
	expectedUser := createTestUser()
	expectedUser.ID = userID

	suite.mockRepo.On("GetByID", userID).Return(expectedUser, nil)

	result, err := suite.service.GetUserByID(userID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedUser.ID, result.ID)
	suite.Equal(expectedUser.Email, result.Email)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByID - Not Found
func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
	userID := uuid.New()

	suite.mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetUserByID(userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrUserNotFound, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByEmail - Success
func (suite *UserServiceTestSuite) TestGetUserByEmail_Success() {
	email := "test@example.com"
	expectedUser := createTestUserWithData(email, "testuser")

	suite.mockRepo.On("GetByEmail", email).Return(expectedUser, nil)

	result, err := suite.service.GetUserByEmail(email)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedUser.Email, result.Email)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdateUser - Success
func (suite *UserServiceTestSuite) TestUpdateUser_Success() {
	userID := uuid.New()
	existingUser := createTestUser()
	existingUser.ID = userID

	newUsername := "updateduser"
	newEmail := "updated@example.com"
	updateRequest := &dto.UpdateUserRequest{
		Username: &newUsername,
		Email:    &newEmail,
	}

	suite.mockRepo.On("GetByID", userID).Return(existingUser, nil)
	// Check that new email doesn't exist
	suite.mockRepo.On("GetByEmail", newEmail).Return(nil, gorm.ErrRecordNotFound)
	suite.mockRepo.On("Update", mock.MatchedBy(func(user *models.User) bool {
		return user.ID == userID && user.Username == newUsername && user.Email == newEmail
	})).Return(nil)

	result, err := suite.service.UpdateUser(userID, updateRequest)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(userID, result.ID)
	suite.Equal(newUsername, result.Username)
	suite.Equal(newEmail, result.Email)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdateUser - User Not Found
func (suite *UserServiceTestSuite) TestUpdateUser_UserNotFound() {
	userID := uuid.New()
	updateRequest := &dto.UpdateUserRequest{
		Username: stringPtr("newusername"),
	}

	suite.mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateUser(userID, updateRequest)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrUserNotFound, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdateUser - Email Already Exists
func (suite *UserServiceTestSuite) TestUpdateUser_EmailAlreadyExists() {
	userID := uuid.New()
	existingUser := createTestUser()
	existingUser.ID = userID

	newEmail := "taken@example.com"
	updateRequest := &dto.UpdateUserRequest{
		Email: &newEmail,
	}

	anotherUser := createTestUserWithData(newEmail, "anotheruser")
	anotherUser.ID = uuid.New() // Different ID

	suite.mockRepo.On("GetByID", userID).Return(existingUser, nil)
	suite.mockRepo.On("GetByEmail", newEmail).Return(anotherUser, nil)

	result, err := suite.service.UpdateUser(userID, updateRequest)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrUserAlreadyExists, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdatePassword - Success
func (suite *UserServiceTestSuite) TestUpdatePassword_Success() {
	userID := uuid.New()
	currentPassword := "password123"
	newPassword := "newpassword123"
	existingUser := createTestUser()
	existingUser.ID = userID
	// Hash the current password for the user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	existingUser.PasswordHash = string(hashedPassword)

	suite.mockRepo.On("GetByID", userID).Return(existingUser, nil)
	suite.mockRepo.On("Update", mock.MatchedBy(func(user *models.User) bool {
		// Verify new password was hashed
		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword))
		return err == nil
	})).Return(nil)

	err := suite.service.UpdatePassword(userID, currentPassword, newPassword)

	suite.NoError(err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdatePassword - Invalid Password
func (suite *UserServiceTestSuite) TestUpdatePassword_InvalidPassword() {
	userID := uuid.New()
	currentPassword := "password123"
	shortPassword := "12345"

	err := suite.service.UpdatePassword(userID, currentPassword, shortPassword)

	suite.Error(err)
	suite.Equal(ErrInvalidInput, err)
}

// Test UpdatePassword - Invalid Current Password
func (suite *UserServiceTestSuite) TestUpdatePassword_InvalidCurrentPassword() {
	userID := uuid.New()
	currentPassword := "password123"
	wrongCurrentPassword := "wrongpassword"
	newPassword := "newpassword123"
	existingUser := createTestUser()
	existingUser.ID = userID
	// Hash the correct current password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	existingUser.PasswordHash = string(hashedPassword)

	suite.mockRepo.On("GetByID", userID).Return(existingUser, nil)

	err := suite.service.UpdatePassword(userID, wrongCurrentPassword, newPassword)

	suite.Error(err)
	suite.Equal(ErrInvalidCredentials, err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test DeleteUser - Success
func (suite *UserServiceTestSuite) TestDeleteUser_Success() {
	userID := uuid.New()
	existingUser := createTestUser()
	existingUser.ID = userID

	suite.mockRepo.On("GetByID", userID).Return(existingUser, nil)
	suite.mockRepo.On("Delete", userID).Return(nil)

	err := suite.service.DeleteUser(userID)

	suite.NoError(err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test DeleteUser - User Not Found
func (suite *UserServiceTestSuite) TestDeleteUser_UserNotFound() {
	userID := uuid.New()

	suite.mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.DeleteUser(userID)

	suite.Error(err)
	suite.Equal(ErrUserNotFound, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AuthenticateUser - Success
func (suite *UserServiceTestSuite) TestAuthenticateUser_Success() {
	email := "test@example.com"
	password := "password123"

	// Hash the password like it would be stored
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := createTestUserWithData(email, "testuser")
	user.PasswordHash = string(hashedPassword)

	suite.mockRepo.On("GetByEmail", email).Return(user, nil)

	result, err := suite.service.AuthenticateUser(email, password)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(user.Email, result.Email)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AuthenticateUser - Invalid Credentials (wrong password)
func (suite *UserServiceTestSuite) TestAuthenticateUser_InvalidPassword() {
	email := "test@example.com"
	password := "wrongpassword"

	// Hash a different password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := createTestUserWithData(email, "testuser")
	user.PasswordHash = string(hashedPassword)

	suite.mockRepo.On("GetByEmail", email).Return(user, nil)

	result, err := suite.service.AuthenticateUser(email, password)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidCredentials, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AuthenticateUser - User Not Found
func (suite *UserServiceTestSuite) TestAuthenticateUser_UserNotFound() {
	email := "nonexistent@example.com"
	password := "password123"

	suite.mockRepo.On("GetByEmail", email).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.AuthenticateUser(email, password)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(ErrInvalidCredentials, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetAllUsers - Success
func (suite *UserServiceTestSuite) TestGetAllUsers_Success() {
	expectedUsers := []*models.User{
		createTestUser(),
		createTestUser(),
	}

	suite.mockRepo.On("GetAll").Return(expectedUsers, nil)

	result, err := suite.service.GetAllUsers()

	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetAllUsers - Repository Error
func (suite *UserServiceTestSuite) TestGetAllUsers_RepositoryError() {
	suite.mockRepo.On("GetAll").Return(nil, assert.AnError)

	result, err := suite.service.GetAllUsers()

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByEmail - Repository Error
func (suite *UserServiceTestSuite) TestGetUserByEmail_RepositoryError() {
	email := "test@example.com"

	suite.mockRepo.On("GetByEmail", email).Return(nil, assert.AnError)

	result, err := suite.service.GetUserByEmail(email)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByID - Repository Error
func (suite *UserServiceTestSuite) TestGetUserByID_RepositoryError() {
	userID := uuid.New()

	suite.mockRepo.On("GetByID", userID).Return(nil, assert.AnError)

	result, err := suite.service.GetUserByID(userID)

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(assert.AnError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdatePassword - User Not Found
func (suite *UserServiceTestSuite) TestUpdatePassword_UserNotFound() {
	userID := uuid.New()
	newPassword := "newpassword123"

	suite.mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.UpdatePassword(userID, "password123", newPassword)

	suite.Error(err)
	suite.Equal(ErrUserNotFound, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdatePassword - Repository Error
func (suite *UserServiceTestSuite) TestUpdatePassword_RepositoryError() {
	userID := uuid.New()
	newPassword := "newpassword123"
	existingUser := createTestUser()
	existingUser.ID = userID

	suite.mockRepo.On("GetByID", userID).Return(existingUser, nil)
	suite.mockRepo.On("Update", mock.AnythingOfType("*models.User")).Return(assert.AnError)

	err := suite.service.UpdatePassword(userID, "password123", newPassword)

	suite.Error(err)
	suite.Equal(assert.AnError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}
