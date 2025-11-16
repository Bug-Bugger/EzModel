package services

import (
	"errors"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type JWTService struct {
	config *config.Config
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

func (s *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	accessToken, err := s.generateToken(user, s.config.JWT.AccessTokenExp)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, s.config.JWT.RefreshTokenExp)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *JWTService) generateToken(user *models.User, expiration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ezmodel-api",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

func (s *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Create a minimal user object to generate new tokens
	user := &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
	}

	return s.GenerateTokenPair(user)
}

func (s *JWTService) GetAccessTokenExpiration() time.Duration {
	return s.config.JWT.AccessTokenExp
}

func (s *JWTService) GetRefreshTokenExpiration() time.Duration {
	return s.config.JWT.RefreshTokenExp
}
