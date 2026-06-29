package sshconfig

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/elev1e1nSure/hop/internal/apperr"
)

type boundary struct {
	line  int
	kind  string
	hosts []string
}

type includeDirective struct {
	line     int
	patterns []string
}

func parseBlocks(path string, lines []string) ([]Block, error) {
	blocks, _, err := parseFile(path, lines)
	return blocks, err
}

func parseFile(path string, lines []string) ([]Block, []includeDirective, error) {
	bounds, includes, err := walkBoundaries(path, lines)
	if err != nil {
		return nil, nil, err
	}
	blocks, err := collectBlocks(path, lines, bounds)
	if err != nil {
		return nil, nil, err
	}
	return blocks, includes, nil
}

func walkBoundaries(path string, lines []string) ([]boundary, []includeDirective, error) {
	bounds := make([]boundary, 0)
	includes := make([]includeDirective, 0)
	for index, line := range lines {
		key, args, ok, err := parseDirective(line)
		if err != nil {
			return nil, nil, apperr.New(apperr.ErrUnclosedQuote, path, index+1)
		}
		if !ok {
			continue
		}
		if requiresValue(key) && (len(args) == 0 || args[0] == "") {
			return nil, nil, apperr.New(apperr.ErrMissingDirectiveArg, path, key, index+1)
		}
		switch key {
		case "host":
			if err := validateHostPatterns(path, args, index); err != nil {
				return nil, nil, err
			}
			bounds = append(bounds, boundary{line: index, kind: "host", hosts: args})
		case "include":
			includes = append(includes, includeDirective{line: index, patterns: args})
		case "match":
			bounds = append(bounds, boundary{line: index, kind: "match"})
		}
	}
	return bounds, includes, nil
}

func validateHostPatterns(path string, args []string, line int) error {
	for _, pattern := range args {
		candidate := strings.TrimPrefix(pattern, "!")
		if candidate == "" {
			return apperr.New(apperr.ErrInvalidHostPattern, path, pattern, line+1)
		}
		if _, matchErr := filepath.Match(candidate, candidate); matchErr != nil {
			return apperr.New(apperr.ErrInvalidHostPattern, path, pattern, line+1)
		}
	}
	return nil
}

func collectBlocks(path string, lines []string, bounds []boundary) ([]Block, error) {
	blocks := make([]Block, 0)
	for boundaryIndex, current := range bounds {
		if current.kind != "host" {
			continue
		}
		end := len(lines)
		if boundaryIndex+1 < len(bounds) {
			end = bounds[boundaryIndex+1].line
		}
		values, directiveLines, err := readBlockDirectives(path, lines, current.line+1, end)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, Block{
			Start:          current.line,
			End:            end,
			HostLine:       current.line,
			Hosts:          append([]string(nil), current.hosts...),
			Values:         values,
			DirectiveLines: directiveLines,
		})
	}
	return blocks, nil
}

func readBlockDirectives(path string, lines []string, start, end int) (map[string]DirectiveValue, map[string][]int, error) {
	values := map[string]DirectiveValue{}
	directiveLines := map[string][]int{}
	for index := start; index < end; index++ {
		key, args, ok, err := parseDirective(lines[index])
		if err != nil {
			return nil, nil, apperr.New(apperr.ErrUnclosedQuote, path, index+1)
		}
		if !ok || len(args) == 0 {
			continue
		}
		switch key {
		case "hostname", "user", "port", "identityfile", "proxyjump", "proxycommand":
			directiveLines[key] = append(directiveLines[key], index)
			if _, exists := values[key]; !exists {
				values[key] = DirectiveValue{Value: args[0], Line: index + 1}
			}
		}
	}
	return values, directiveLines, nil
}

func requiresValue(key string) bool {
	switch key {
	case "host", "match", "include", "hostname", "user", "port", "identityfile", "proxyjump", "proxycommand":
		return true
	default:
		return false
	}
}

func parseDirective(line string) (string, []string, bool, error) {
	words, err := splitSSHWords(line)
	if err != nil {
		return "", nil, false, err
	}
	if len(words) == 0 {
		return "", nil, false, nil
	}
	key := words[0]
	args := words[1:]
	if equalAt := strings.IndexByte(key, '='); equalAt > 0 {
		value := key[equalAt+1:]
		key = key[:equalAt]
		if value != "" {
			args = append([]string{value}, args...)
		}
	} else if len(args) > 0 && strings.HasPrefix(args[0], "=") {
		value := strings.TrimPrefix(args[0], "=")
		args = args[1:]
		if value != "" {
			args = append([]string{value}, args...)
		}
	}
	return strings.ToLower(key), args, true, nil
}

func splitSSHWords(line string) ([]string, error) {
	words := make([]string, 0)
	var builder strings.Builder
	var quote rune
	escaped := false
	inWord := false

	flush := func() {
		if inWord {
			words = append(words, builder.String())
			builder.Reset()
			inWord = false
		}
	}

	for _, current := range line {
		if escaped {
			builder.WriteRune(current)
			inWord = true
			escaped = false
			continue
		}
		if current == '\\' {
			escaped = true
			inWord = true
			continue
		}
		if quote != 0 {
			if current == quote {
				quote = 0
			} else {
				builder.WriteRune(current)
				inWord = true
			}
			continue
		}
		if current == '\'' || current == '"' {
			quote = current
			inWord = true
			continue
		}
		if current == '#' {
			flush()
			break
		}
		if unicode.IsSpace(current) {
			flush()
			continue
		}
		builder.WriteRune(current)
		inWord = true
	}
	if quote != 0 {
		return nil, errUnclosedQuote{}
	}
	if escaped {
		builder.WriteRune('\\')
		inWord = true
	}
	flush()
	return words, nil
}

type errUnclosedQuote struct{}

func (errUnclosedQuote) Error() string { return "unclosed quote" }
