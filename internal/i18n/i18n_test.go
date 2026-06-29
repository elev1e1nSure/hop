package i18n

import (
	"os"
	"testing"

	"sshm/internal/apperr"
)

func TestCatalogsContainSameKeys(t *testing.T) {
	for key := range english {
		if _, ok := russian[key]; !ok {
			t.Errorf("Russian catalog is missing %q", key)
		}
	}
	for key := range russian {
		if _, ok := english[key]; !ok {
			t.Errorf("English catalog is missing %q", key)
		}
	}
}

func TestRussianErrorIsLocalized(t *testing.T) {
	translator := New(Russian)
	got := translator.Error(apperr.New(apperr.ErrSSHUnavailable))
	want := russian[apperr.ErrSSHUnavailable]
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestPermissionErrorIsLocalizedWithoutSystemLanguage(t *testing.T) {
	translator := New(Russian)
	err := apperr.Wrap(apperr.ErrReadSSHConfig, os.ErrPermission, "/tmp/config")
	got := translator.Error(err)
	want := "Недостаточно прав для доступа к \"/tmp/config\"."
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
