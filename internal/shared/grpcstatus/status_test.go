package grpcstatus

import (
	"net/http"
	"testing"

	httphelper "go-url-shortener/internal/shared/http"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHTTPStatusToGRPCCode(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		want       codes.Code
	}{
		{name: "bad request", httpStatus: http.StatusBadRequest, want: codes.InvalidArgument},
		{name: "unsupported media type", httpStatus: http.StatusUnsupportedMediaType, want: codes.InvalidArgument},
		{name: "unauthorized", httpStatus: http.StatusUnauthorized, want: codes.Unauthenticated},
		{name: "forbidden", httpStatus: http.StatusForbidden, want: codes.PermissionDenied},
		{name: "not found", httpStatus: http.StatusNotFound, want: codes.NotFound},
		{name: "no content", httpStatus: http.StatusNoContent, want: codes.NotFound},
		{name: "conflict", httpStatus: http.StatusConflict, want: codes.AlreadyExists},
		{name: "gone", httpStatus: http.StatusGone, want: codes.FailedPrecondition},
		{name: "too many requests", httpStatus: http.StatusTooManyRequests, want: codes.ResourceExhausted},
		{name: "not implemented", httpStatus: http.StatusNotImplemented, want: codes.Unimplemented},
		{name: "service unavailable", httpStatus: http.StatusServiceUnavailable, want: codes.Unavailable},
		{name: "gateway timeout", httpStatus: http.StatusGatewayTimeout, want: codes.DeadlineExceeded},
		{name: "internal server error", httpStatus: http.StatusInternalServerError, want: codes.Internal},
		{name: "unknown client error", httpStatus: http.StatusTeapot, want: codes.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, HTTPStatusToGRPCCode(tt.httpStatus))
		})
	}
}

func TestFromError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		require.NoError(t, FromError(nil))
	})

	t.Run("http error", func(t *testing.T) {
		err := FromError(httphelper.NewError("url not found", http.StatusNotFound))
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.NotFound, st.Code())
		require.Equal(t, "url not found", st.Message())
	})

	t.Run("generic error", func(t *testing.T) {
		err := FromError(http.ErrAbortHandler)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Internal, st.Code())
	})
}
