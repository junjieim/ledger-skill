package cmd

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"

	ledger "github.com/junjieim/ledger-skill/ledger-cli/internal"
)

func (r *Runner) runUpdate(args []string) int {
	set := newFlagSet("update")

	var (
		datetime optionalStringFlag
		amount   optionalStringFlag
		currency optionalStringFlag
		category optionalStringFlag
		note     optionalStringFlag
	)

	set.Var(&datetime, "datetime", "updated timestamp in RFC3339 format")
	set.Var(&amount, "amount", "updated amount as a decimal string")
	set.Var(&currency, "currency", "updated currency")
	set.Var(&category, "category", "updated category")
	set.Var(&note, "note", "updated note")

	id := ""
	parseArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		id = args[0]
		parseArgs = args[1:]
	}

	if err := set.Parse(parseArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = io.WriteString(r.Stdout, updateHelpText())
			return 0
		}
		return r.writeError(ledger.NewInvalidArgumentError(err.Error()))
	}

	if id == "" {
		if set.NArg() != 1 {
			return r.writeError(ledger.NewInvalidArgumentError("update requires exactly one id argument"))
		}
		id = set.Arg(0)
	} else if set.NArg() != 0 {
		return r.writeError(ledger.NewInvalidArgumentError("update accepts a single id argument"))
	}
	input := ledger.UpdateInput{
		Datetime: datetime.Pointer(),
		Amount:   amount.Pointer(),
		Currency: currency.Pointer(),
		Category: category.Pointer(),
		Note:     note.Pointer(),
	}

	return r.withApp(func(ctx context.Context, app *ledger.App) (any, error) {
		return app.Update(ctx, id, input)
	})
}
