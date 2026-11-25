// Command migrate runs once and exits. For each -module, in the given order:
// goose up on the module's migration directory, then the grants.sql in that
// directory for the application role. Connects as the migrator role.
//
//	migrate -module account=path/to/account/migrations -module user=path/to/user/migrations
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/htan06/Moments/migrate/config"
	"github.com/htan06/Moments/migrate/internal/migrate"
)

const timeout = 5 * time.Minute

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("migrate failed", slog.Any("err", err))
		os.Exit(1)
	}
	log.Info("migrate done")
}

func run(log *slog.Logger) error {
	var modules moduleFlags
	flag.Var(&modules, "module", "name=migration_dir; repeat for each module, run in the given order")
	flag.Parse()
	if len(modules) == 0 {
		return errors.New("no -module given")
	}

	// Local dev: load .env if present; real environment variables win.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Host, port and database come from PGHOST / PGPORT / PGDATABASE.
	connCfg, err := pgx.ParseConfig("")
	if err != nil {
		return fmt.Errorf("postgres config: %w", err)
	}
	connCfg.User = cfg.MigratorUser
	connCfg.Password = cfg.MigratorPassword
	db := stdlib.OpenDB(*connCfg)
	defer db.Close()

	for _, m := range modules {
		if err := migrate.Run(ctx, db, m, cfg.ApplicationRole, log); err != nil {
			return err
		}
	}
	return nil
}

// moduleFlags collects repeated -module name=dir flags.
type moduleFlags []migrate.Module

// Module names end up in the goose version table name.
var moduleName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func (f *moduleFlags) String() string { return fmt.Sprint(*f) }

func (f *moduleFlags) Set(v string) error {
	name, dir, ok := strings.Cut(v, "=")
	if !ok || dir == "" {
		return fmt.Errorf("want name=dir, got %q", v)
	}
	if !moduleName.MatchString(name) {
		return fmt.Errorf("module name %q: lowercase letters, digits, _", name)
	}
	for _, m := range *f {
		if m.Name == name {
			return fmt.Errorf("module %q given twice", name)
		}
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return fmt.Errorf("module %q: %q is not a directory", name, dir)
	}
	*f = append(*f, migrate.Module{Name: name, Dir: dir})
	return nil
}
