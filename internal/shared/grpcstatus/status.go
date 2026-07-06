package grpcstatus

import (
	"errors"
	"net/http"

	httphelper "go-url-shortener/internal/shared/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPStatusToGRPCCode maps an HTTP status code to the closest gRPC status code.
func HTTPStatusToGRPCCode(code int) codes.Code {
	switch code {
	case http.StatusBadRequest, http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound, http.StatusNoContent:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusGone:
		return codes.FailedPrecondition
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	default:
		if code >= http.StatusInternalServerError {
			return codes.Internal
		}
		return codes.Unknown
	}
}

// FromError converts err into a gRPC status, mapping [httphelper.ShowHTTPError] values.
func FromError(err error) error {
	if err == nil {
		return nil
	}

	var httpErr *httphelper.ShowHTTPError
	if errors.As(err, &httpErr) {
		return status.Error(HTTPStatusToGRPCCode(httpErr.StatusCode), httpErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
