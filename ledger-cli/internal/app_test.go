package ledger

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubStore struct {
	closeCalled      bool
	createEntryInput Entry
	getEntryInput    string
	listFilterInput  ListFilter
	searchQueryInput SearchQuery
	updateEntryInput Entry
	deleteIDInput    string

	createEntryFunc func(context.Context, Entry) error
	getEntryFunc    func(context.Context, string) (Entry, error)
	listEntriesFunc func(context.Context, ListFilter) ([]Entry, error)
	searchFunc      func(context.Context, SearchQuery) ([]Entry, error)
	updateEntryFunc func(context.Context, Entry) (bool, error)
	deleteEntryFunc func(context.Context, string) (bool, error)
	closeFunc       func() error
}

func (s *stubStore) Close() error {
	s.closeCalled = true
	if s.closeFunc != nil {
		return s.closeFunc()
	}
	return nil
}

func (s *stubStore) CreateEntry(ctx context.Context, entry Entry) error {
	s.createEntryInput = entry
	if s.createEntryFunc != nil {
		return s.createEntryFunc(ctx, entry)
	}
	return nil
}

func (s *stubStore) GetEntry(ctx context.Context, id string) (Entry, error) {
	s.getEntryInput = id
	if s.getEntryFunc != nil {
		return s.getEntryFunc(ctx, id)
	}
	return Entry{}, nil
}

func (s *stubStore) ListEntries(ctx context.Context, filter ListFilter) ([]Entry, error) {
	s.listFilterInput = filter
	if s.listEntriesFunc != nil {
		return s.listEntriesFunc(ctx, filter)
	}
	return nil, nil
}

func (s *stubStore) SearchEntries(ctx context.Context, query SearchQuery) ([]Entry, error) {
	s.searchQueryInput = query
	if s.searchFunc != nil {
		return s.searchFunc(ctx, query)
	}
	return nil, nil
}

func (s *stubStore) UpdateEntry(ctx context.Context, entry Entry) (bool, error) {
	s.updateEntryInput = entry
	if s.updateEntryFunc != nil {
		return s.updateEntryFunc(ctx, entry)
	}
	return true, nil
}

func (s *stubStore) DeleteEntry(ctx context.Context, id string) (bool, error) {
	s.deleteIDInput = id
	if s.deleteEntryFunc != nil {
		return s.deleteEntryFunc(ctx, id)
	}
	return true, nil
}

func TestNewAppAndClose(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	app := NewApp(store)

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
	if app.store != store {
		t.Fatal("NewApp() did not keep the provided store")
	}
	if app.now == nil || app.newID == nil {
		t.Fatal("NewApp() did not initialize dependencies")
	}

	if err := app.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !store.closeCalled {
		t.Fatal("Close() did not call store.Close()")
	}

	var nilApp *App
	if err := nilApp.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}
}

func TestAppAdd(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	app := NewApp(store)
	app.now = func() time.Time { return time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC) }
	app.newID = func() (string, error) { return "fixed-id", nil }

	entry, err := app.Add(context.Background(), CreateInput{
		Datetime: "2026-04-01T08:00:00+08:00",
		Amount:   "10.50",
		Currency: "HKD",
		Category: "food",
		Note:     " ",
	})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if entry.ID != "fixed-id" {
		t.Fatalf("Add() id = %q", entry.ID)
	}
	if entry.CreatedAt != "2026-04-01T12:00:00Z" || entry.UpdatedAt != "2026-04-01T12:00:00Z" {
		t.Fatalf("Add() timestamps = %q / %q", entry.CreatedAt, entry.UpdatedAt)
	}
	if entry.Note != "none" {
		t.Fatalf("Add() note = %q", entry.Note)
	}
	if store.createEntryInput != entry {
		t.Fatalf("CreateEntry() input = %+v, want %+v", store.createEntryInput, entry)
	}
}

func TestAppAddGenerateIDError(t *testing.T) {
	t.Parallel()

	app := NewApp(&stubStore{})
	app.newID = func() (string, error) { return "", errors.New("boom") }

	_, err := app.Add(context.Background(), CreateInput{
		Datetime: "2026-04-01T00:00:00Z",
		Amount:   "10.00",
		Currency: "HKD",
		Category: "food",
		Note:     "breakfast",
	})
	if err == nil || err.Error() != "generate entry id: boom" {
		t.Fatalf("Add() error = %v", err)
	}
}

func TestAppGetMapsNotFound(t *testing.T) {
	t.Parallel()

	store := &stubStore{
		getEntryFunc: func(context.Context, string) (Entry, error) {
			return Entry{}, ErrNotFound
		},
	}
	app := NewApp(store)

	_, err := app.Get(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("Get() error = nil, want not found")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Code != CodeNotFound {
		t.Fatalf("Get() error = %+v, want not_found", err)
	}
}

func TestAppListNormalizesFilter(t *testing.T) {
	t.Parallel()

	wantEntries := []Entry{{ID: "entry-1"}}
	store := &stubStore{
		listEntriesFunc: func(_ context.Context, filter ListFilter) ([]Entry, error) {
			if filter.Currency != "HKD" || filter.Category != "food" {
				t.Fatalf("List() filter text = %+v", filter)
			}
			if filter.From != "2026-04-01T00:00:00Z" || filter.To != "2026-04-02T00:00:00Z" {
				t.Fatalf("List() filter time = %+v", filter)
			}
			return wantEntries, nil
		},
	}
	app := NewApp(store)

	got, err := app.List(context.Background(), ListFilter{
		Currency: " HKD ",
		Category: " food ",
		From:     "2026-04-01T08:00:00+08:00",
		To:       "2026-04-02T08:00:00+08:00",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "entry-1" {
		t.Fatalf("List() = %+v", got)
	}
}

func TestAppSearchNormalizesQuery(t *testing.T) {
	t.Parallel()

	store := &stubStore{
		searchFunc: func(_ context.Context, query SearchQuery) ([]Entry, error) {
			if query.Term != "lunch" || query.Limit != 2 {
				t.Fatalf("Search() query = %+v", query)
			}
			return []Entry{{ID: "entry-1"}}, nil
		},
	}
	app := NewApp(store)

	got, err := app.Search(context.Background(), SearchQuery{Term: "  lunch ", Limit: 2})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "entry-1" {
		t.Fatalf("Search() = %+v", got)
	}
}

func TestAppUpdate(t *testing.T) {
	t.Parallel()

	baseEntry := Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "10.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "breakfast",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}

	t.Run("updates merged entry and timestamp", func(t *testing.T) {
		t.Parallel()

		store := &stubStore{
			getEntryFunc: func(context.Context, string) (Entry, error) {
				return baseEntry, nil
			},
		}
		app := NewApp(store)
		app.now = func() time.Time { return time.Date(2026, 4, 2, 12, 30, 0, 0, time.UTC) }

		note := " "
		amount := "15.25"
		got, err := app.Update(context.Background(), "entry-1", UpdateInput{
			Amount: &amount,
			Note:   &note,
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if got.Amount != "15.25" || got.Note != "none" {
			t.Fatalf("Update() entry = %+v", got)
		}
		if got.UpdatedAt != "2026-04-02T12:30:00Z" {
			t.Fatalf("Update() updated_at = %q", got.UpdatedAt)
		}
		if store.updateEntryInput != got {
			t.Fatalf("UpdateEntry() input = %+v, want %+v", store.updateEntryInput, got)
		}
	})

	t.Run("maps missing entry from get", func(t *testing.T) {
		t.Parallel()

		store := &stubStore{
			getEntryFunc: func(context.Context, string) (Entry, error) {
				return Entry{}, ErrNotFound
			},
		}
		app := NewApp(store)
		value := "15.25"

		_, err := app.Update(context.Background(), "entry-1", UpdateInput{Amount: &value})
		if err == nil {
			t.Fatal("Update() error = nil, want not found")
		}

		var appErr *AppError
		if !errors.As(err, &appErr) || appErr.Code != CodeNotFound {
			t.Fatalf("Update() error = %+v", err)
		}
	})

	t.Run("maps missing entry from update result", func(t *testing.T) {
		t.Parallel()

		store := &stubStore{
			getEntryFunc: func(context.Context, string) (Entry, error) {
				return baseEntry, nil
			},
			updateEntryFunc: func(context.Context, Entry) (bool, error) {
				return false, nil
			},
		}
		app := NewApp(store)
		value := "15.25"

		_, err := app.Update(context.Background(), "entry-1", UpdateInput{Amount: &value})
		if err == nil {
			t.Fatal("Update() error = nil, want not found")
		}

		var appErr *AppError
		if !errors.As(err, &appErr) || appErr.Code != CodeNotFound {
			t.Fatalf("Update() error = %+v", err)
		}
	})
}

func TestAppDelete(t *testing.T) {
	t.Parallel()

	t.Run("returns deleted id", func(t *testing.T) {
		t.Parallel()

		store := &stubStore{}
		app := NewApp(store)

		got, err := app.Delete(context.Background(), " entry-1 ")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if got != "entry-1" {
			t.Fatalf("Delete() id = %q", got)
		}
		if store.deleteIDInput != "entry-1" {
			t.Fatalf("DeleteEntry() id = %q", store.deleteIDInput)
		}
	})

	t.Run("maps not found from delete result", func(t *testing.T) {
		t.Parallel()

		store := &stubStore{
			deleteEntryFunc: func(context.Context, string) (bool, error) {
				return false, nil
			},
		}
		app := NewApp(store)

		_, err := app.Delete(context.Background(), "entry-1")
		if err == nil {
			t.Fatal("Delete() error = nil, want not found")
		}

		var appErr *AppError
		if !errors.As(err, &appErr) || appErr.Code != CodeNotFound {
			t.Fatalf("Delete() error = %+v", err)
		}
	})
}
