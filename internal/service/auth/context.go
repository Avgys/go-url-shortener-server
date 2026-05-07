package auth

import (
	"context"
	"errors"
)

func GetFromContext(ctx context.Context) (*TokenClaims, error) {
	claims, ok := ctx.Value(claimsName).(*TokenClaims)

	if !ok || claims == nil || claims.UserID == 0 {
		return nil, errors.New("wrong auth token")
	}

	return claims, nil
}

func (c TokenClaims) WithContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(claimsName).(*TokenClaims); ok {
		return ctx
	}
	return context.WithValue(ctx, claimsName, &c)
}
