package sshconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elev1e1nSure/hop/internal/apperr"
	"github.com/elev1e1nSure/hop/internal/domain"
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

func TestLoadOnDirectory(t *testing.T) {
	_, err := Load(t.TempDir())
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

func TestIncludeLoadsServersFromIncludedFile(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "config")
	included := filepath.Join(dir, "hosts.conf")
	if err := os.WriteFile(root, []byte("Include hosts.conf\nHost root\n  HostName root.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(included, []byte("Host included\n  HostName included.example.com\n  Port 2202\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	servers, err := ResolveServers(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 || servers[0].Alias != "included" || servers[1].Alias != "root" {
		t.Fatalf("servers = %#v", servers)
	}
	if servers[0].Host != "included.example.com" || servers[0].Port != 2202 {
		t.Fatalf("included server = %#v", servers[0])
	}
}

func TestIncludeGlobUsesLexicalOrder(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "config")
	includes := filepath.Join(dir, "config.d")
	if err := os.MkdirAll(includes, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root, []byte("Include config.d/*.conf\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(includes, "b.conf"), []byte("Host b\n  HostName b.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(includes, "a.conf"), []byte("Host a\n  HostName a.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	servers, err := ResolveServers(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 || servers[0].Alias != "a" || servers[1].Alias != "b" {
		t.Fatalf("servers = %#v", servers)
	}
}

func TestEditIncludedServerWritesIncludedFile(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "config")
	included := filepath.Join(dir, "hosts.conf")
	if err := os.WriteFile(root, []byte("Include hosts.conf\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(included, []byte("Host old\n  HostName old.example.com\n  Port 22\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	updated := domain.Server{Alias: "new", Host: "new.example.com", User: "deploy", Port: 2203}
	if err := EditServer(config, "old", updated); err != nil {
		t.Fatal(err)
	}
	rootData, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	if string(rootData) != "Include hosts.conf\n" {
		t.Fatalf("root content = %q", rootData)
	}
	includedData, err := os.ReadFile(included)
	if err != nil {
		t.Fatal(err)
	}
	want := "Host new\n  HostName new.example.com\n  Port 2203\n  User deploy\n"
	if string(includedData) != want {
		t.Fatalf("included content = %q, want %q", includedData, want)
	}
}

func TestDeleteIncludedServerWritesIncludedFile(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "config")
	included := filepath.Join(dir, "hosts.conf")
	if err := os.WriteFile(root, []byte("Include hosts.conf\nHost root\n  HostName root.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(included, []byte("Host doomed\n  HostName doomed.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteServer(config, "doomed"); err != nil {
		t.Fatal(err)
	}
	includedData, err := os.ReadFile(included)
	if err != nil {
		t.Fatal(err)
	}
	if string(includedData) != "" {
		t.Fatalf("included content = %q, want empty", includedData)
	}
	rootData, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rootData), "Host root") {
		t.Fatalf("root content was modified incorrectly: %q", rootData)
	}
}

func TestAddServerWritesRootWhenIncludesExist(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "config")
	included := filepath.Join(dir, "hosts.conf")
	if err := os.WriteFile(root, []byte("Include hosts.conf\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(included, []byte("Host included\n  HostName included.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	server := domain.Server{Alias: "added", Host: "added.example.com", Port: 22}
	if err := AddServer(config, server); err != nil {
		t.Fatal(err)
	}
	rootData, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	wantRoot := "Include hosts.conf\n\nHost added\n  HostName added.example.com\n  Port 22\n"
	if string(rootData) != wantRoot {
		t.Fatalf("root content = %q, want %q", rootData, wantRoot)
	}
	includedData, err := os.ReadFile(included)
	if err != nil {
		t.Fatal(err)
	}
	if string(includedData) != "Host included\n  HostName included.example.com\n" {
		t.Fatalf("included content = %q", includedData)
	}
}
