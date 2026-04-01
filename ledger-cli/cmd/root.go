package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

const cliName = "ledger"

type AppFactory func(context.Context) (*ledger.App, error)

type Runner struct {
	Stdout     io.Writer
	Stderr     io.Writer
	AppFactory AppFactory
}

func Execute(args []string, stdout io.Writer, stderr io.Writer) int {
	return NewRunner(stdout, stderr, defaultAppFactory()).Run(args)
}

func NewRunner(stdout io.Writer, stderr io.Writer, appFactory AppFactory) *Runner {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if appFactory == nil {
		appFactory = defaultAppFactory()
	}

	return &Runner{
		Stdout:     stdout,
		Stderr:     stderr,
		AppFactory: appFactory,
	}
}

func (r *Runner) Run(args []string) int {
	if len(args) == 0 {
		_, _ = io.WriteString(r.Stdout, rootHelpText())
		return 0
	}

	switch args[0] {
	case "add":
		return r.runAdd(args[1:])
	case "list":
		return r.runList(args[1:])
	case "search":
		return r.runSearch(args[1:])
	case "get":
		return r.runGet(args[1:])
	case "update":
		return r.runUpdate(args[1:])
	case "delete":
		return r.runDelete(args[1:])
	case "help", "-h", "--help":
		return r.runHelp(args[1:])
	default:
		_, _ = fmt.Fprintf(r.Stderr, "unknown command: %s\n\n%s", args[0], rootHelpText())
		return 1
	}
}

func defaultAppFactory() AppFactory {
	return func(ctx context.Context) (*ledger.App, error) {
		dbPath, err := defaultDBPath()
		if err != nil {
			return nil, err
		}

		store, err := ledger.OpenSQLiteStore(ctx, dbPath)
		if err != nil {
			return nil, err
		}

		return ledger.NewApp(store), nil
	}
}

func defaultDBPath() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	return filepath.Clean(filepath.Join(filepath.Dir(executablePath), "..", "data", "ledger.db")), nil
}

func (r *Runner) writeSuccess(data any) int {
	if err := ledger.WriteResponse(r.Stdout, ledger.SuccessResponse(data)); err != nil {
		_, _ = fmt.Fprintf(r.Stderr, "failed to write response: %v\n", err)
		return 1
	}

	return 0
}

func (r *Runner) writeError(err error) int {
	if err := ledger.WriteResponse(r.Stdout, ledger.ErrorResponse(err)); err != nil {
		_, _ = fmt.Fprintf(r.Stderr, "failed to write response: %v\n", err)
		return 1
	}

	return 1
}

func (r *Runner) withApp(fn func(context.Context, *ledger.App) (any, error)) int {
	app, err := r.AppFactory(context.Background())
	if err != nil {
		return r.writeError(err)
	}
	defer func() {
		_ = app.Close()
	}()

	data, err := fn(context.Background(), app)
	if err != nil {
		return r.writeError(err)
	}

	return r.writeSuccess(data)
}
