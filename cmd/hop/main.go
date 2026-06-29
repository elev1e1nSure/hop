package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/hop/internal/app"
	"github.com/elev1e1nSure/hop/internal/cli"
	"github.com/elev1e1nSure/hop/internal/i18n"
	"github.com/elev1e1nSure/hop/internal/pathenv"
)

func main() {
	options, err := cli.Parse(os.Args[1:], os.Getenv)
	translator := i18n.New(options.Language)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, cli.RenderError(translator, err))
		os.Exit(2)
	}
	if options.Help {
		_, _ = fmt.Fprintln(os.Stdout, cli.RenderHelp(translator))
		return
	}
	if options.Path != cli.PathActionNone {
		result, err := pathenv.Apply(pathenv.Action(options.Path))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, cli.RenderError(translator, err))
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, cli.RenderPathResult(translator, options.Path, result.Directory, result.Changed))
		return
	}
	if err := app.Run(translator, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, cli.RenderError(translator, err))
		os.Exit(1)
	}
}
