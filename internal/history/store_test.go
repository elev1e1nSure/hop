package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmptyHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	records, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("records length = %d, want 0", len(records))
	}
}

func TestLoadOnDirectory(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil {
		t.Fatal("Load() error = nil, want read error")
	}
}

func TestLoadMissingFile(t *testing.T) {
	records, err := Load(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("records length = %d, want 0", len(records))
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
}
