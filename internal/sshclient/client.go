package sshclient

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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
	return []string{Target(server)}, nil
}

func Target(server domain.Server) string {
	if server.User == "" {
		return server.Alias
	}
	return server.User + "@" + server.Alias
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

