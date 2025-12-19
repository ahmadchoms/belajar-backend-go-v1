package utils

import (
	"errors"
	"os"
	"phase3-api-architecture/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getSecretKey() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		return []byte("jwt_secret_buat_local")
	}
	return []byte(key)
}

func GenerateToken(userID int, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &models.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecretKey())
}

func ParseToken(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getSecretKey(), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
