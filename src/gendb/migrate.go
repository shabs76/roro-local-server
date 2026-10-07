package gendb

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations applies the SQL files in migrations/ in name order. Every statement
// in those files must be idempotent, because the files run on every start. It waits
// up to waitFor for the database to accept connections first.
func RunMigrations(waitFor time.Duration) error {
	db, err := InitDb()
	if err != nil {
		return err
	}

	deadline := time.Now().Add(waitFor)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("database not reachable for migrations: %w", err)
		}
		slog.Warn("Waiting for database before running migrations", "error", err)
		time.Sleep(2 * time.Second)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	var errs []error
	for _, name := range names {
		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, stmt := range splitSQLStatements(string(content)) {
			if _, err := db.Exec(stmt); err != nil {
				var myErr *mysql.MySQLError
				if errors.As(err, &myErr) && (myErr.Number == 1142 || myErr.Number == 1044) {
					// The application user has no ALTER/CREATE rights. Every other
					// statement would fail the same way, so stop here with one message.
					resetColumnCache()
					return fmt.Errorf("database user may not change the schema; apply gendb/migrations/%s as an administrator: %w", name, err)
				}
				slog.Error("Migration statement failed", "file", name, "statement", stmt, "error", err)
				errs = append(errs, fmt.Errorf("%s: %w", name, err))
			}
		}
		slog.Info("Migration file applied", "file", name)
	}

	resetColumnCache()
	return errors.Join(errs...)
}

// splitSQLStatements drops "--" comment lines and splits the rest on semicolons that
// end a line.
func splitSQLStatements(content string) []string {
	var b strings.Builder
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	stmts := []string{}
	for _, part := range strings.Split(b.String(), ";\n") {
		stmt := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(part), ";"))
		if stmt != "" {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}
