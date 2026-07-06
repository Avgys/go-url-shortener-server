package interceptors

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const headerAuthorization = "authorization"

func UnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	traceLogger, close, err := logger.Middleware(ctx, "AuthUnaryInterceptor")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create request logger")
	}
	defer func() { _ = close() }()

	tokenString, claims, isNewToken, err := resolveAuthToken(ctx, traceLogger)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "auth token: %v", err)
	}

	traceLogger.Info().Bool("IsNewToken", isNewToken).Send()

	authCtx := claims.WithContext(ctx)

	if err := setAuthTokenMetadata(authCtx, tokenString); err != nil {
		traceLogger.Error().Err(err).Msg("failed to set auth token metadata")
		return nil, status.Errorf(codes.Internal, "set auth token metadata: %v", err)
	}

	return handler(authCtx, req)
}

func resolveAuthToken(ctx context.Context, traceLogger *zerolog.Logger) (string, *auth.TokenClaims, bool, error) {
	tokenString := tokenFromIncomingMetadata(ctx)

	resetToken := tokenString == ""
	var claims *auth.TokenClaims
	var err error

	if !resetToken {
		claims, err = auth.ParseToken(tokenString)
		if err != nil || claims == nil || claims.UserID == 0 {
			resetToken = true
		}
	}

	if resetToken {
		userID, err := rand.Int(rand.Reader, big.NewInt(1<<62))
		if err != nil {
			traceLogger.Error().Err(err).Msg("failed to generate random userID for auth token")
			return "", nil, false, err
		}

		claims = auth.NewToken(userID.Int64(), "")
		tokenString, err = claims.ToString()
		if err != nil || tokenString == "" {
			if err != nil {
				traceLogger.Error().Err(err).Msg("failed to create JWT token")
			} else {
				traceLogger.Error().Msg("generated invalid JWT token")
			}
			return "", nil, false, err
		}

		return tokenString, claims, true, nil
	}

	return tokenString, claims, false, nil
}

func tokenFromIncomingMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	tokens := md.Get(headerAuthorization)
	if len(tokens) == 0 {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(tokens[0], "Bearer "))
}

func setAuthTokenMetadata(ctx context.Context, tokenString string) error {
	return grpc.SetHeader(ctx, metadata.Pairs(headerAuthorization, tokenString))
}
