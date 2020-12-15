package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"

	"github.com/pen/fotolife/internal/cmd"
)

var version string

func main() {
	cli := cmd.CLI{
		Options: cmd.Options{
			Logger: log.New(os.Stderr, "", log.LstdFlags),
		},
	}
	ctx := kong.Parse(&cli,
		kong.Name("fotolife"),
		kong.Description("Hatena Fotolife client"),
		kong.Vars{"version": version},

		kong.UsageOnMissing(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)

	ctx.FatalIfErrorf(ctx.Run(&cli.Options))
}
