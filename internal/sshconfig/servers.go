package sshconfig

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"sshm/internal/apperr"
	"sshm/internal/domain"
)

func ResolveServers(config *Config, records map[string]domain.HistoryRecord) ([]domain.Server, error) {
	type aliasOrigin struct {
		alias      string
		blockIndex int
	}
	origins := make([]aliasOrigin, 0)
	seen := map[string]bool{}
	for index, block := range config.Blocks {
		for _, host := range block.Hosts {
			if isConcreteAlias(host) && !seen[host] {
				seen[host] = true
				origins = append(origins, aliasOrigin{alias: host, blockIndex: index})
			}
		}
	}

	servers := make([]domain.Server, 0, len(origins))
	for _, origin := range origins {
		values := map[string]DirectiveValue{}
		hasProxy := false
		for _, block := range config.Blocks {
			if !hostBlockMatches(origin.alias, block.Hosts) {
				continue
			}
			for _, key := range []string{"hostname", "user", "port", "identityfile"} {
				if _, exists := values[key]; !exists {
					if value, ok := block.Values[key]; ok {
						values[key] = value
					}
				}
			}
			if _, ok := block.Values["proxyjump"]; ok {
				hasProxy = true
			}
			if _, ok := block.Values["proxycommand"]; ok {
				hasProxy = true
			}
		}

		host := values["hostname"].Value
		if host == "" || host == "%h" {
			host = origin.alias
		}
		host = strings.ReplaceAll(host, "%h", origin.alias)
		port := 22
		if raw, exists := values["port"]; exists {
			parsed, err := strconv.Atoi(raw.Value)
			if err != nil || parsed < 1 || parsed > 65535 {
				return nil, apperr.New(apperr.ErrInvalidPortConfig, config.Path, raw.Value, raw.Line)
			}
			port = parsed
		}
		record := records[origin.alias]
		servers = append(servers, domain.Server{
			Alias:        origin.alias,
			Host:         host,
			User:         values["user"].Value,
			Port:         port,
			IdentityFile: values["identityfile"].Value,
			BlockIndex:   origin.blockIndex,
			HasProxy:     hasProxy,
			LastUsed:     record.LastConnected,
			UseCount:     record.Count,
		})
	}
	return servers, nil
}

func AddServer(config *Config, server domain.Server) error {
	saved := append([]string(nil), config.Lines...)
	savedTrailing := config.HadTrailing

	lines := append([]string(nil), config.Lines...)
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	if len(config.Lines) == 0 {
		config.HadTrailing = true
	}
	lines = append(lines, renderServerBlock(server)...)
	config.Lines = lines
	if err := config.write(); err != nil {
		config.Lines = saved
		config.HadTrailing = savedTrailing
		return err
	}
	return nil
}

func EditServer(config *Config, oldAlias string, updated domain.Server) error {
	blockIndex := findOriginBlock(config, oldAlias)
	if blockIndex < 0 {
		return apperr.New(apperr.ErrServerNotFound, oldAlias, config.Path)
	}
	saved := append([]string(nil), config.Lines...)
	block := config.Blocks[blockIndex]
	if len(block.Hosts) == 1 && block.Hosts[0] == oldAlias {
		config.Lines = rewriteSingleHostBlock(config.Lines, block, updated)
		if err := config.write(); err != nil {
			config.Lines = saved
			return err
		}
		return nil
	}

	remaining := make([]string, 0, len(block.Hosts)-1)
	for _, host := range block.Hosts {
		if host != oldAlias {
			remaining = append(remaining, host)
		}
	}
	if len(remaining) == len(block.Hosts) {
		return apperr.New(apperr.ErrDetachServer, oldAlias)
	}

	lines := append([]string(nil), config.Lines...)
	lines[block.HostLine] = replaceHostLine(lines[block.HostLine], remaining)
	insert := renderUpdatedBlock(config.Lines, block, updated)
	withInsert := make([]string, 0, len(lines)+len(insert)+1)
	withInsert = append(withInsert, lines[:block.Start]...)
	withInsert = append(withInsert, insert...)
	withInsert = append(withInsert, "")
	withInsert = append(withInsert, lines[block.Start:]...)
	config.Lines = withInsert
	if err := config.write(); err != nil {
		config.Lines = saved
		return err
	}
	return nil
}

func DeleteServer(config *Config, alias string) error {
	blockIndex := findOriginBlock(config, alias)
	if blockIndex < 0 {
		return apperr.New(apperr.ErrServerNotFound, alias, config.Path)
	}
	saved := append([]string(nil), config.Lines...)
	block := config.Blocks[blockIndex]
	if len(block.Hosts) == 1 && block.Hosts[0] == alias {
		lines := make([]string, 0, len(config.Lines)-(block.End-block.Start))
		lines = append(lines, config.Lines[:block.Start]...)
		lines = append(lines, config.Lines[block.End:]...)
		for len(lines) > 1 && strings.TrimSpace(lines[len(lines)-1]) == "" && strings.TrimSpace(lines[len(lines)-2]) == "" {
			lines = lines[:len(lines)-1]
		}
		config.Lines = lines
		if err := config.write(); err != nil {
			config.Lines = saved
			return err
		}
		return nil
	}

	remaining := make([]string, 0, len(block.Hosts)-1)
	for _, host := range block.Hosts {
		if host != alias {
			remaining = append(remaining, host)
		}
	}
	if len(remaining) == len(block.Hosts) {
		return apperr.New(apperr.ErrAliasNotInHost, alias)
	}
	config.Lines[block.HostLine] = replaceHostLine(config.Lines[block.HostLine], remaining)
	if err := config.write(); err != nil {
		config.Lines = saved
		return err
	}
	return nil
}

func isConcreteAlias(host string) bool {
	return host != "" && !strings.HasPrefix(host, "!") && !strings.ContainsAny(host, "*?[")
}

func hostBlockMatches(alias string, patterns []string) bool {
	matched := false
	for _, pattern := range patterns {
		negated := strings.HasPrefix(pattern, "!")
		if negated {
			pattern = strings.TrimPrefix(pattern, "!")
		}
		if wildcardMatch(pattern, alias) {
			if negated {
				return false
			}
			matched = true
		}
	}
	return matched
}

func wildcardMatch(pattern, value string) bool {
	matched, err := filepath.Match(strings.ToLower(pattern), strings.ToLower(value))
	return err == nil && matched
}

func findOriginBlock(config *Config, alias string) int {
	for index, block := range config.Blocks {
		for _, host := range block.Hosts {
			if host == alias {
				return index
			}
		}
	}
	return -1
}

func rewriteSingleHostBlock(lines []string, block Block, server domain.Server) []string {
	body := renderUpdatedBlock(lines, block, server)
	out := make([]string, 0, len(lines)-(block.End-block.Start)+len(body))
	out = append(out, lines[:block.Start]...)
	out = append(out, body...)
	out = append(out, lines[block.End:]...)
	return out
}

func renderUpdatedBlock(lines []string, block Block, server domain.Server) []string {
	desired := map[string]string{
		"hostname":     server.Host,
		"user":         server.User,
		"port":         strconv.Itoa(server.Port),
		"identityfile": server.IdentityFile,
	}
	canonical := map[string]string{
		"hostname":     "HostName",
		"user":         "User",
		"port":         "Port",
		"identityfile": "IdentityFile",
	}
	seen := map[string]bool{}
	body := make([]string, 0, block.End-block.Start+4)
	body = append(body, replaceHostLine(lines[block.HostLine], []string{server.Alias}))

	for index := block.HostLine + 1; index < block.End; index++ {
		line := lines[index]
		key, _, ok, _ := parseDirective(line)
		if !ok {
			body = append(body, line)
			continue
		}
		value, supported := desired[key]
		if !supported {
			body = append(body, line)
			continue
		}
		if seen[key] {
			if key == "identityfile" {
				body = append(body, line)
			}
			continue
		}
		seen[key] = true
		if value != "" {
			body = append(body, replaceDirectiveLine(line, canonical[key], value))
		}
	}

	missing := make([]string, 0, 4)
	for _, key := range []string{"hostname", "user", "port", "identityfile"} {
		if !seen[key] && desired[key] != "" {
			missing = append(missing, "  "+canonical[key]+" "+quoteSSH(desired[key]))
		}
	}
	if len(missing) > 0 {
		insertAt := len(body)
		for insertAt > 1 {
			trimmed := strings.TrimSpace(body[insertAt-1])
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				insertAt--
				continue
			}
			break
		}
		body = append(body[:insertAt], append(missing, body[insertAt:]...)...)
	}

	return body
}

func renderServerBlock(server domain.Server) []string {
	lines := []string{"Host " + server.Alias, "  HostName " + quoteSSH(server.Host)}
	if server.User != "" {
		lines = append(lines, "  User "+quoteSSH(server.User))
	}
	lines = append(lines, "  Port "+strconv.Itoa(server.Port))
	if server.IdentityFile != "" {
		lines = append(lines, "  IdentityFile "+quoteSSH(server.IdentityFile))
	}
	return lines
}

func replaceHostLine(line string, hosts []string) string {
	indent, keyword, comment := lineParts(line, "Host")
	out := indent + keyword + " " + strings.Join(hosts, " ")
	if comment != "" {
		out += " " + comment
	}
	return out
}

func replaceDirectiveLine(line, fallbackKey, value string) string {
	indent, keyword, comment := lineParts(line, fallbackKey)
	out := indent + keyword + " " + quoteSSH(value)
	if comment != "" {
		out += " " + comment
	}
	return out
}

func lineParts(line, fallbackKey string) (indent, keyword, comment string) {
	trimmedLeft := strings.TrimLeftFunc(line, unicode.IsSpace)
	indent = line[:len(line)-len(trimmedLeft)]
	keyword = fallbackKey
	for index, current := range trimmedLeft {
		if unicode.IsSpace(current) || current == '=' {
			if index > 0 {
				keyword = trimmedLeft[:index]
			}
			break
		}
	}
	if commentAt := unquotedCommentIndex(line); commentAt >= 0 {
		comment = strings.TrimSpace(line[commentAt:])
	}
	return indent, keyword, comment
}

func unquotedCommentIndex(value string) int {
	var quote rune
	escaped := false
	for index, current := range value {
		if escaped {
			escaped = false
			continue
		}
		if current == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if current == quote {
				quote = 0
			}
			continue
		}
		if current == '\'' || current == '"' {
			quote = current
			continue
		}
		if current == '#' {
			return index
		}
	}
	return -1
}

func quoteSSH(value string) string {
	if value == "" {
		return "\"\""
	}
	if !strings.ContainsAny(value, " \t\r\n#\"'") {
		return value
	}
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}
