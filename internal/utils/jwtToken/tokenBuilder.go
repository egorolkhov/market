package jwtToken

import (
	"github.com/golang-jwt/jwt/v4"
	"log"
	"os"
	"time"
)

const TokenExp = time.Hour * 24

var secretKey = GetJWTSecret()

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func BuidToken(UUID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp))},
		UserID: UUID,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("JWT_SECRET is not set in environment variables. Used default value")
		secret = "12345"
		return secret
	}
	return secret
}
