package handler

import (
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service/audit"
	"go-url-shortener/internal/service/shortifier"
)

// Handlers holds dependencies shared by all HTTP handlers in this package.
type Handlers struct {
	Shortifier   *shortifier.Shortifier
	Store        repository.Repository
	AuditService audit.Publisher
}

// NewHandlers returns a Handlers value wired with the given shortifier, store, and audit publisher.
func NewHandlers(shortifier *shortifier.Shortifier, store repository.Repository, auditService audit.Publisher) *Handlers {
	return &Handlers{Shortifier: shortifier, Store: store, AuditService: auditService}
}
