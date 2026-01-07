package handler

type Handlers struct {
	Shortifier Shortifier
}

type Shortifier interface {
	ResolveShortURL(shortURL string) (string, error)
	ShortifyURL(url string) (string, bool, error)
}

func NewHandlers(shortifier Shortifier) *Handlers {
	return &Handlers{Shortifier: shortifier}
}
