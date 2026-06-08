package audit

//go:generate mockgen -destination=mocks/mock_publisher.go -package=mocks go-url-shortener/internal/service/audit Publisher

// Publisher emits audit events. *AuditService implements this interface.
type Publisher interface {
	Publish(action string, userID string, url string)
}
