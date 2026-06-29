package cli

import (
	"strings"

	"hop/internal/apperr"
	"hop/internal/i18n"
)

type Options struct {
	Language i18n.Language
}

func Parse(args []string, getenv func(string) string) (Options, error) {
	options := Options{Language: i18n.Detect(getenv)}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--language":
			if i+1 >= len(args) {
				return options, apperr.New(apperr.ErrMissingLanguage)
			}
			i++
			language, ok := i18n.ParseLanguage(args[i])
			if !ok {
				return options, apperr.New(apperr.ErrInvalidLanguage, args[i])
			}
			options.Language = language
		case strings.HasPrefix(arg, "--language="):
			value := strings.TrimPrefix(arg, "--language=")
			if value == "" {
				return options, apperr.New(apperr.ErrMissingLanguage)
			}
			language, ok := i18n.ParseLanguage(value)
			if !ok {
				return options, apperr.New(apperr.ErrInvalidLanguage, value)
			}
			options.Language = language
		default:
			// The original binary ignored unrelated arguments. Preserve that contract.
			continue
		}
	}
	return options, nil
}
