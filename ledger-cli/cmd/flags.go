package cmd

import (
	"flag"
	"io"
)

type optionalStringFlag struct {
	value string
	set   bool
}

func newFlagSet(name string) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	return set
}

func (f *optionalStringFlag) String() string {
	return f.value
}

func (f *optionalStringFlag) Set(value string) error {
	f.value = value
	f.set = true
	return nil
}

func (f *optionalStringFlag) Pointer() *string {
	if !f.set {
		return nil
	}

	value := f.value
	return &value
}
