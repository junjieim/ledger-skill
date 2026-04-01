package cmd

import (
	"context"
	"errors"
	"flag"
	"io"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

func (r *Runner) runSearch(args []string) int {
	set := newFlagSet("search")

	var query ledger.SearchQuery
	set.StringVar(&query.Term, "query", "", "text to search in note")
	set.IntVar(&query.Limit, "limit", 0, "maximum number of entries to return; 0 means no limit")

	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(r.Stdout, searchHelpText())
			return 0
		}
		return r.writeError(ledger.NewInvalidArgumentError(err.Error()))
	}

	if set.NArg() != 0 {
		return r.writeError(ledger.NewInvalidArgumentError("search does not accept positional arguments"))
	}

	return r.withApp(func(ctx context.Context, app *ledger.App) (any, error) {
		return app.Search(ctx, query)
	})
}
