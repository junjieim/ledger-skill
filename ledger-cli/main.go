package main

import (
	"os"

	"github.com/junjieim/ledger-skill/ledger-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
