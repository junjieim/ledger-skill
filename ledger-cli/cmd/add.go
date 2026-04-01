package cmd

import (
	"context"
	"errors"
	"flag"
	"io"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

func (r *Runner) runAdd(args []string) int {
	set := newFlagSet("add")

	var input ledger.CreateInput
	set.StringVar(&input.Datetime, "datetime", "", "entry timestamp in RFC3339 format")
	set.StringVar(&input.Amount, "amount", "", "entry amount as a decimal string")
	set.StringVar(&input.Currency, "currency", "", "entry currency")
	set.StringVar(&input.Category, "category", "", "entry category")
	set.StringVar(&input.Note, "note", "", "entry note")

	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(r.Stdout, addHelpText())
			return 0
		}
		return r.writeError(ledger.NewInvalidArgumentError(err.Error()))
	}

	if set.NArg() != 0 {
		return r.writeError(ledger.NewInvalidArgumentError("add does not accept positional arguments"))
	}

	return r.withApp(func(ctx context.Context, app *ledger.App) (any, error) {
		return app.Add(ctx, input)
	})
}
