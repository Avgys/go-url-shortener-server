package config

import (
	"fmt"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/shared"
)

type Filepath string

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

func (f *Filepath) String() string {

	return string(*f)
}

func (f *Filepath) Set(input string) error {

	*f = Filepath(input)
	return nil
}
