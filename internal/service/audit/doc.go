// Package audit records user actions on shortened URLs.
//
// Events are published through [Publisher] (implemented by [AuditService]) and
// forwarded to registered [Observer] implementations. Observers run asynchronously;
// delivery errors are logged and do not block callers.
//
// # Configuration
//
// [AuditService.Init] registers observers from [config.Config]:
//
//   - AUDIT_FILE / -audit-file — append events as CSV via [AuditFile]
//   - AUDIT_URL / -audit-url — POST JSON events to a remote service via [AuditClient]
//
// # Event shape
//
// Each [AuditEvent] carries a Unix timestamp, action name (for example "shorten"
// or "follow"), user ID, and affected URL. The shortifier publishes on successful
// shortens; redirect handlers may publish when an authenticated user follows a link.
//
// Use [Publisher] as the dependency in handlers and services; mock it in tests with
// the generated mock in the mocks subpackage.
package audit
