package flagvalues

//go:generate go run go-url-shortener/cmd/reset

import (
	"fmt"
	"strings"

	"go-url-shortener/internal/shared"
)

// generate:reset
type NetAddress struct {
	Scheme         string
	Host           string
	SchemeRequired bool
}

func (addr *NetAddress) String() string {

	strBuilder := strings.Builder{}

	if addr.Scheme != "" {
		strBuilder.WriteString(addr.Scheme)
		strBuilder.WriteString("://")
	}

	strBuilder.WriteString(addr.Host)

	return strBuilder.String()
}

func (addr *NetAddress) Set(input string) error {

	u, err := shared.GetURL(input, addr.SchemeRequired)
	u.Host = strings.Trim(u.Host, ":")

	if err != nil {
		return fmt.Errorf("%w, %s not compatible with [http://]hostname:port", err, input)
	}

	addr.Host = u.Host
	addr.Scheme = u.Scheme

	return nil
}
