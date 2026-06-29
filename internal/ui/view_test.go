package ui

import (
	"strings"
	"testing"
	"time"

	"hop/internal/domain"
	"hop/internal/i18n"
)

func en() i18n.Translator { return i18n.New(i18n.English) }
func ru() i18n.Translator { return i18n.New(i18n.Russian) }

func serverFixture() domain.Server {
	return domain.Server{
		Alias:        "test-server",
		Host:         "example.com",
		User:         "deploy",
		Port:         22,
		IdentityFile: "~/.ssh/id_ed25519",
	}
}

func TestShowDetailsPanelWide(t *testing.T) {
	model := Model{width: detailsPanelMinWidth}
	if !model.showDetailsPanel() {
		t.Fatal("showDetailsPanel() = false, want true at min width")
	}
}

func TestShowDetailsPanelNarrow(t *testing.T) {
	model := Model{width: detailsPanelMinWidth - 1}
	if model.showDetailsPanel() {
		t.Fatal("showDetailsPanel() = true, want false below min width")
	}
}

func TestOptionalValueEmpty(t *testing.T) {
	got := optionalValue(en(), "")
	if got != "none" {
		t.Fatalf("optionalValue(empty) = %q, want %q", got, "none")
	}
	got = optionalValue(en(), "  ")
	if got != "none" {
		t.Fatalf("optionalValue(whitespace) = %q, want %q", got, "none")
	}
}

func TestOptionalValuePresent(t *testing.T) {
	got := optionalValue(en(), "deploy")
	if got != "deploy" {
		t.Fatalf("optionalValue(present) = %q, want %q", got, "deploy")
	}
}

func TestOptionalValueRussian(t *testing.T) {
	got := optionalValue(ru(), "")
	if got != "нет" {
		t.Fatalf("optionalValue(empty, ru) = %q, want %q", got, "нет")
	}
}

func TestProxyValueWithProxy(t *testing.T) {
	server := domain.Server{HasProxy: true}
	got := proxyValue(en(), server)
	if got != "via proxy" {
		t.Fatalf("proxyValue(proxy) = %q, want %q", got, "via proxy")
	}
}

func TestProxyValueWithoutProxy(t *testing.T) {
	got := proxyValue(en(), domain.Server{})
	if got != "none" {
		t.Fatalf("proxyValue(no-proxy) = %q, want %q", got, "none")
	}
}

func TestProxyValueRussian(t *testing.T) {
	got := proxyValue(ru(), domain.Server{HasProxy: true})
	if got != "через прокси" {
		t.Fatalf("proxyValue(proxy, ru) = %q, want %q", got, "через прокси")
	}
}

func TestStatusValueUnchecked(t *testing.T) {
	server := domain.Server{}
	got := statusValue(en(), server)
	if got != "checking" {
		t.Fatalf("statusValue(unchecked) = %q, want %q", got, "checking")
	}
}

func TestStatusValueOnline(t *testing.T) {
	server := domain.Server{Checked: true, Online: true}
	got := statusValue(en(), server)
	if got != "online" {
		t.Fatalf("statusValue(online) = %q, want %q", got, "online")
	}
}

func TestStatusValueOffline(t *testing.T) {
	server := domain.Server{Checked: true, Online: false}
	got := statusValue(en(), server)
	if got != "offline" {
		t.Fatalf("statusValue(offline) = %q, want %q", got, "offline")
	}
}

func TestStatusValueProxy(t *testing.T) {
	server := domain.Server{HasProxy: true}
	got := statusValue(en(), server)
	if got != "via proxy" {
		t.Fatalf("statusValue(proxy) = %q, want %q", got, "via proxy")
	}
}

func TestStatusValueRussian(t *testing.T) {
	server := domain.Server{Checked: true, Online: true}
	got := statusValue(ru(), server)
	if got != "доступен" {
		t.Fatalf("statusValue(online, ru) = %q, want %q", got, "доступен")
	}
}

func TestLastUsedNever(t *testing.T) {
	server := domain.Server{}
	got := lastUsedValue(en(), server)
	if got != "never" {
		t.Fatalf("lastUsedValue(zero) = %q, want %q", got, "never")
	}
}

func TestLastUsedDate(t *testing.T) {
	server := domain.Server{LastUsed: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)}
	got := lastUsedValue(en(), server)
	if got != "2025-06-15" {
		t.Fatalf("lastUsedValue(2025-06-15) = %q, want %q", got, "2025-06-15")
	}
}

func TestLastUsedRussian(t *testing.T) {
	got := lastUsedValue(ru(), domain.Server{})
	if got != "никогда" {
		t.Fatalf("lastUsedValue(zero, ru) = %q, want %q", got, "никогда")
	}
}

func TestCompactDetailsViewContainsRequiredFields(t *testing.T) {
	model := Model{translator: en(), width: 120}
	got := model.compactDetailsView(serverFixture())
	for _, want := range []string{"Host", "User", "Port", "Uses"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compactDetailsView() missing %q in %q", want, got)
		}
	}
}

func TestDetailsPanelViewContainsRequiredFields(t *testing.T) {
	model := Model{translator: en()}
	server := serverFixture()
	server.HasProxy = true
	got := model.detailsPanelView(server)
	for _, want := range []string{"test-server", "example.com", "deploy", "Status", "via proxy", "Last used"} {
		if !strings.Contains(got, want) {
			t.Fatalf("detailsPanelView() missing %q in %q", want, got)
		}
	}
}

func TestDetailsPanelViewEmptyFieldsShownAsNone(t *testing.T) {
	model := Model{translator: en()}
	server := domain.Server{
		Alias: "min",
		Host:  "1.2.3.4",
		Port:  22,
	}
	got := model.detailsPanelView(server)
	if !strings.Contains(got, "none") {
		t.Fatalf("detailsPanelView() missing 'none' for empty fields: %q", got)
	}
}

func TestSanitizeStripsControlCharacters(t *testing.T) {
	dirty := "hello\x00world\x1Btest\x7Fend"
	model := Model{translator: en()}
	server := domain.Server{
		Alias:        dirty,
		Host:         dirty,
		User:         dirty,
		IdentityFile: dirty,
		Port:         22,
	}
	got := model.detailsPanelView(server)
	for _, forbidden := range []string{"\x00", "\x1B", "\x7F"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("detailsPanelView() contains control char %q in %q", forbidden, got)
		}
	}
}

func TestDetailsRowFormatsLabelAndValue(t *testing.T) {
	model := Model{translator: en()}
	got := model.detailsRow("testKey", "testValue", 8, 20)
	if !strings.Contains(got, "testKey") || !strings.Contains(got, "testValue") {
		t.Fatalf("detailsRow() = %q, want label and value present", got)
	}
}
