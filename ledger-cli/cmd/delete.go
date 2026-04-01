package cmd

import (
	"context"
	"errors"
	"flag"
	"io"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

func (r *Runner) runDelete(args []string) int {
	set := newFlagSet("delete")

	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(r.Stdout, deleteHelpText())
			return 0
		}
		return r.writeError(ledger.NewInvalidArgumentError(err.Error()))
	}

	if set.NArg() != 1 {
		return r.writeError(ledger.NewInvalidArgumentError("delete requires exactly one id argument"))
	}

	id := set.Arg(0)

	return r.withApp(func(ctx context.Context, app *ledger.App) (any, error) {
		deletedID, err := app.Delete(ctx, id)
		if err != nil {
			return nil, err
		}
		return map[string]string{"id": deletedID}, nil
	})
}
