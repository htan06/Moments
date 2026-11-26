// Package config loads the migrate command configuration from environment
// variables. Names are listed in services/migrate/.env.example. Values are never logged.
//
// The Postgres address is not read here: pgx reads PGHOST / PGPORT / PGDATABASE itself.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	MigratorUser     string // POSTGRES_MIGRATOR_USER: owns the module schemas
	MigratorPassword string // POSTGRES_MIGRATOR_PWD
	ApplicationRole  string // POSTGRES_APPLICATION_USER: role granted access by grants.sql
}

// Load reads configuration from the environment. It reports every missing
// variable at once.
func Load() (Config, error) {
	var errs []error
	str := func(name string) string {
		v := strings.TrimSpace(os.Getenv(name))
		if v == "" {
			errs = append(errs, fmt.Errorf("%s is required", name))
		}
		return v
	}
	cfg := Config{
		MigratorUser:     str("POSTGRES_MIGRATOR_USER"),
		MigratorPassword: str("POSTGRES_MIGRATOR_PWD"),
		ApplicationRole:  str("POSTGRES_APPLICATION_USER"),
	}
	str("PGHOST") // checked here so a missing address is reported with the rest
	str("PGDATABASE")
	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
