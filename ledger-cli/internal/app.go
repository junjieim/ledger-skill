package ledger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type App struct {
	store Store
	now   func() time.Time
	newID func() (string, error)
}

func NewApp(store Store) *App {
	return &App{
		store: store,
		now:   time.Now,
		newID: GenerateID,
	}
}

func (a *App) Close() error {
	if a == nil || a.store == nil {
		return nil
	}

	return a.store.Close()
}

func (a *App) Add(ctx context.Context, input CreateInput) (Entry, error) {
	normalized, err := ValidateCreateInput(input)
	if err != nil {
		return Entry{}, err
	}

	id, err := a.newID()
	if err != nil {
		return Entry{}, fmt.Errorf("generate entry id: %w", err)
	}

	now := a.now().UTC().Format(time.RFC3339)
	entry := Entry{
		ID:        id,
		Datetime:  normalized.Datetime,
		Amount:    normalized.Amount,
		Currency:  normalized.Currency,
		Category:  normalized.Category,
		Note:      normalized.Note,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := a.store.CreateEntry(ctx, entry); err != nil {
		return Entry{}, err
	}

	return entry, nil
}

func (a *App) Get(ctx context.Context, id string) (Entry, error) {
	normalizedID, err := normalizeID(id)
	if err != nil {
		return Entry{}, err
	}

	entry, err := a.store.GetEntry(ctx, normalizedID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Entry{}, NewNotFoundError("entry not found")
		}
		return Entry{}, err
	}

	return entry, nil
}

func (a *App) List(ctx context.Context, filter ListFilter) ([]Entry, error) {
	normalized, err := NormalizeListFilter(filter)
	if err != nil {
		return nil, err
	}

	return a.store.ListEntries(ctx, normalized)
}

func (a *App) Search(ctx context.Context, query SearchQuery) ([]Entry, error) {
	normalized, err := NormalizeSearchQuery(query)
	if err != nil {
		return nil, err
	}

	return a.store.SearchEntries(ctx, normalized)
}

func (a *App) Update(ctx context.Context, id string, input UpdateInput) (Entry, error) {
	normalizedID, err := normalizeID(id)
	if err != nil {
		return Entry{}, err
	}

	entry, err := a.store.GetEntry(ctx, normalizedID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Entry{}, NewNotFoundError("entry not found")
		}
		return Entry{}, err
	}

	updated, err := ApplyUpdate(entry, input)
	if err != nil {
		return Entry{}, err
	}

	updated.UpdatedAt = a.now().UTC().Format(time.RFC3339)

	ok, err := a.store.UpdateEntry(ctx, updated)
	if err != nil {
		return Entry{}, err
	}
	if !ok {
		return Entry{}, NewNotFoundError("entry not found")
	}

	return updated, nil
}

func (a *App) Delete(ctx context.Context, id string) (string, error) {
	normalizedID, err := normalizeID(id)
	if err != nil {
		return "", err
	}

	ok, err := a.store.DeleteEntry(ctx, normalizedID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", NewNotFoundError("entry not found")
	}

	return normalizedID, nil
}

func normalizeID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return "", NewInvalidArgumentError("id is required")
	}

	return trimmed, nil
}
