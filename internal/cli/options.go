package cli

import (
	"strings"

	"hop/internal/apperr"
	"hop/internal/i18n"
)

type Options struct {
	Language i18n.Language
	Path     PathAction
	Help     bool
}

type PathAction string

const (
	PathActionNone   PathAction = ""
	PathActionAdd    PathAction = "add"
	PathActionRemove PathAction = "remove"
)

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
		case arg == "--path":
			if i+1 >= len(args) {
				return options, apperr.New(apperr.ErrMissingPathAction)
			}
			i++
			action, ok := parsePathAction(args[i])
			if !ok {
				return options, apperr.New(apperr.ErrInvalidPathAction, args[i])
			}
			options.Path = action
		case strings.HasPrefix(arg, "--path="):
			value := strings.TrimPrefix(arg, "--path=")
			if value == "" {
				return options, apperr.New(apperr.ErrMissingPathAction)
			}
			action, ok := parsePathAction(value)
			if !ok {
				return options, apperr.New(apperr.ErrInvalidPathAction, value)
			}
			options.Path = action
		case arg == "--help", arg == "-h":
			options.Help = true
		default:
			// The original binary ignored unrelated arguments. Preserve that contract.
			continue
		}
	}
	return options, nil
}

func parsePathAction(value string) (PathAction, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PathActionAdd):
		return PathActionAdd, true
	case string(PathActionRemove):
		return PathActionRemove, true
	default:
		return PathActionNone, false
	}
}
