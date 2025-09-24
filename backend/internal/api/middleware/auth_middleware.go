package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/services"
)

const (
	authorizationHeader = "Authorization"
	userIDKey           = "userID"
)

type AuthMiddleware struct {
	jwtService services.JWTServiceInterface
}

func NewAuthMiddleware(jwtService services.JWTServiceInterface) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get(authorizationHeader)
		if authHeader == "" {
			responses.RespondWithError(w, http.StatusUnauthorized, "No authorization header provided")
			return
		}

		// Check if the format is "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			responses.RespondWithError(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		token := parts[1]
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			if err == services.ErrExpiredToken {
				responses.RespondWithError(w, http.StatusUnauthorized, "Token has expired")
			} else {
				responses.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			}
			return
		}

		// Set the userID in the request context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
