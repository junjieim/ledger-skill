package cmd

import (
	"context"
	"errors"
	"flag"
	"io"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

func (r *Runner) runGet(args []string) int {
	set := newFlagSet("get")

	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(r.Stdout, getHelpText())
			return 0
		}
		return r.writeError(ledger.NewInvalidArgumentError(err.Error()))
	}

	if set.NArg() != 1 {
		return r.writeError(ledger.NewInvalidArgumentError("get requires exactly one id argument"))
	}

	id := set.Arg(0)

	return r.withApp(func(ctx context.Context, app *ledger.App) (any, error) {
		return app.Get(ctx, id)
	})
}
