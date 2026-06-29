package sshconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hop/internal/apperr"
	"hop/internal/domain"
)

func TestLoadMissingConfigReturnsEmptyConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ssh", "config")
	config, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(config.Lines) != 0 || len(config.Blocks) != 0 {
		t.Fatalf("config is not empty: %#v", config)
	}
}

func TestLoadRejectsUnclosedQuote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host broken\n  HostName \"example.com\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Key != apperr.ErrUnclosedQuote {
		t.Fatalf("error = %v, want %s", err, apperr.ErrUnclosedQuote)
	}
}

func TestResolveRejectsInvalidPort(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host broken\n  HostName example.com\n  Port nope\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolveServers(config, nil)
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Key != apperr.ErrInvalidPortConfig {
		t.Fatalf("error = %v, want %s", err, apperr.ErrInvalidPortConfig)
	}
}

func TestEditPreservesMissingTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host old\n  HostName old.example.com\n  Port 22"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := domain.Server{Alias: "new", Host: "new.example.com", Port: 2222}
	if err := EditServer(config, "old", updated); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(string(data), "\n") {
		t.Fatalf("file unexpectedly gained a trailing newline: %q", data)
	}
}

func TestAddToEmptyConfigCreatesValidBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ssh", "config")
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	server := domain.Server{Alias: "prod", Host: "prod.example.com", User: "deploy", Port: 22}
	if err := AddServer(config, server); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "Host prod\n  HostName prod.example.com\n  User deploy\n  Port 22\n" {
		t.Fatalf("content = %q", got)
	}
}

func TestEqualsSyntaxIsParsed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host = prod\n  HostName=prod.example.com\n  Port = 2200\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	servers, err := ResolveServers(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || servers[0].Alias != "prod" || servers[0].Port != 2200 {
		t.Fatalf("servers = %#v", servers)
	}
}

func TestLoadWrapsInaccessiblePath(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(filepath.Join(parent, "config"))
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Key != apperr.ErrReadSSHConfig {
		t.Fatalf("error = %v, want %s", err, apperr.ErrReadSSHConfig)
	}
}

func TestEqualsWithoutWhitespaceIsParsed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host =prod\n  HostName =prod.example.com\n  Port =2201\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	servers, err := ResolveServers(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || servers[0].Alias != "prod" || servers[0].Host != "prod.example.com" || servers[0].Port != 2201 {
		t.Fatalf("servers = %#v", servers)
	}
}
