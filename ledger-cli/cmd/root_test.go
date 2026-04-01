package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

type responseEnvelope struct {
	Success bool              `json:"success"`
	Data    json.RawMessage   `json:"data"`
	Error   *ledger.ErrorInfo `json:"error"`
}

func TestRunnerAddSuccess(t *testing.T) {
	t.Parallel()

	runner := newTestRunner(t)
	var stdout bytes.Buffer
	runner.Stdout = &stdout

	code := runner.Run([]string{
		"add",
		"--datetime", "2026-04-01T08:00:00+08:00",
		"--amount", "10.50",
		"--currency", "HKD",
		"--category", "food",
		"--note", "breakfast",
	})
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}

	var response responseEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !response.Success {
		t.Fatalf("response success = false, error = %+v", response.Error)
	}

	var entry ledger.Entry
	if err := json.Unmarshal(response.Data, &entry); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
	if entry.Currency != "HKD" || entry.Category != "food" {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestRunnerGetNotFound(t *testing.T) {
	t.Parallel()

	runner := newTestRunner(t)
	var stdout bytes.Buffer
	runner.Stdout = &stdout

	code := runner.Run([]string{"get", "missing-id"})
	if code == 0 {
		t.Fatal("Run() code = 0, want non-zero")
	}

	var response responseEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Success {
		t.Fatal("response success = true, want false")
	}
	if response.Error == nil || response.Error.Code != ledger.CodeNotFound {
		t.Fatalf("response error = %+v, want not_found", response.Error)
	}
}

func TestRunnerUpdateRequiresFields(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "ledger.db")
	seedTestEntry(t, dbPath, ledger.Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "10.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "breakfast",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	})

	runner := newRunnerWithDBPath(t, dbPath)
	var stdout bytes.Buffer
	runner.Stdout = &stdout

	code := runner.Run([]string{"update", "entry-1"})
	if code == 0 {
		t.Fatal("Run() code = 0, want non-zero")
	}

	var response responseEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Error == nil || response.Error.Code != ledger.CodeInvalidArgument {
		t.Fatalf("response error = %+v, want invalid_argument", response.Error)
	}
}

func TestRunnerHelpText(t *testing.T) {
	t.Parallel()

	runner := newTestRunner(t)
	var stdout bytes.Buffer
	runner.Stdout = &stdout

	code := runner.Run([]string{"help", "add"})
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}

	output := stdout.String()
	if !strings.Contains(output, "ledger add --datetime") {
		t.Fatalf("help output = %q", output)
	}
	if strings.Contains(output, `"success"`) {
		t.Fatalf("help output should be text, got %q", output)
	}
}

func TestRunnerListAndSearchSuccess(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "ledger.db")
	seedTestEntry(t, dbPath, ledger.Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "10.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "breakfast with team",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	})
	seedTestEntry(t, dbPath, ledger.Entry{
		ID:        "entry-2",
		Datetime:  "2026-04-02T00:00:00Z",
		Amount:    "20.00",
		Currency:  "USD",
		Category:  "travel",
		Note:      "taxi to airport",
		CreatedAt: "2026-04-02T00:00:00Z",
		UpdatedAt: "2026-04-02T00:00:00Z",
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()

		runner := newRunnerWithDBPath(t, dbPath)
		var stdout bytes.Buffer
		runner.Stdout = &stdout

		code := runner.Run([]string{"list", "--currency", "HKD", "--limit", "1"})
		if code != 0 {
			t.Fatalf("Run() code = %d, want 0", code)
		}

		var response responseEnvelope
		if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if !response.Success {
			t.Fatalf("response = %+v", response)
		}

		var entries []ledger.Entry
		if err := json.Unmarshal(response.Data, &entries); err != nil {
			t.Fatalf("json.Unmarshal(data) error = %v", err)
		}
		if len(entries) != 1 || entries[0].ID != "entry-1" {
			t.Fatalf("entries = %+v", entries)
		}
	})

	t.Run("search", func(t *testing.T) {
		t.Parallel()

		runner := newRunnerWithDBPath(t, dbPath)
		var stdout bytes.Buffer
		runner.Stdout = &stdout

		code := runner.Run([]string{"search", "--query", "taxi"})
		if code != 0 {
			t.Fatalf("Run() code = %d, want 0", code)
		}

		var response responseEnvelope
		if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		var entries []ledger.Entry
		if err := json.Unmarshal(response.Data, &entries); err != nil {
			t.Fatalf("json.Unmarshal(data) error = %v", err)
		}
		if len(entries) != 1 || entries[0].ID != "entry-2" {
			t.Fatalf("entries = %+v", entries)
		}
	})
}

func TestRunnerDeleteSuccess(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "ledger.db")
	seedTestEntry(t, dbPath, ledger.Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "10.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "breakfast",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	})

	runner := newRunnerWithDBPath(t, dbPath)
	var stdout bytes.Buffer
	runner.Stdout = &stdout

	code := runner.Run([]string{"delete", "entry-1"})
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}

	var response responseEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var deleted map[string]string
	if err := json.Unmarshal(response.Data, &deleted); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
	if deleted["id"] != "entry-1" {
		t.Fatalf("deleted = %+v", deleted)
	}
}

func TestRunnerUpdateWithPositionalIDAndFlags(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "ledger.db")
	seedTestEntry(t, dbPath, ledger.Entry{
		ID:        "entry-1",
		Datetime:  "2026-04-01T00:00:00Z",
		Amount:    "10.00",
		Currency:  "HKD",
		Category:  "food",
		Note:      "breakfast",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	})

	runner := newRunnerWithDBPath(t, dbPath)
	var stdout bytes.Buffer
	runner.Stdout = &stdout

	code := runner.Run([]string{"update", "entry-1", "--note", "dinner"})
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}

	var response responseEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var entry ledger.Entry
	if err := json.Unmarshal(response.Data, &entry); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
	if entry.Note != "dinner" {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestRunnerParameterAndDispatchErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "root help",
			args:       nil,
			wantCode:   0,
			wantStdout: "Usage:\n  ledger <command> [options]",
		},
		{
			name:       "unknown command",
			args:       []string{"unknown"},
			wantCode:   1,
			wantStderr: "unknown command: unknown",
		},
		{
			name:       "help unknown topic",
			args:       []string{"help", "unknown"},
			wantCode:   1,
			wantStderr: "unknown help topic: unknown",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			runner := newTestRunner(t)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			runner.Stdout = &stdout
			runner.Stderr = &stderr

			code := runner.Run(test.args)
			if code != test.wantCode {
				t.Fatalf("Run() code = %d, want %d", code, test.wantCode)
			}
			if test.wantStdout != "" && !strings.Contains(stdout.String(), test.wantStdout) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), test.wantStdout)
			}
			if test.wantStderr != "" && !strings.Contains(stderr.String(), test.wantStderr) {
				t.Fatalf("stderr = %q, want substring %q", stderr.String(), test.wantStderr)
			}
		})
	}
}

func TestRunnerCommandValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantMsg string
	}{
		{
			name:    "add missing amount",
			args:    []string{"add", "--datetime", "2026-04-01T00:00:00Z", "--currency", "HKD", "--category", "food"},
			wantMsg: "amount is required",
		},
		{
			name:    "add unsupported currency",
			args:    []string{"add", "--datetime", "2026-04-01T00:00:00Z", "--amount", "12.50", "--currency", "CNY", "--category", "food"},
			wantMsg: "currency must be one of: RMB, HKD, USD, EUR, JPY, GBP, AUD, CAD, SGD, TWD",
		},
		{
			name:    "list positional argument",
			args:    []string{"list", "extra"},
			wantMsg: "list does not accept positional arguments",
		},
		{
			name:    "search positional argument",
			args:    []string{"search", "--query", "note", "extra"},
			wantMsg: "search does not accept positional arguments",
		},
		{
			name:    "delete missing id",
			args:    []string{"delete"},
			wantMsg: "delete requires exactly one id argument",
		},
		{
			name:    "update extra positional",
			args:    []string{"update", "entry-1", "extra"},
			wantMsg: "update accepts a single id argument",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			runner := newTestRunner(t)
			var stdout bytes.Buffer
			runner.Stdout = &stdout

			code := runner.Run(test.args)
			if code == 0 {
				t.Fatal("Run() code = 0, want non-zero")
			}

			var response responseEnvelope
			if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if response.Error == nil || response.Error.Code != ledger.CodeInvalidArgument || response.Error.Message != test.wantMsg {
				t.Fatalf("response error = %+v", response.Error)
			}
		})
	}
}

func TestRunnerHelpTopics(t *testing.T) {
	t.Parallel()

	topics := []struct {
		name      string
		topic     string
		wantUsage string
	}{
		{name: "list", topic: "list", wantUsage: "ledger list"},
		{name: "search", topic: "search", wantUsage: "ledger search"},
		{name: "get", topic: "get", wantUsage: "ledger get"},
		{name: "update", topic: "update", wantUsage: "ledger update"},
		{name: "delete", topic: "delete", wantUsage: "ledger delete"},
		{name: "help", topic: "help", wantUsage: "ledger help"},
	}

	for _, topic := range topics {
		topic := topic
		t.Run(topic.name, func(t *testing.T) {
			t.Parallel()

			runner := newTestRunner(t)
			var stdout bytes.Buffer
			runner.Stdout = &stdout

			code := runner.Run([]string{"help", topic.topic})
			if code != 0 {
				t.Fatalf("Run() code = %d, want 0", code)
			}
			if !strings.Contains(stdout.String(), topic.wantUsage) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), topic.wantUsage)
			}
		})
	}
}

func TestRunnerExecuteAndDefaultDBPath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute(nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "ledger <command>") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	path, err := defaultDBPath()
	if err != nil {
		t.Fatalf("defaultDBPath() error = %v", err)
	}
	if !strings.Contains(filepath.ToSlash(path), "/data/ledger.db") {
		t.Fatalf("defaultDBPath() = %q", path)
	}
}

func TestDefaultAppFactory(t *testing.T) {
	t.Parallel()

	app, err := defaultAppFactory()(context.Background())
	if err != nil {
		t.Fatalf("defaultAppFactory() error = %v", err)
	}
	defer func() {
		_ = app.Close()
	}()
}

func TestRunnerWriteFallbacks(t *testing.T) {
	t.Parallel()

	runner := newTestRunner(t)
	runner.Stdout = failingWriter{}
	var stderr bytes.Buffer
	runner.Stderr = &stderr

	if code := runner.writeSuccess(map[string]string{"id": "entry-1"}); code != 1 {
		t.Fatalf("writeSuccess() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "failed to write response") {
		t.Fatalf("stderr = %q", stderr.String())
	}

	stderr.Reset()
	if code := runner.writeError(errors.New("boom")); code != 1 {
		t.Fatalf("writeError() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "failed to write response") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	return newRunnerWithDBPath(t, filepath.Join(t.TempDir(), "ledger.db"))
}

func newRunnerWithDBPath(t *testing.T, dbPath string) *Runner {
	t.Helper()

	return NewRunner(nil, nil, func(ctx context.Context) (*ledger.App, error) {
		store, err := ledger.OpenSQLiteStore(ctx, dbPath)
		if err != nil {
			return nil, err
		}
		return ledger.NewApp(store), nil
	})
}

func seedTestEntry(t *testing.T, dbPath string, entry ledger.Entry) {
	t.Helper()

	store, err := ledger.OpenSQLiteStore(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if err := store.CreateEntry(context.Background(), entry); err != nil {
		t.Fatalf("CreateEntry() error = %v", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}
