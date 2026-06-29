package sshclient

import (
	"errors"
	"testing"

	"sshm/internal/apperr"
)

func TestLookupUnavailable(t *testing.T) {
	t.Setenv("PATH", "")
	_, err := Lookup()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Key != apperr.ErrSSHUnavailable {
		t.Fatalf("error = %v, want %s", err, apperr.ErrSSHUnavailable)
	}
}
