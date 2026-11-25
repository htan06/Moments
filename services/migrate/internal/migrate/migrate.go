// Package migrate migrates one module: goose up on its migration directory,
// then the grants.sql in that same directory for the application role.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
)

// Module is one module to migrate.
type Module struct {
	Name string // used in logs and in the goose version table: goose_version_<Name>
	Dir  string // directory with goose migrations and grants.sql
}

// GrantsFile is the grant file inside a module's migration directory. goose
// ignores it because its name has no version number.
const GrantsFile = "grants.sql"

// AppRole is how grants.sql names the application role. It is a valid quoted
// identifier, so sqlc (which reads the migration directory as schema) can parse
// the file.
const AppRole = `"__app_role__"`

// Run applies the module's pending migrations, then its grants.sql in one
// transaction. db must be connected as the migrator role.
func Run(ctx context.Context, db *sql.DB, m Module, appRole string, log *slog.Logger) error {
	if err := up(ctx, db, m, log); err != nil {
		return fmt.Errorf("%s: migrate: %w", m.Name, err)
	}
	if err := grant(ctx, db, m, appRole); err != nil {
		return fmt.Errorf("%s: grants: %w", m.Name, err)
	}
	log.Info("module done", slog.String("module", m.Name))
	return nil
}

func up(ctx context.Context, db *sql.DB, m Module, log *slog.Logger) error {
	p, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(m.Dir),
		goose.WithTableName("goose_version_"+m.Name),
		goose.WithDisableGlobalRegistry(true), // SQL migrations only
	)
	if err != nil {
		return err
	}
	// p.Close is not called: it would close the shared db.
	results, err := p.Up(ctx)
	for _, r := range results {
		log.Info("migration applied", slog.String("module", m.Name), slog.String("file", r.Source.Path))
	}
	return err
}

func grant(ctx context.Context, db *sql.DB, m Module, appRole string) error {
	raw, err := os.ReadFile(filepath.Join(m.Dir, GrantsFile))
	if err != nil {
		return err
	}
	script, err := RenderGrants(string(raw), appRole)
	if err != nil {
		return err
	}
	// One transaction: the application never sees a REVOKE without its GRANT.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit
	// No arguments: pgx uses the simple protocol, which runs several statements.
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return err
	}
	return tx.Commit()
}

// RenderGrants replaces the AppRole placeholder with the quoted role name.
func RenderGrants(sqlText, appRole string) (string, error) {
	if appRole == "" {
		return "", fmt.Errorf("application role is empty")
	}
	if !strings.Contains(sqlText, AppRole) {
		return "", fmt.Errorf("%s does not mention %s", GrantsFile, AppRole)
	}
	return strings.ReplaceAll(sqlText, AppRole, pgx.Identifier{appRole}.Sanitize()), nil
}
