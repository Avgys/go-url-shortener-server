package config

import "flag"

func parseFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.Var(&cfg.AppURL, "a", "address of HTTP server")
	fs.Var(&cfg.RedirectDomain, "b", "address of redirect")
	fs.StringVar(&cfg.FileStoragePath, "f", "", "file storage name")
	fs.StringVar(&cfg.DBConnectionString, "d", "", "db connection string url")
	fs.StringVar(&cfg.AuditFile, "audit-file", "", "path to audit log file (empty disables file audit)")
	fs.StringVar(&cfg.AuditURL, "audit-url", "", "remote audit receiver URL (empty disables remote audit)")
	fs.BoolVar(&cfg.HttpsEnabled, "s", false, "enable https")

	return fs.Parse(args)
}
