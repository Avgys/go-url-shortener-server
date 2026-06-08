package auth

import (
	"context"
	"errors"
)

// GetFromContext returns validated JWT claims stored by auth middleware.
func GetFromContext(ctx context.Context) (*TokenClaims, error) {
	claims, ok := ctx.Value(claimsName).(*TokenClaims)

	if !ok || !claims.IsOk() {
		return nil, errors.New("wrong auth token")
	}

	return claims, nil
}

// WithContext stores a copy of claims in ctx for downstream handlers and services.
func (c TokenClaims) WithContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(claimsName).(*TokenClaims); ok {
		return ctx
	}
	return context.WithValue(ctx, claimsName, &c)
}
