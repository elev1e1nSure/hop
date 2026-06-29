package config

import (
	"fmt"
	"os"
	"path/filepath"

	"hop/internal/apperr"
)

const AppName = "hop"

type Paths struct {
	SSHConfig string
	History   string
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		cause := fmt.Errorf("resolve user home directory: %w", err)
		return Paths{}, apperr.Wrap(apperr.ErrHomeDir, cause)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(home, ".config")
	}
	return Paths{
		SSHConfig: filepath.Join(home, ".ssh", "config"),
		History:   filepath.Join(configDir, AppName, "history.json"),
	}, nil
}
