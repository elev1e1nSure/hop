package util

import (
	"strings"
	"unicode/utf8"
)

func FuzzyMatch(query, target string) bool {
	if query == "" {
		return true
	}
	qi := 0
	qr := []rune(query)
	for _, r := range target {
		if qi < len(qr) && r == qr[qi] {
			qi++
			if qi == len(qr) {
				return true
			}
		}
	}
	return false
}

func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:width-1]) + "…"
}

func PadRight(s string, width int) string {
	missing := width - utf8.RuneCountInString(s)
	if missing <= 0 {
		return s
	}
	return s + strings.Repeat(" ", missing)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
