package config

import (
	"flag"
	"strings"
)

type flagName struct {
	short string
	long  string
}

var configNames []flagName = []flagName{
	{short: "-a", long: "-a="},
	{short: "-b", long: "-b="},
	{short: "-f", long: "-f="},
	{short: "-d", long: "-d="},
	{short: "-audit-file", long: "-audit-file="},
	{short: "-audit-url", long: "-audit-url="},
	{short: "-s", long: "-s="},
	{short: "-grpc", long: "-grpc="},
	{short: "-t", long: "-t="},
}

func parseFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.Var(&cfg.AppURL, "a", "address of HTTP server")
	fs.Var(&cfg.RedirectDomain, "b", "address of redirect")
	fs.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage name")
	fs.StringVar(&cfg.DBConnectionString, "d", cfg.DBConnectionString, "db connection string url")
	fs.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "path to audit log file (empty disables file audit)")
	fs.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "remote audit receiver URL (empty disables remote audit)")
	fs.BoolVar(&cfg.HttpsEnabled, "s", cfg.HttpsEnabled, "enable https")
	fs.IntVar(&cfg.GRPCPort, "grpc", cfg.GRPCPort, "gRPC server port")
	fs.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "trusted subnet to fetch stats")

	args = filterFlags(args, configNames)

	if err := fs.Parse(args); err != nil {
		return err
	}

	return nil
}

func filterFlags(args []string, flagNames []flagName) []string {
	filtered := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]

		for _, flagName := range flagNames {
			if a == flagName.short {

				filtered = append(filtered, a)

				if i+1 < len(args) {
					i++
					value := args[i]

					if !strings.HasPrefix(value, "-") {
						filtered = append(filtered, value)
					}
				}
				continue
			}

			if strings.HasPrefix(a, flagName.long) {
				filtered = append(filtered, a)
			}
		}
	}

	return filtered
}
