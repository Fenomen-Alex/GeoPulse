package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/alex/geopulse/server/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT verifies a JWT token and returns the claims
func ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the token with the JWT secret
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	return userID, ok
}
