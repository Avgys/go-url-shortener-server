// Package httphelper provides HTTP helpers for handlers and middleware.
//
// Import path: go-url-shortener/internal/shared/http.
//
// Typical usage:
//
//   - [GetJSONBody] — decode JSON request bodies; syntax errors become 400 responses
//   - [GetRequestBody] — read plain-text bodies with a size limit
//   - [WriteResponse] / [WriteResponseStr] — write responses with trace logging
//   - [HandleErr] — map errors (including [ShowHTTPError]) to HTTP status codes
//   - [NewError] — attach an HTTP status to an error for upstream layers
//
// Services such as shortifier return [ShowHTTPError] values so handlers can
// respond with the intended status without duplicating mapping logic.
package httphelper
