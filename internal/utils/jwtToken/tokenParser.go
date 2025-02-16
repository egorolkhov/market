package jwtToken

import (
	"github.com/golang-jwt/jwt/v4"
)

func GetUserID(JwtToken string) string {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(JwtToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return ""
	}
	if !token.Valid {
		return ""
	}

	return claims.UserID
}
