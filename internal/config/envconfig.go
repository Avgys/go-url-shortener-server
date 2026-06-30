package config

import (
	flagvalues "go-url-shortener/internal/config/flag_values"
	"reflect"

	"github.com/caarlos0/env/v11"
)

func parseEnv(cfg *Config) error {
	err := env.ParseWithOptions(cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeFor[flagvalues.NetAddress](): func(v string) (interface{}, error) {
				netAddress := flagvalues.NetAddress{}
				err := netAddress.Set(v)
				return netAddress, err
			},
		}})

	return err
}
