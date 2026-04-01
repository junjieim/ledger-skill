package cmd

import (
	"context"
	"errors"
	"flag"
	"io"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

func (r *Runner) runList(args []string) int {
	set := newFlagSet("list")

	var filter ledger.ListFilter
	set.StringVar(&filter.Currency, "currency", "", "filter by exact currency")
	set.StringVar(&filter.Category, "category", "", "filter by exact category")
	set.StringVar(&filter.From, "from", "", "inclusive start timestamp in RFC3339 format")
	set.StringVar(&filter.To, "to", "", "inclusive end timestamp in RFC3339 format")
	set.IntVar(&filter.Limit, "limit", 0, "maximum number of entries to return; 0 means no limit")

	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(r.Stdout, listHelpText())
			return 0
		}
		return r.writeError(ledger.NewInvalidArgumentError(err.Error()))
	}

	if set.NArg() != 0 {
		return r.writeError(ledger.NewInvalidArgumentError("list does not accept positional arguments"))
	}

	return r.withApp(func(ctx context.Context, app *ledger.App) (any, error) {
		return app.List(ctx, filter)
	})
}
