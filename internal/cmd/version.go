package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type Version struct{}

func (Version) Run(vars kong.Vars) error {
	fmt.Println(vars["version"])
	return nil
}

type PrintVersion bool

func (PrintVersion) Decode(ctx *kong.DecodeContext) error { return nil }
func (PrintVersion) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	Version{}.Run(vars) //nolint:errcheck
	app.Exit(0)

	return nil
}
