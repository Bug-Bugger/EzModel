package service

import (
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(user *models.User) (*services.TokenPair, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenPair), args.Error(1)
}

func (m *MockJWTService) RefreshTokens(refreshToken string) (*services.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenPair), args.Error(1)
}

func (m *MockJWTService) GetAccessTokenExpiration() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockJWTService) GetRefreshTokenExpiration() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*services.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CustomClaims), args.Error(1)
}
