package cmd

import (
	"fmt"
	"io"
)

func (r *Runner) runHelp(args []string) int {
	if len(args) == 0 {
		_, _ = io.WriteString(r.Stdout, rootHelpText())
		return 0
	}

	if len(args) > 1 {
		_, _ = fmt.Fprintf(r.Stderr, "help accepts at most one command name\n")
		return 1
	}

	text, ok := commandHelpText(args[0])
	if !ok {
		_, _ = fmt.Fprintf(r.Stderr, "unknown help topic: %s\n\n%s", args[0], rootHelpText())
		return 1
	}

	_, _ = io.WriteString(r.Stdout, text)
	return 0
}

func commandHelpText(name string) (string, bool) {
	switch name {
	case "add":
		return addHelpText(), true
	case "list":
		return listHelpText(), true
	case "search":
		return searchHelpText(), true
	case "get":
		return getHelpText(), true
	case "update":
		return updateHelpText(), true
	case "delete":
		return deleteHelpText(), true
	case "help":
		return helpHelpText(), true
	default:
		return "", false
	}
}

func rootHelpText() string {
	return `Usage:
  ledger <command> [options]

Commands:
  add       Add a ledger entry
  list      List entries with exact field filters
  search    Search entries by note text
  get       Fetch a single entry by id
  update    Update a single entry by id
  delete    Delete a single entry by id
  help      Show command help

Use "ledger help <command>" for command-specific usage.
`
}

func addHelpText() string {
	return `Usage:
  ledger add --datetime <RFC3339> --amount <decimal> --currency <text> --category <text> [--note <text>]

Adds a new ledger entry and prints the created entry as JSON.
`
}

func listHelpText() string {
	return `Usage:
  ledger list [--currency <text>] [--category <text>] [--from <RFC3339>] [--to <RFC3339>] [--limit <n>]

Lists entries using exact field filters and prints the result as JSON.
`
}

func searchHelpText() string {
	return `Usage:
  ledger search --query <text> [--limit <n>]

Searches the note field with a case-insensitive match and prints the result as JSON.
`
}

func getHelpText() string {
	return `Usage:
  ledger get <id>

Fetches a single entry by id and prints it as JSON.
`
}

func updateHelpText() string {
	return `Usage:
  ledger update <id> [--datetime <RFC3339>] [--amount <decimal>] [--currency <text>] [--category <text>] [--note <text>]

Updates one or more fields on an existing entry and prints the updated entry as JSON.
`
}

func deleteHelpText() string {
	return `Usage:
  ledger delete <id>

Deletes a single entry by id and prints the deleted id as JSON.
`
}

func helpHelpText() string {
	return `Usage:
  ledger help [command]

Shows general help or detailed help for a single command.
`
}
