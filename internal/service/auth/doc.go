// Package auth implements JWT authentication for the URL shortener.
//
// Tokens are HS256-signed [TokenClaims] stored in the [CookieName] HTTP cookie.
// Middleware in internal/middlewares parses the cookie, validates the JWT, and
// attaches claims to the request context via [TokenClaims.WithContext].
// Handlers and services read the user ID with [GetFromContext].
//
// # Token lifecycle
//
//   - [NewToken] builds claims for a user ID and login
//   - [TokenClaims.ToString] serializes a signed JWT string
//   - [ParseToken] validates and decodes a JWT from the cookie value
//   - [TokenClaims.InjectCookie] sets the auth cookie on a response
//
// Anonymous requests omit valid claims; protected endpoints rely on middleware
// to reject them before handler code runs.
package auth
