package shortifier

//go:generate mockgen -destination=mocks/mock_string_generator.go -package=mocks go-url-shortener/internal/service/shortifier StringGenerator
