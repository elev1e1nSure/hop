package sshconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elev1e1nSure/hop/internal/apperr"
	"github.com/elev1e1nSure/hop/internal/util"
)

type DirectiveValue struct {
	Value string
	Line  int
}

type Block struct {
	SourcePath     string
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
	files       map[string]*configFile
	rootKey     string
}

type configFile struct {
	Path        string
	Lines       []string
	Newline     string
	HadTrailing bool
	Mode        os.FileMode
}

func Load(path string) (*Config, error) {
	files := map[string]*configFile{}
	blocks, rootKey, err := loadConfigTree(path, true, files, map[string]bool{})
	if err != nil {
		return nil, err
	}
	root := files[rootKey]
	return &Config{
		Path:        root.Path,
		Lines:       append([]string(nil), root.Lines...),
		Newline:     root.Newline,
		HadTrailing: root.HadTrailing,
		Mode:        root.Mode,
		Blocks:      blocks,
		files:       files,
		rootKey:     rootKey,
	}, nil
}

func loadConfigTree(path string, root bool, files map[string]*configFile, seen map[string]bool) ([]Block, string, error) {
	key := canonicalPath(path)
	if seen[key] {
		return nil, key, nil
	}
	seen[key] = true

	file, err := readConfigFile(path, root)
	if err != nil {
		return nil, key, err
	}
	if file == nil {
		return nil, key, nil
	}
	file.Path = key
	files[key] = file

	localBlocks, includes, err := parseFile(key, file.Lines)
	if err != nil {
		return nil, key, err
	}
	for index := range localBlocks {
		localBlocks[index].SourcePath = key
	}

	blocks, err := expandIncludes(key, localBlocks, includes, files, seen)
	if err != nil {
		return nil, key, err
	}
	return blocks, key, nil
}

func readConfigFile(path string, root bool) (*configFile, error) {
	data, err := os.ReadFile(path)
	mode := os.FileMode(0o600)
	if errors.Is(err, os.ErrNotExist) {
		if !root {
			return nil, nil
		}
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

	return &configFile{
		Path:        path,
		Lines:       lines,
		Newline:     newline,
		HadTrailing: hadTrailing,
		Mode:        mode,
	}, nil
}

func expandIncludes(
	path string,
	localBlocks []Block,
	includes []includeDirective,
	files map[string]*configFile,
	seen map[string]bool,
) ([]Block, error) {
	events := make([]configEvent, 0, len(localBlocks)+len(includes))
	for _, block := range localBlocks {
		events = append(events, configEvent{line: block.Start, block: &block})
	}
	for _, include := range includes {
		include := include
		events = append(events, configEvent{line: include.line, include: &include})
	}
	sort.SliceStable(events, func(left, right int) bool {
		return events[left].line < events[right].line
	})

	blocks := make([]Block, 0, len(localBlocks))
	for _, event := range events {
		if event.block != nil {
			blocks = append(blocks, *event.block)
			continue
		}
		paths, err := resolveIncludePaths(path, event.include.patterns)
		if err != nil {
			return nil, err
		}
		for _, includePath := range paths {
			childBlocks, _, err := loadConfigTree(includePath, false, files, seen)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, childBlocks...)
		}
	}
	return blocks, nil
}

type configEvent struct {
	line    int
	block   *Block
	include *includeDirective
}

func resolveIncludePaths(basePath string, patterns []string) ([]string, error) {
	paths := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		resolved := resolveIncludePattern(basePath, pattern)
		matches, err := filepath.Glob(resolved)
		if err != nil {
			cause := fmt.Errorf("expand Include pattern %q in %q: %w", pattern, basePath, err)
			return nil, apperr.Wrap(apperr.ErrReadSSHConfig, cause, basePath)
		}
		if len(matches) == 0 && !hasGlobMeta(resolved) {
			if _, err := os.Stat(resolved); err == nil {
				matches = []string{resolved}
			}
		}
		sort.Strings(matches)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			paths = append(paths, match)
		}
	}
	return paths, nil
}

func resolveIncludePattern(basePath, pattern string) string {
	pattern = expandHome(pattern)
	if filepath.IsAbs(pattern) {
		return pattern
	}
	return filepath.Join(filepath.Dir(basePath), pattern)
}

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

func hasGlobMeta(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func canonicalPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(absolute)
}

func detectNewline(text string) string {
	if index := strings.IndexByte(text, '\n'); index > 0 && text[index-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

func writeConfigFile(file *configFile) error {
	content := strings.Join(file.Lines, file.Newline)
	if len(file.Lines) > 0 && file.HadTrailing {
		content += file.Newline
	}
	if err := util.AtomicWrite(file.Path, []byte(content), file.Mode, 0o700); err != nil {
		cause := fmt.Errorf("write SSH config %q: %w", file.Path, err)
		return apperr.Wrap(apperr.ErrWriteSSHConfig, cause, file.Path)
	}
	return nil
}

func (config *Config) syncFromDisk() error {
	fresh, err := Load(config.Path)
	if err != nil {
		return err
	}
	*config = *fresh
	return nil
}

func (config *Config) rootFile() *configFile {
	if file, ok := config.files[config.rootKey]; ok {
		return file
	}
	return &configFile{
		Path:        config.Path,
		Lines:       config.Lines,
		Newline:     config.Newline,
		HadTrailing: config.HadTrailing,
		Mode:        config.Mode,
	}
}

func (config *Config) fileForBlock(block Block) *configFile {
	if file, ok := config.files[block.SourcePath]; ok {
		return file
	}
	return config.rootFile()
}

func (config *Config) mirrorRoot(file *configFile) {
	if file.Path != config.rootKey {
		return
	}
	config.Lines = append([]string(nil), file.Lines...)
	config.Newline = file.Newline
	config.HadTrailing = file.HadTrailing
	config.Mode = file.Mode
}
