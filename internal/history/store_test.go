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

func TestLoadWrapsInaccessiblePath(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(filepath.Join(parent, "history.json"))
	if err == nil {
		t.Fatal("Load() error = nil, want path error")
	}
}
