package ledger

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

var amountPattern = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

type Entry struct {
	ID        string `json:"id"`
	Datetime  string `json:"datetime"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Category  string `json:"category"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateInput struct {
	Datetime string
	Amount   string
	Currency string
	Category string
	Note     string
}

type UpdateInput struct {
	Datetime *string
	Amount   *string
	Currency *string
	Category *string
	Note     *string
}

type ListFilter struct {
	Currency string
	Category string
	From     string
	To       string
	Limit    int
}

type SearchQuery struct {
	Term  string
	Limit int
}

func ValidateCreateInput(input CreateInput) (CreateInput, error) {
	datetime, err := normalizeDatetime(input.Datetime)
	if err != nil {
		return CreateInput{}, err
	}

	amount, err := normalizeAmount(input.Amount)
	if err != nil {
		return CreateInput{}, err
	}

	currency, err := normalizeRequiredText("currency", input.Currency)
	if err != nil {
		return CreateInput{}, err
	}

	category, err := normalizeRequiredText("category", input.Category)
	if err != nil {
		return CreateInput{}, err
	}

	return CreateInput{
		Datetime: datetime,
		Amount:   amount,
		Currency: currency,
		Category: category,
		Note:     normalizeNote(input.Note),
	}, nil
}

func ApplyUpdate(entry Entry, update UpdateInput) (Entry, error) {
	if !update.hasValues() {
		return Entry{}, NewInvalidArgumentError("at least one field must be provided for update")
	}

	merged := entry

	if update.Datetime != nil {
		merged.Datetime = *update.Datetime
	}
	if update.Amount != nil {
		merged.Amount = *update.Amount
	}
	if update.Currency != nil {
		merged.Currency = *update.Currency
	}
	if update.Category != nil {
		merged.Category = *update.Category
	}
	if update.Note != nil {
		merged.Note = *update.Note
	}

	normalized, err := ValidateCreateInput(CreateInput{
		Datetime: merged.Datetime,
		Amount:   merged.Amount,
		Currency: merged.Currency,
		Category: merged.Category,
		Note:     merged.Note,
	})
	if err != nil {
		return Entry{}, err
	}

	merged.Datetime = normalized.Datetime
	merged.Amount = normalized.Amount
	merged.Currency = normalized.Currency
	merged.Category = normalized.Category
	merged.Note = normalized.Note

	return merged, nil
}

func NormalizeListFilter(filter ListFilter) (ListFilter, error) {
	if filter.Limit < 0 {
		return ListFilter{}, NewInvalidArgumentError("limit must be greater than or equal to 0")
	}

	normalized := ListFilter{
		Currency: strings.TrimSpace(filter.Currency),
		Category: strings.TrimSpace(filter.Category),
		Limit:    filter.Limit,
	}

	if filter.From != "" {
		value, err := normalizeDatetime(filter.From)
		if err != nil {
			return ListFilter{}, NewInvalidArgumentError("from must be a valid RFC3339 timestamp")
		}
		normalized.From = value
	}

	if filter.To != "" {
		value, err := normalizeDatetime(filter.To)
		if err != nil {
			return ListFilter{}, NewInvalidArgumentError("to must be a valid RFC3339 timestamp")
		}
		normalized.To = value
	}

	if normalized.From != "" && normalized.To != "" && normalized.From > normalized.To {
		return ListFilter{}, NewInvalidArgumentError("from must be earlier than or equal to to")
	}

	return normalized, nil
}

func NormalizeSearchQuery(query SearchQuery) (SearchQuery, error) {
	if query.Limit < 0 {
		return SearchQuery{}, NewInvalidArgumentError("limit must be greater than or equal to 0")
	}

	term := strings.TrimSpace(query.Term)
	if term == "" {
		return SearchQuery{}, NewInvalidArgumentError("query must not be empty")
	}

	return SearchQuery{
		Term:  term,
		Limit: query.Limit,
	}, nil
}

func GenerateID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(raw[:]), nil
}

func (u UpdateInput) hasValues() bool {
	return u.Datetime != nil ||
		u.Amount != nil ||
		u.Currency != nil ||
		u.Category != nil ||
		u.Note != nil
}

func normalizeDatetime(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", NewInvalidArgumentError("datetime is required")
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return "", NewInvalidArgumentError("datetime must be a valid RFC3339 timestamp")
	}

	return parsed.UTC().Format(time.RFC3339), nil
}

func normalizeAmount(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", NewInvalidArgumentError("amount is required")
	}

	if !amountPattern.MatchString(trimmed) {
		return "", NewInvalidArgumentError("amount must be a valid decimal string")
	}

	return trimmed, nil
}

func normalizeRequiredText(field string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", NewInvalidArgumentError(field + " is required")
	}

	return trimmed, nil
}

func normalizeNote(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}

	return trimmed
}
