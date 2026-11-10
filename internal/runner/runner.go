package runner

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// RunOptions configures the execution environment.
type RunOptions struct {
	DatabaseURL string
	Query       string
	SafeMode    bool
	TimeoutSec  int
}

// ExecuteExplain connects to the specified database, runs EXPLAIN, and returns raw plan output.
func ExecuteExplain(opts RunOptions) (string, error) {
	dbURL := strings.TrimSpace(opts.DatabaseURL)
	query := strings.TrimSpace(opts.Query)

	if dbURL == "" {
		return "", fmt.Errorf("database URL is required")
	}
	if query == "" {
		return "", fmt.Errorf("query string is required")
	}

	driverName := "postgres"
	isMySQL := strings.HasPrefix(dbURL, "mysql://") || strings.Contains(dbURL, "@tcp(")
	if isMySQL {
		driverName = "mysql"
		dbURL = strings.TrimPrefix(dbURL, "mysql://")
	}

	db, err := sql.Open(driverName, dbURL)
	if err != nil {
		return "", fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	timeout := time.Duration(opts.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return "", fmt.Errorf("database connection ping failed: %w", err)
	}

	if isMySQL {
		return runMySQL(ctx, db, query)
	}
	return runPostgres(ctx, db, query, opts.SafeMode)
}

func runPostgres(ctx context.Context, db *sql.DB, query string, safeMode bool) (string, error) {
	explainSQL := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, SETTINGS, WAL, FORMAT JSON) %s", query)

	if safeMode {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin safe transaction: %w", err)
		}
		defer tx.Rollback()

		rows, err := tx.QueryContext(ctx, explainSQL)
		if err != nil {
			return "", fmt.Errorf("failed to execute explain in transaction: %w", err)
		}
		defer rows.Close()

		return scanSingleColumn(rows)
	}

	rows, err := db.QueryContext(ctx, explainSQL)
	if err != nil {
		return "", fmt.Errorf("failed to execute explain: %w", err)
	}
	defer rows.Close()

	return scanSingleColumn(rows)
}

func runMySQL(ctx context.Context, db *sql.DB, query string) (string, error) {
	explainSQL := fmt.Sprintf("EXPLAIN FORMAT=JSON %s", query)
	rows, err := db.QueryContext(ctx, explainSQL)
	if err != nil {
		return "", fmt.Errorf("failed to execute MySQL explain: %w", err)
	}
	defer rows.Close()

	return scanSingleColumn(rows)
}

func scanSingleColumn(rows *sql.Rows) (string, error) {
	var builder strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(builder.String()), nil
}
