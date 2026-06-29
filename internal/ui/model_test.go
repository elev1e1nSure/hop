package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateBrowseFiltersServers(t *testing.T) {
	model := newTestModel(t, "Host prod\n  HostName prod.example.com\nHost stage\n  HostName stage.example.com\n")

	for _, char := range "prod" {
		model = updateTestModel(t, model, runeKey(char))
	}

	if got := model.filter.Value(); got != "prod" {
		t.Fatalf("filter value = %q, want %q", got, "prod")
	}
	items := model.list.Items()
	if len(items) != 1 {
		t.Fatalf("filtered item count = %d, want 1", len(items))
	}
	row := items[0].(rowItem)
	if row.Server.Alias != "prod" {
		t.Fatalf("filtered alias = %q, want prod", row.Server.Alias)
	}
}

func TestUpdateStartsAddForm(t *testing.T) {
	model := newTestModel(t, "Host prod\n  HostName prod.example.com\n")

	model = updateTestModel(t, model, key(tea.KeyCtrlN))

	if model.mode != modeForm || model.editing {
		t.Fatalf("mode/editing = %v/%v, want add form", model.mode, model.editing)
	}
	if len(model.form) != 5 || model.form[3].Value() != "22" {
		t.Fatalf("form was not initialized for add: %#v", model.form)
	}
}

func TestUpdateStartsEditForm(t *testing.T) {
	model := newTestModel(t, "Host prod\n  HostName prod.example.com\n  User deploy\n  Port 2200\n")

	model = updateTestModel(t, model, key(tea.KeyCtrlE))

	if model.mode != modeForm || !model.editing || model.editAlias != "prod" {
		t.Fatalf("mode/editing/editAlias = %v/%v/%q, want edit prod", model.mode, model.editing, model.editAlias)
	}
	if model.form[0].Value() != "prod" || model.form[1].Value() != "prod.example.com" || model.form[3].Value() != "2200" {
		t.Fatalf("form values = %q/%q/%q", model.form[0].Value(), model.form[1].Value(), model.form[3].Value())
	}
}

func TestUpdateDeleteConfirmCanCancel(t *testing.T) {
	model := newTestModel(t, "Host prod\n  HostName prod.example.com\n")

	model = updateTestModel(t, model, key(tea.KeyCtrlD))
	if model.mode != modeConfirmDelete || model.confirmFor != "prod" {
		t.Fatalf("mode/confirmFor = %v/%q, want confirm prod", model.mode, model.confirmFor)
	}

	model = updateTestModel(t, model, key(tea.KeyEsc))
	if model.mode != modeBrowse || model.confirmFor != "" {
		t.Fatalf("mode/confirmFor = %v/%q, want browse with empty confirm", model.mode, model.confirmFor)
	}
}

func TestUpdateDeleteConfirmRemovesServer(t *testing.T) {
	model := newTestModel(t, "Host doomed\n  HostName doomed.example.com\nHost keep\n  HostName keep.example.com\n")

	model = updateTestModel(t, model, key(tea.KeyCtrlD))
	model = updateTestModel(t, model, runeKey('y'))

	if model.mode != modeBrowse || model.errorText != "" {
		t.Fatalf("mode/error = %v/%q, want browse without error", model.mode, model.errorText)
	}
	if len(model.servers) != 1 || model.servers[0].Alias != "keep" {
		t.Fatalf("servers = %#v, want only keep", model.servers)
	}
	data, err := os.ReadFile(model.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "doomed") {
		t.Fatalf("config still contains deleted server: %q", data)
	}
}

func newTestModel(t *testing.T, config string) Model {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	model, err := newModel(configPath, filepath.Join(dir, "history.json"), en())
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func updateTestModel(t *testing.T, model Model, message tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(message)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want ui.Model", updated)
	}
	return next
}

func key(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: keyType}
}

func runeKey(char rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
}
