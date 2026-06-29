package sshclient

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"sshm/internal/apperr"
	"sshm/internal/domain"
)

func Lookup() (string, error) {
	path, err := exec.LookPath("ssh")
	if err != nil {
		cause := fmt.Errorf("locate ssh executable: %w", err)
		return "", apperr.Wrap(apperr.ErrSSHUnavailable, cause)
	}
	return path, nil
}

func Args(server domain.Server) ([]string, error) {
	args := []string{"-p", strconv.Itoa(server.Port)}
	if server.IdentityFile != "" {
		identity, err := expandIdentity(server.IdentityFile, server)
		if err != nil {
			return nil, err
		}
		args = append(args, "-i", identity)
	}
	args = append(args, Target(server))
	return args, nil
}

func Target(server domain.Server) string {
	if server.User == "" {
		return server.Host
	}
	return server.User + "@" + server.Host
}

func Run(binary string, server domain.Server) error {
	args, err := Args(server)
	if err != nil {
		return err
	}
	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			cause := fmt.Errorf("run ssh for %s: %w", Target(server), err)
			return apperr.Wrap(apperr.ErrSSHExit, cause, Target(server), exitErr.ExitCode())
		}
		cause := fmt.Errorf("start ssh for %s: %w", Target(server), err)
		return apperr.Wrap(apperr.ErrSSHStart, cause, Target(server))
	}
	return nil
}

func expandIdentity(path string, server domain.Server) (string, error) {
	needsHome := path == "~" || strings.HasPrefix(path, "~/") || strings.Contains(path, "%d")
	home := ""
	if needsHome {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			cause := fmt.Errorf("resolve home directory while expanding identity file %q: %w", path, err)
			return "", apperr.Wrap(apperr.ErrSSHHomeDir, cause)
		}
	}
	path = os.ExpandEnv(path)
	path = strings.ReplaceAll(path, "%d", home)
	path = strings.ReplaceAll(path, "%h", server.Host)
	path = strings.ReplaceAll(path, "%r", server.User)
	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
