package cli

import (
	"errors"
	"testing"

	"sshm/internal/apperr"
	"sshm/internal/i18n"
)

func TestParseUsesLocaleByDefault(t *testing.T) {
	options, err := Parse(nil, func(name string) string {
		if name == "LANG" {
			return "ru_RU.UTF-8"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if options.Language != i18n.Russian {
		t.Fatalf("language = %q, want %q", options.Language, i18n.Russian)
	}
}

func TestParseLanguageOverridesLocale(t *testing.T) {
	options, err := Parse([]string{"--language=en"}, func(string) string { return "ru_RU.UTF-8" })
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if options.Language != i18n.English {
		t.Fatalf("language = %q, want %q", options.Language, i18n.English)
	}
}

func TestParseRejectsInvalidLanguage(t *testing.T) {
	_, err := Parse([]string{"--language", "de"}, func(string) string { return "en_US.UTF-8" })
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Key != apperr.ErrInvalidLanguage {
		t.Fatalf("error = %v, want %s", err, apperr.ErrInvalidLanguage)
	}
}

func TestParsePreservesUnknownArgumentBehavior(t *testing.T) {
	options, err := Parse([]string{"--legacy-argument", "value"}, func(string) string { return "en_US.UTF-8" })
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if options.Language != i18n.English {
		t.Fatalf("language = %q, want %q", options.Language, i18n.English)
	}
}
