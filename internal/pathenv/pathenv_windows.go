//go:build windows

package pathenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"

	"github.com/elev1e1nSure/hop/internal/apperr"
)

const pathValueName = "Path"

func Apply(action Action) (Result, error) {
	directory, err := executableDirectory()
	if err != nil {
		return Result{}, err
	}

	current, err := readUserPath()
	if err != nil {
		return Result{}, err
	}

	next, changed := updatePath(current, directory, action)
	if !changed {
		return Result{Directory: directory, Changed: false}, nil
	}
	if err := writeUserPath(next); err != nil {
		return Result{}, err
	}
	return Result{Directory: directory, Changed: true}, nil
}

func executableDirectory() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		cause := fmt.Errorf("locate executable: %w", err)
		return "", apperr.Wrap(apperr.ErrPathExecutable, cause)
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err == nil {
		executable = resolved
	}
	directory := filepath.Dir(executable)
	if directory == "." || directory == "" {
		cause := fmt.Errorf("resolve executable directory from %q", executable)
		return "", apperr.Wrap(apperr.ErrPathExecutable, cause)
	}
	return filepath.Clean(directory), nil
}

func readUserPath() ([]string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		cause := fmt.Errorf("open user environment registry key: %w", err)
		return nil, apperr.Wrap(apperr.ErrPathUpdate, cause)
	}
	defer key.Close()

	value, _, err := key.GetStringValue(pathValueName)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil, nil
		}
		cause := fmt.Errorf("read user PATH value: %w", err)
		return nil, apperr.Wrap(apperr.ErrPathUpdate, cause)
	}
	return splitPath(value), nil
}

func writeUserPath(parts []string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		cause := fmt.Errorf("open user environment registry key for write: %w", err)
		return apperr.Wrap(apperr.ErrPathUpdate, cause)
	}
	defer key.Close()

	value := strings.Join(parts, ";")
	if err := key.SetStringValue(pathValueName, value); err != nil {
		cause := fmt.Errorf("write user PATH value: %w", err)
		return apperr.Wrap(apperr.ErrPathUpdate, cause)
	}
	return nil
}

func updatePath(parts []string, directory string, action Action) ([]string, bool) {
	switch action {
	case ActionAdd:
		if containsPath(parts, directory) {
			return parts, false
		}
		return append([]string{directory}, parts...), true
	case ActionRemove:
		filtered := parts[:0]
		changed := false
		for _, part := range parts {
			if samePath(part, directory) {
				changed = true
				continue
			}
			filtered = append(filtered, part)
		}
		return filtered, changed
	default:
		return parts, false
	}
}

func splitPath(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ";")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return filtered
}

func containsPath(parts []string, target string) bool {
	for _, part := range parts {
		if samePath(part, target) {
			return true
		}
	}
	return false
}

func samePath(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return false
	}
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	return strings.EqualFold(left, right)
}
