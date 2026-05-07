package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CookieNameType string

const CookieName string = "AUTH_COOKIE"
const claimsName CookieNameType = "CLAIMS"
const secretKey = "SECRETTOKEN"

const TokenExp = time.Hour * 3

var signMethod = jwt.SigningMethodHS256

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID int64  `json:"user_id,omitempty"`
	Login  string `json:"login,omitempty"`
}

func NewToken(userID int64, login string) *TokenClaims {

	claims := TokenClaims{
		UserID:           userID,
		Login:            login,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp))},
	}

	return &claims
}

func (t TokenClaims) ToString() (string, error) {

	token := jwt.NewWithClaims(signMethod, t)

	tokenStr, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", fmt.Errorf("error serialize token, %w", err)
	}

	return tokenStr, nil
}

func ParseToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, verifyToken)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("auth token is invalid")
	}

	return claims, nil
}

func verifyToken(t *jwt.Token) (interface{}, error) {
	if t.Method.Alg() != signMethod.Alg() {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}

	return []byte(secretKey), nil
}
