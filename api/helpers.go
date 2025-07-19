package main

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID int, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateJWTWithDefaultExpiry(userID int) (string, error) {
	return GenerateJWT(userID, time.Hour*24)
}

func GenerateTemporaryJWT(userID int) (string, error) {
	return GenerateJWT(userID, time.Minute*15)
}

func VerifyJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return int(claims["user_id"].(float64)), nil
}

// returns month and year formatted as MMYYYY
func GetCurrentMonthYear() int {
	now := time.Now()
	return int(now.Month())*10000 + now.Year()
}
