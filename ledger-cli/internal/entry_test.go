package ledger

import "testing"

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     CreateInput
		wantNote  string
		wantTime  string
		wantError string
	}{
		{
			name: "normalizes datetime and blank note",
			input: CreateInput{
				Datetime: "2026-04-01T08:00:00+08:00",
				Amount:   "12.50",
				Currency: "HKD",
				Category: "food",
				Note:     "   ",
			},
			wantNote: "none",
			wantTime: "2026-04-01T00:00:00Z",
		},
		{
			name: "rejects invalid amount",
			input: CreateInput{
				Datetime: "2026-04-01T08:00:00+08:00",
				Amount:   "12.",
				Currency: "HKD",
				Category: "food",
				Note:     "breakfast",
			},
			wantError: "amount must be a valid decimal string",
		},
		{
			name: "rejects missing currency",
			input: CreateInput{
				Datetime: "2026-04-01T08:00:00+08:00",
				Amount:   "12.50",
				Currency: " ",
				Category: "food",
				Note:     "breakfast",
			},
			wantError: "currency is required",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateCreateInput(test.input)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("ValidateCreateInput() error = nil, want %q", test.wantError)
				}
				if err.Error() != test.wantError {
					t.Fatalf("ValidateCreateInput() error = %q, want %q", err.Error(), test.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("ValidateCreateInput() error = %v", err)
			}
			if got.Note != test.wantNote {
				t.Fatalf("ValidateCreateInput() note = %q, want %q", got.Note, test.wantNote)
			}
			if got.Datetime != test.wantTime {
				t.Fatalf("ValidateCreateInput() datetime = %q, want %q", got.Datetime, test.wantTime)
			}
		})
	}
}

func TestApplyUpdate(t *testing.T) {
	t.Parallel()

	entry := Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "20.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "dinner",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}

	t.Run("requires at least one field", func(t *testing.T) {
		t.Parallel()

		_, err := ApplyUpdate(entry, UpdateInput{})
		if err == nil {
			t.Fatal("ApplyUpdate() error = nil, want invalid argument error")
		}
		if err.Error() != "at least one field must be provided for update" {
			t.Fatalf("ApplyUpdate() error = %q", err.Error())
		}
	})

	t.Run("normalizes updated values", func(t *testing.T) {
		t.Parallel()

		note := "   "
		datetime := "2026-04-02T08:00:00+08:00"

		got, err := ApplyUpdate(entry, UpdateInput{
			Datetime: &datetime,
			Note:     &note,
		})
		if err != nil {
			t.Fatalf("ApplyUpdate() error = %v", err)
		}

		if got.Datetime != "2026-04-02T00:00:00Z" {
			t.Fatalf("ApplyUpdate() datetime = %q", got.Datetime)
		}
		if got.Note != "none" {
			t.Fatalf("ApplyUpdate() note = %q", got.Note)
		}
	})
}

func TestNormalizeListFilter(t *testing.T) {
	t.Parallel()

	_, err := NormalizeListFilter(ListFilter{
		From: "2026-04-02T00:00:00Z",
		To:   "2026-04-01T00:00:00Z",
	})
	if err == nil {
		t.Fatal("NormalizeListFilter() error = nil, want validation error")
	}
	if err.Error() != "from must be earlier than or equal to to" {
		t.Fatalf("NormalizeListFilter() error = %q", err.Error())
	}
}

func TestNormalizeSearchQuery(t *testing.T) {
	t.Parallel()

	got, err := NormalizeSearchQuery(SearchQuery{Term: "  lunch  ", Limit: 2})
	if err != nil {
		t.Fatalf("NormalizeSearchQuery() error = %v", err)
	}
	if got.Term != "lunch" {
		t.Fatalf("NormalizeSearchQuery() term = %q", got.Term)
	}
}
