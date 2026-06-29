package main

import (
	"fmt"
	"os"

	"hop/internal/app"
	"hop/internal/cli"
	"hop/internal/i18n"
)

func main() {
	options, err := cli.Parse(os.Args[1:], os.Getenv)
	translator := i18n.New(options.Language)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, translator.Error(err))
		os.Exit(2)
	}
	if err := app.Run(translator, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, translator.Error(err))
		os.Exit(1)
	}
}
