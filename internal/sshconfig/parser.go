package sshconfig

import (
	"path/filepath"
	"strings"
	"unicode"

	"hop/internal/apperr"
)

type boundary struct {
	line  int
	kind  string
	hosts []string
}

func parseBlocks(path string, lines []string) ([]Block, error) {
	bounds := make([]boundary, 0)
	for index, line := range lines {
		key, args, ok, err := parseDirective(line)
		if err != nil {
			return nil, apperr.New(apperr.ErrUnclosedQuote, path, index+1)
		}
		if !ok {
			continue
		}
		if requiresValue(key) && (len(args) == 0 || args[0] == "") {
			return nil, apperr.New(apperr.ErrMissingDirectiveArg, path, key, index+1)
		}
		switch key {
		case "host":
			for _, pattern := range args {
				candidate := strings.TrimPrefix(pattern, "!")
				if candidate == "" {
					return nil, apperr.New(apperr.ErrInvalidHostPattern, path, pattern, index+1)
				}
				if _, matchErr := filepath.Match(candidate, candidate); matchErr != nil {
					return nil, apperr.New(apperr.ErrInvalidHostPattern, path, pattern, index+1)
				}
			}
			bounds = append(bounds, boundary{line: index, kind: "host", hosts: args})
		case "match":
			bounds = append(bounds, boundary{line: index, kind: "match"})
		}
	}

	blocks := make([]Block, 0)
	for boundaryIndex, current := range bounds {
		if current.kind != "host" {
			continue
		}
		end := len(lines)
		if boundaryIndex+1 < len(bounds) {
			end = bounds[boundaryIndex+1].line
		}

		values := map[string]DirectiveValue{}
		directiveLines := map[string][]int{}
		for index := current.line + 1; index < end; index++ {
			key, args, ok, err := parseDirective(lines[index])
			if err != nil {
				return nil, apperr.New(apperr.ErrUnclosedQuote, path, index+1)
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

func requiresValue(key string) bool {
	switch key {
	case "host", "match", "hostname", "user", "port", "identityfile", "proxyjump", "proxycommand":
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
