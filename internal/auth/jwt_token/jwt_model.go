package jwt_token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TOKEN_EXP = time.Hour * 3

var sign_method = jwt.SigningMethodHS256
var SECRET_KEY = "SECRETTOKEN" //service.NewStringGenerator().GetRandomString(15)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id,omitempty"`
}

func NewTokenWithUserId(userID int64) (string, *Claims, error) {

	claims := Claims{
		UserID:           userID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP))},
	}

	token := jwt.NewWithClaims(sign_method, claims)

	tokenString, err := token.SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", nil, err
	}

	return tokenString, &claims, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, verifyToken)

	if err != nil && !token.Valid {
		return nil, err
	}

	return claims, nil
}

func verifyToken(t *jwt.Token) (interface{}, error) {
	if t.Method.Alg() != sign_method.Alg() {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}

	return []byte(SECRET_KEY), nil
}

// func NewJwtToken(claims Claims) (string, error) {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// }
