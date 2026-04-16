package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var GetJWTSecret = func() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("super-secret-default-key-please-change-in-prod")
	}
	return []byte(secret)
}

func GenerateTokens(email string) (string, string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	})
	accessString, err := accessToken.SignedString(GetJWTSecret())
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshString, err := refreshToken.SignedString(GetJWTSecret())
	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}

func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return GetJWTSecret(), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, exists := claims["email"].(string); exists {
			return email, nil
		}
	}
	return "", errors.New("invalid token format")
}
