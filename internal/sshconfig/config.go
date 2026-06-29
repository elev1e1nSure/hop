package sshconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sshm/internal/apperr"
	"sshm/internal/util"
)

type DirectiveValue struct {
	Value string
	Line  int
}

type Block struct {
	Start          int
	End            int
	HostLine       int
	Hosts          []string
	Values         map[string]DirectiveValue
	DirectiveLines map[string][]int
}

type Config struct {
	Path        string
	Lines       []string
	Newline     string
	HadTrailing bool
	Mode        os.FileMode
	Blocks      []Block
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	mode := os.FileMode(0o600)
	if errors.Is(err, os.ErrNotExist) {
		dir := filepath.Dir(path)
		if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
			cause := fmt.Errorf("create SSH directory %q: %w", dir, mkErr)
			return nil, apperr.Wrap(apperr.ErrCreateDirectory, cause, dir)
		}
		data = nil
	} else if err != nil {
		cause := fmt.Errorf("read SSH config %q: %w", path, err)
		return nil, apperr.Wrap(apperr.ErrReadSSHConfig, cause, path)
	} else {
		info, statErr := os.Stat(path)
		if statErr != nil {
			cause := fmt.Errorf("stat SSH config %q: %w", path, statErr)
			return nil, apperr.Wrap(apperr.ErrStatSSHConfig, cause, path)
		}
		mode = info.Mode().Perm()
	}

	text := string(data)
	newline := detectNewline(text)
	hadTrailing := len(data) > 0 && strings.HasSuffix(text, newline)
	if hadTrailing {
		text = strings.TrimSuffix(text, newline)
	}
	var lines []string
	if text != "" {
		lines = strings.Split(text, newline)
	}

	blocks, err := parseBlocks(path, lines)
	if err != nil {
		return nil, err
	}
	return &Config{
		Path:        path,
		Lines:       lines,
		Newline:     newline,
		HadTrailing: hadTrailing,
		Mode:        mode,
		Blocks:      blocks,
	}, nil
}

func detectNewline(text string) string {
	if index := strings.IndexByte(text, '\n'); index > 0 && text[index-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

func (config *Config) write() error {
	content := strings.Join(config.Lines, config.Newline)
	if len(config.Lines) > 0 && config.HadTrailing {
		content += config.Newline
	}
	if err := util.AtomicWrite(config.Path, []byte(content), config.Mode, 0o700); err != nil {
		cause := fmt.Errorf("write SSH config %q: %w", config.Path, err)
		return apperr.Wrap(apperr.ErrWriteSSHConfig, cause, config.Path)
	}
	return nil
}
