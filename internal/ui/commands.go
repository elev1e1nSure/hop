package ui

import (
	"net"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"hop/internal/domain"
	"hop/internal/sshclient"
	"hop/internal/util"
)

const checkTimeout = 1500 * time.Millisecond
const maxCheckConcurrency = 20

type statusCheckMsg struct {
	Alias  string
	Online bool
}

type connectReadyMsg struct {
	Server domain.Server
	Online bool
	Binary string
	Err    error
}

func (model Model) checkAllCmd() tea.Cmd {
	limit := util.Max(1, maxCheckConcurrency)
	sem := make(chan struct{}, limit)
	commands := make([]tea.Cmd, 0, len(model.servers))
	for _, server := range model.servers {
		if server.HasProxy {
			continue
		}
		commands = append(commands, checkServerCmd(server, sem))
	}
	return tea.Batch(commands...)
}

func checkServerCmd(server domain.Server, sem chan struct{}) tea.Cmd {
	return func() tea.Msg {
		sem <- struct{}{}
		defer func() { <-sem }()
		connection, err := net.DialTimeout("tcp", dialAddress(server), checkTimeout)
		if err == nil {
			_ = connection.Close()
		}
		return statusCheckMsg{Alias: server.Alias, Online: err == nil}
	}
}

func prepareConnectCmd(server domain.Server) tea.Cmd {
	return func() tea.Msg {
		started := time.Now()
		binary, err := sshclient.Lookup()
		if err == nil {
			_, err = sshclient.Args(server)
		}
		online := false
		if err == nil {
			connection, dialErr := net.DialTimeout("tcp", dialAddress(server), checkTimeout)
			if dialErr == nil {
				online = true
				_ = connection.Close()
			}
		}
		if elapsed := time.Since(started); elapsed < 350*time.Millisecond {
			time.Sleep(350*time.Millisecond - elapsed)
		}
		return connectReadyMsg{Server: server, Online: online, Binary: binary, Err: err}
	}
}

func dialAddress(server domain.Server) string {
	host := server.Host
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	}
	return net.JoinHostPort(host, strconv.Itoa(server.Port))
}
