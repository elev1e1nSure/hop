package app

import (
	"fmt"
	"io"
	"time"

	"github.com/elev1e1nSure/hop/internal/config"
	"github.com/elev1e1nSure/hop/internal/history"
	"github.com/elev1e1nSure/hop/internal/i18n"
	"github.com/elev1e1nSure/hop/internal/sshclient"
	"github.com/elev1e1nSure/hop/internal/ui"
)

func Run(translator i18n.Translator, stderr io.Writer) error {
	paths, err := config.DefaultPaths()
	if err != nil {
		return err
	}

	for {
		result, err := ui.Run(paths.SSHConfig, paths.History, translator)
		if err != nil {
			return err
		}
		if result.Connect == nil {
			return nil
		}

		server := *result.Connect
		record := result.History[server.Alias]
		record.LastConnected = time.Now()
		record.Count++
		result.History[server.Alias] = record
		if err := history.Save(paths.History, result.History); err != nil {
			_, _ = fmt.Fprintln(stderr, translator.Error(err))
		}
		if err := sshclient.Run(result.Binary, server); err != nil {
			_, _ = fmt.Fprintln(stderr, translator.Error(err))
		}
	}
}
