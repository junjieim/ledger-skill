package ledger

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestSQLiteStoreCRUDAndQueries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)

	entries := []Entry{
		{
			ID:        "entry-1",
			Datetime:  "2026-04-01T00:00:00Z",
			Amount:    "10.00",
			Currency:  "HKD",
			Category:  "food",
			Note:      "breakfast with team",
			CreatedAt: "2026-04-01T00:00:00Z",
			UpdatedAt: "2026-04-01T00:00:00Z",
		},
		{
			ID:        "entry-2",
			Datetime:  "2026-04-03T00:00:00Z",
			Amount:    "80.00",
			Currency:  "USD",
			Category:  "travel",
			Note:      "Taxi to airport",
			CreatedAt: "2026-04-03T00:00:00Z",
			UpdatedAt: "2026-04-03T00:00:00Z",
		},
		{
			ID:        "entry-3",
			Datetime:  "2026-04-02T00:00:00Z",
			Amount:    "15.00",
			Currency:  "HKD",
			Category:  "food",
			Note:      "Lunch at cafe",
			CreatedAt: "2026-04-02T00:00:00Z",
			UpdatedAt: "2026-04-02T00:00:00Z",
		},
	}

	for _, entry := range entries {
		if err := store.CreateEntry(ctx, entry); err != nil {
			t.Fatalf("CreateEntry() error = %v", err)
		}
	}

	got, err := store.GetEntry(ctx, "entry-1")
	if err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
	if got.Note != "breakfast with team" {
		t.Fatalf("GetEntry() note = %q", got.Note)
	}

	listed, err := store.ListEntries(ctx, ListFilter{
		Currency: "HKD",
		Category: "food",
		From:     "2026-04-01T00:00:00Z",
		To:       "2026-04-02T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("ListEntries() len = %d, want 2", len(listed))
	}
	if listed[0].ID != "entry-3" || listed[1].ID != "entry-1" {
		t.Fatalf("ListEntries() ids = [%s %s], want [entry-3 entry-1]", listed[0].ID, listed[1].ID)
	}

	searched, err := store.SearchEntries(ctx, SearchQuery{
		Term: "taxi",
	})
	if err != nil {
		t.Fatalf("SearchEntries() error = %v", err)
	}
	if len(searched) != 1 || searched[0].ID != "entry-2" {
		t.Fatalf("SearchEntries() = %+v, want entry-2", searched)
	}

	updated := entries[0]
	updated.Note = "breakfast with client"
	updated.UpdatedAt = "2026-04-04T00:00:00Z"

	ok, err := store.UpdateEntry(ctx, updated)
	if err != nil {
		t.Fatalf("UpdateEntry() error = %v", err)
	}
	if !ok {
		t.Fatal("UpdateEntry() ok = false, want true")
	}

	got, err = store.GetEntry(ctx, "entry-1")
	if err != nil {
		t.Fatalf("GetEntry() after update error = %v", err)
	}
	if got.Note != "breakfast with client" {
		t.Fatalf("GetEntry() after update note = %q", got.Note)
	}

	deleted, err := store.DeleteEntry(ctx, "entry-2")
	if err != nil {
		t.Fatalf("DeleteEntry() error = %v", err)
	}
	if !deleted {
		t.Fatal("DeleteEntry() deleted = false, want true")
	}

	_, err = store.GetEntry(ctx, "entry-2")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetEntry() missing error = %v, want ErrNotFound", err)
	}
}

func TestOpenSQLiteStoreRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	_, err := OpenSQLiteStore(context.Background(), " ")
	if err == nil {
		t.Fatal("OpenSQLiteStore() error = nil, want invalid argument")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Code != CodeInvalidArgument {
		t.Fatalf("OpenSQLiteStore() error = %+v", err)
	}
}

func TestSQLiteStoreMissingRowsAndLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)

	entries := []Entry{
		{
			ID:        "entry-1",
			Datetime:  "2026-04-01T00:00:00Z",
			Amount:    "10.00",
			Currency:  "HKD",
			Category:  "food",
			Note:      "breakfast",
			CreatedAt: "2026-04-01T00:00:00Z",
			UpdatedAt: "2026-04-01T00:00:00Z",
		},
		{
			ID:        "entry-2",
			Datetime:  "2026-04-02T00:00:00Z",
			Amount:    "20.00",
			Currency:  "HKD",
			Category:  "food",
			Note:      "lunch",
			CreatedAt: "2026-04-02T00:00:00Z",
			UpdatedAt: "2026-04-02T00:00:00Z",
		},
	}

	for _, entry := range entries {
		if err := store.CreateEntry(ctx, entry); err != nil {
			t.Fatalf("CreateEntry() error = %v", err)
		}
	}

	listed, err := store.ListEntries(ctx, ListFilter{Limit: 1})
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(listed) != 1 || listed[0].ID != "entry-2" {
		t.Fatalf("ListEntries() = %+v", listed)
	}

	searched, err := store.SearchEntries(ctx, SearchQuery{Term: "dinner", Limit: 1})
	if err != nil {
		t.Fatalf("SearchEntries() error = %v", err)
	}
	if len(searched) != 0 {
		t.Fatalf("SearchEntries() = %+v, want empty", searched)
	}

	updated, err := store.UpdateEntry(ctx, Entry{ID: "missing"})
	if err != nil {
		t.Fatalf("UpdateEntry() error = %v", err)
	}
	if updated {
		t.Fatal("UpdateEntry() updated = true, want false")
	}

	deleted, err := store.DeleteEntry(ctx, "missing")
	if err != nil {
		t.Fatalf("DeleteEntry() error = %v", err)
	}
	if deleted {
		t.Fatal("DeleteEntry() deleted = true, want false")
	}
}

func TestSQLiteStoreRejectsDuplicateID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)

	entry := Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "10.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "breakfast",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}

	if err := store.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry() error = %v", err)
	}

	if err := store.CreateEntry(ctx, entry); err == nil {
		t.Fatal("CreateEntry() error = nil, want duplicate key error")
	}
}

func openTestStore(t *testing.T) *SQLiteStore {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ledger.db")
	store, err := OpenSQLiteStore(context.Background(), path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}

	t.Cleanup(func() {
		_ = store.Close()
	})

	return store
}
