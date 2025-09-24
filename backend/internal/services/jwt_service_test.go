package services

import (
	"strings"
	"testing"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type JWTServiceTestSuite struct {
	suite.Suite
	service  *JWTService
	testUser *models.User
	config   *config.Config
}

func (suite *JWTServiceTestSuite) SetupTest() {
	suite.config = &config.Config{}
	suite.config.JWT.Secret = "test-secret-key-for-jwt-testing"
	suite.config.JWT.AccessTokenExp = 15 * time.Minute
	suite.config.JWT.RefreshTokenExp = 7 * 24 * time.Hour

	suite.service = NewJWTService(suite.config)

	suite.testUser = &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
	}
}

func TestJWTServiceSuite(t *testing.T) {
	suite.Run(t, new(JWTServiceTestSuite))
}

// Test GenerateTokenPair - Success
func (suite *JWTServiceTestSuite) TestGenerateTokenPair_Success() {
	tokenPair, err := suite.service.GenerateTokenPair(suite.testUser)

	suite.NoError(err)
	suite.NotNil(tokenPair)
	suite.NotEmpty(tokenPair.AccessToken)
	suite.NotEmpty(tokenPair.RefreshToken)

	// Validate access token structure
	accessParts := strings.Split(tokenPair.AccessToken, ".")
	suite.Len(accessParts, 3, "Access token should have 3 parts (header.payload.signature)")

	// Validate refresh token structure
	refreshParts := strings.Split(tokenPair.RefreshToken, ".")
	suite.Len(refreshParts, 3, "Refresh token should have 3 parts (header.payload.signature)")
}

// Test GenerateTokenPair - Nil User (should panic - this is expected behavior)
func (suite *JWTServiceTestSuite) TestGenerateTokenPair_NilUser() {
	suite.Panics(func() {
		suite.service.GenerateTokenPair(nil)
	})
}

// Test ValidateToken - Valid Access Token
func (suite *JWTServiceTestSuite) TestValidateToken_ValidAccessToken() {
	// First generate a token
	tokenPair, err := suite.service.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Then validate it
	claims, err := suite.service.ValidateToken(tokenPair.AccessToken)

	suite.NoError(err)
	suite.NotNil(claims)
	suite.Equal(suite.testUser.ID, claims.UserID)
	suite.Equal(suite.testUser.Email, claims.Email)
}

// Test ValidateToken - Valid Refresh Token
func (suite *JWTServiceTestSuite) TestValidateToken_ValidRefreshToken() {
	// First generate a token
	tokenPair, err := suite.service.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Then validate it
	claims, err := suite.service.ValidateToken(tokenPair.RefreshToken)

	suite.NoError(err)
	suite.NotNil(claims)
	suite.Equal(suite.testUser.ID, claims.UserID)
	suite.Equal(suite.testUser.Email, claims.Email)
}

// Test ValidateToken - Invalid Token Format
func (suite *JWTServiceTestSuite) TestValidateToken_InvalidFormat() {
	claims, err := suite.service.ValidateToken("invalid-token")

	suite.Error(err)
	suite.Nil(claims)
}

// Test ValidateToken - Empty Token
func (suite *JWTServiceTestSuite) TestValidateToken_EmptyToken() {
	claims, err := suite.service.ValidateToken("")

	suite.Error(err)
	suite.Nil(claims)
}

// Test ValidateToken - Token with Wrong Secret
func (suite *JWTServiceTestSuite) TestValidateToken_WrongSecret() {
	// Create token with different secret
	wrongConfig := &config.Config{}
	wrongConfig.JWT.Secret = "wrong-secret"
	wrongConfig.JWT.AccessTokenExp = suite.config.JWT.AccessTokenExp
	wrongConfig.JWT.RefreshTokenExp = suite.config.JWT.RefreshTokenExp
	wrongService := NewJWTService(wrongConfig)
	tokenPair, err := wrongService.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Try to validate with correct service (different secret)
	claims, err := suite.service.ValidateToken(tokenPair.AccessToken)

	suite.Error(err)
	suite.Nil(claims)
}

// Test ValidateToken - Expired Token
func (suite *JWTServiceTestSuite) TestValidateToken_ExpiredToken() {
	// Create a service with very short expiration
	shortConfig := &config.Config{}
	shortConfig.JWT.Secret = suite.config.JWT.Secret
	shortConfig.JWT.AccessTokenExp = 1 * time.Nanosecond
	shortConfig.JWT.RefreshTokenExp = suite.config.JWT.RefreshTokenExp
	shortService := NewJWTService(shortConfig)
	tokenPair, err := shortService.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Wait a bit to ensure expiration
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	claims, err := suite.service.ValidateToken(tokenPair.AccessToken)

	suite.Error(err)
	suite.Nil(claims)
}

// Test RefreshTokens - Success
func (suite *JWTServiceTestSuite) TestRefreshTokens_Success() {
	// First generate a token pair
	originalPair, err := suite.service.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Refresh the tokens
	newPair, err := suite.service.RefreshTokens(originalPair.RefreshToken)

	suite.NoError(err)
	suite.NotNil(newPair)
	suite.NotEmpty(newPair.AccessToken)
	suite.NotEmpty(newPair.RefreshToken)

	// Validate new access token works correctly
	accessClaims, err := suite.service.ValidateToken(newPair.AccessToken)
	suite.NoError(err)
	suite.Equal(suite.testUser.ID, accessClaims.UserID)
	suite.Equal(suite.testUser.Email, accessClaims.Email)

	// Validate new refresh token works correctly
	refreshClaims, err := suite.service.ValidateToken(newPair.RefreshToken)
	suite.NoError(err)
	suite.Equal(suite.testUser.ID, refreshClaims.UserID)
	suite.Equal(suite.testUser.Email, refreshClaims.Email)
}

// Test RefreshTokens - Invalid Refresh Token
func (suite *JWTServiceTestSuite) TestRefreshTokens_InvalidToken() {
	newPair, err := suite.service.RefreshTokens("invalid-token")

	suite.Error(err)
	suite.Nil(newPair)
}

// Test RefreshTokens - Expired Refresh Token
func (suite *JWTServiceTestSuite) TestRefreshTokens_ExpiredRefreshToken() {
	// Create service with very short refresh token expiration
	shortConfig := &config.Config{}
	shortConfig.JWT.Secret = suite.config.JWT.Secret
	shortConfig.JWT.AccessTokenExp = suite.config.JWT.AccessTokenExp
	shortConfig.JWT.RefreshTokenExp = 1 * time.Nanosecond
	shortService := NewJWTService(shortConfig)
	tokenPair, err := shortService.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Wait to ensure expiration
	time.Sleep(10 * time.Millisecond)

	// Try to refresh with expired token
	newPair, err := suite.service.RefreshTokens(tokenPair.RefreshToken)

	suite.Error(err)
	suite.Nil(newPair)
}

// Test GetAccessTokenExpiration
func (suite *JWTServiceTestSuite) TestGetAccessTokenExpiration() {
	expiration := suite.service.GetAccessTokenExpiration()
	suite.Equal(suite.config.JWT.AccessTokenExp, expiration)
}

// Test Token Claims Structure
func (suite *JWTServiceTestSuite) TestTokenClaims_Structure() {
	tokenPair, err := suite.service.GenerateTokenPair(suite.testUser)
	suite.NoError(err)

	// Parse the access token manually to check claims structure
	token, err := jwt.ParseWithClaims(tokenPair.AccessToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(suite.config.JWT.Secret), nil
	})

	suite.NoError(err)
	suite.True(token.Valid)

	claims, ok := token.Claims.(*CustomClaims)
	suite.True(ok)
	suite.Equal(suite.testUser.ID, claims.UserID)
	suite.Equal(suite.testUser.Email, claims.Email)
	suite.NotZero(claims.ExpiresAt.Unix())
	suite.NotZero(claims.IssuedAt.Unix())
}

// Test Token Generation Consistency
func (suite *JWTServiceTestSuite) TestTokenGeneration_ConsistentClaims() {
	// Generate multiple token pairs and ensure consistency
	for i := 0; i < 3; i++ {
		tokenPair, err := suite.service.GenerateTokenPair(suite.testUser)
		suite.NoError(err)

		// Validate access token claims
		accessClaims, err := suite.service.ValidateToken(tokenPair.AccessToken)
		suite.NoError(err)
		suite.Equal(suite.testUser.ID, accessClaims.UserID)
		suite.Equal(suite.testUser.Email, accessClaims.Email)

		// Validate refresh token claims
		refreshClaims, err := suite.service.ValidateToken(tokenPair.RefreshToken)
		suite.NoError(err)
		suite.Equal(suite.testUser.ID, refreshClaims.UserID)
		suite.Equal(suite.testUser.Email, refreshClaims.Email)
	}
}

// Test Edge Case - User with Empty Email
func (suite *JWTServiceTestSuite) TestGenerateTokenPair_EmptyEmail() {
	userWithEmptyEmail := &models.User{
		ID:       uuid.New(),
		Email:    "", // Empty email
		Username: "testuser",
	}

	tokenPair, err := suite.service.GenerateTokenPair(userWithEmptyEmail)

	suite.NoError(err)
	suite.NotNil(tokenPair)

	// Validate the token can still be parsed
	claims, err := suite.service.ValidateToken(tokenPair.AccessToken)
	suite.NoError(err)
	suite.Equal(userWithEmptyEmail.ID, claims.UserID)
	suite.Equal("", claims.Email)
}
