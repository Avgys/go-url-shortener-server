package jwttoken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenExp = time.Hour * 3

var signMethod = jwt.SigningMethodHS256
var secretKey = "SECRETTOKEN" //service.NewStringGenerator().GetRandomString(15)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id,omitempty"`
}

func NewTokenWithUserID(userID int64) (string, *Claims, error) {

	claims := Claims{
		UserID:           userID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp))},
	}

	token := jwt.NewWithClaims(signMethod, claims)

	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", nil, err
	}

	return tokenString, &claims, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, verifyToken)

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func verifyToken(t *jwt.Token) (interface{}, error) {
	if t.Method.Alg() != signMethod.Alg() {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}

	return []byte(secretKey), nil
}
