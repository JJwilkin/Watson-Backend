package main

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID int, isSandbox bool, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"is_sandbox": isSandbox,
		"exp":        time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateJWTWithDefaultExpiry(userID int, isSandbox bool) (string, error) {
	return GenerateJWT(userID, isSandbox, time.Hour*24)
}

func GenerateTemporaryJWT(userID int, isSandbox bool) (string, error) {
	return GenerateJWT(userID, isSandbox, time.Minute*15)
}

func VerifyJWT(tokenString string) (int, bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return 0, false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false, errors.New("invalid token claims")
	}

	return int(claims["user_id"].(float64)), claims["is_sandbox"].(bool), nil
}

// returns month and year formatted as MMYYYY
func GetCurrentMonthYear() int {
	// now := time.Now()
	// return int(now.Month())*10000 + now.Year()
	return 72025
}
