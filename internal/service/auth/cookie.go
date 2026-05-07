package auth

import (
	"net/http"
	"time"
)

func (c TokenClaims) InjectCookie(w http.ResponseWriter) error {
	tokenString, err := c.ToString()

	if err != nil {
		return err
	}

	newCookie := createCookie(tokenString)
	http.SetCookie(w, newCookie)

	return nil
}

func createCookie(cookieValue string) *http.Cookie {
	newCookie := &http.Cookie{
		Name:     string(CookieName),
		Value:    cookieValue,
		Expires:  time.Now().Add(TokenExp),
		HttpOnly: true, Path: "/"}
	return newCookie
}
