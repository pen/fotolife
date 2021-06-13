package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"

	"github.com/pen/fotolife/internal/cmd"
)

var version string = "0.5.6"

func main() {
	cli := cmd.CLI{
		Options: cmd.Options{
			Logger: log.New(os.Stderr, "", log.LstdFlags),
		},
	}
	ctx := kong.Parse(&cli,
		kong.Name("fotolife"),
		kong.Description("Hatena Fotolife client"),

		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{"version": version},
	)

	ctx.FatalIfErrorf(ctx.Run(&cli.Options))
}
