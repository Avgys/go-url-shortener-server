// Package shared provides small utilities shared across handlers, services, and config.
//
// [GetURL] parses and validates user-supplied URLs before they are stored or used
// in redirects. HTTP request/response helpers live in the nested package
// internal/shared/http (import path; package name httphelper).
package shared
