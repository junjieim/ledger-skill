package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("entry not found")

type Store interface {
	Close() error
	CreateEntry(context.Context, Entry) error
	GetEntry(context.Context, string) (Entry, error)
	ListEntries(context.Context, ListFilter) ([]Entry, error)
	SearchEntries(context.Context, SearchQuery) ([]Entry, error)
	UpdateEntry(context.Context, Entry) (bool, error)
	DeleteEntry(context.Context, string) (bool, error)
}

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLiteStore(ctx context.Context, path string) (*SQLiteStore, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil, NewInvalidArgumentError("database path is required")
	}

	if err := os.MkdirAll(filepath.Dir(trimmed), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", trimmed)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	store := &SQLiteStore{db: db}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if err := store.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) CreateEntry(ctx context.Context, entry Entry) error {
	const query = `
INSERT INTO entries (
	id,
	datetime,
	amount,
	currency,
	category,
	note,
	created_at,
	updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := s.db.ExecContext(
		ctx,
		query,
		entry.ID,
		entry.Datetime,
		entry.Amount,
		entry.Currency,
		entry.Category,
		entry.Note,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert entry: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetEntry(ctx context.Context, id string) (Entry, error) {
	const query = `
SELECT id, datetime, amount, currency, category, note, created_at, updated_at
FROM entries
WHERE id = ?;`

	entry, err := scanSingleEntry(s.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Entry{}, ErrNotFound
		}
		return Entry{}, fmt.Errorf("get entry: %w", err)
	}

	return entry, nil
}

func (s *SQLiteStore) ListEntries(ctx context.Context, filter ListFilter) ([]Entry, error) {
	var (
		builder strings.Builder
		args    []any
		where   []string
	)

	builder.WriteString(`
SELECT id, datetime, amount, currency, category, note, created_at, updated_at
FROM entries`)

	if filter.Currency != "" {
		where = append(where, "currency = ?")
		args = append(args, filter.Currency)
	}
	if filter.Category != "" {
		where = append(where, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.From != "" {
		where = append(where, "datetime >= ?")
		args = append(args, filter.From)
	}
	if filter.To != "" {
		where = append(where, "datetime <= ?")
		args = append(args, filter.To)
	}

	if len(where) > 0 {
		builder.WriteString("\nWHERE ")
		builder.WriteString(strings.Join(where, " AND "))
	}

	builder.WriteString("\nORDER BY datetime DESC, created_at DESC")
	if filter.Limit > 0 {
		builder.WriteString("\nLIMIT ?")
		args = append(args, filter.Limit)
	}
	builder.WriteString(";")

	rows, err := s.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}

	return entries, nil
}

func (s *SQLiteStore) SearchEntries(ctx context.Context, query SearchQuery) ([]Entry, error) {
	var (
		builder strings.Builder
		args    = []any{"%" + strings.ToLower(query.Term) + "%"}
	)

	builder.WriteString(`
SELECT id, datetime, amount, currency, category, note, created_at, updated_at
FROM entries
WHERE lower(note) LIKE ?
ORDER BY datetime DESC, created_at DESC`)

	if query.Limit > 0 {
		builder.WriteString("\nLIMIT ?")
		args = append(args, query.Limit)
	}
	builder.WriteString(";")

	rows, err := s.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
	}

	return entries, nil
}

func (s *SQLiteStore) UpdateEntry(ctx context.Context, entry Entry) (bool, error) {
	const query = `
UPDATE entries
SET datetime = ?, amount = ?, currency = ?, category = ?, note = ?, updated_at = ?
WHERE id = ?;`

	result, err := s.db.ExecContext(
		ctx,
		query,
		entry.Datetime,
		entry.Amount,
		entry.Currency,
		entry.Category,
		entry.Note,
		entry.UpdatedAt,
		entry.ID,
	)
	if err != nil {
		return false, fmt.Errorf("update entry: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("inspect updated rows: %w", err)
	}

	return affected > 0, nil
}

func (s *SQLiteStore) DeleteEntry(ctx context.Context, id string) (bool, error) {
	const query = `DELETE FROM entries WHERE id = ?;`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("delete entry: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("inspect deleted rows: %w", err)
	}

	return affected > 0, nil
}

func (s *SQLiteStore) initSchema(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS entries (
	id TEXT PRIMARY KEY,
	datetime TEXT NOT NULL,
	amount TEXT NOT NULL,
	currency TEXT NOT NULL,
	category TEXT NOT NULL,
	note TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_entries_datetime ON entries(datetime);
CREATE INDEX IF NOT EXISTS idx_entries_currency ON entries(currency);
CREATE INDEX IF NOT EXISTS idx_entries_category ON entries(category);`

	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("initialize sqlite schema: %w", err)
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSingleEntry(scanner rowScanner) (Entry, error) {
	var entry Entry
	if err := scanner.Scan(
		&entry.ID,
		&entry.Datetime,
		&entry.Amount,
		&entry.Currency,
		&entry.Category,
		&entry.Note,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return Entry{}, err
	}

	return entry, nil
}

func scanEntries(rows *sql.Rows) ([]Entry, error) {
	entries := make([]Entry, 0)

	for rows.Next() {
		entry, err := scanSingleEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
