package domain

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxScriptRunes = 1500

var (
	ErrInvalidScript = errors.New("script is empty or too long")
	ErrUnsafeScript  = errors.New("script was blocked by content rules")
)

var promptTagReplacer = strings.NewReplacer("<", " ", ">", " ")

func SanitizeUntrusted(text string, maxRunes int) string {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
	cleaned = promptTagReplacer.Replace(cleaned)
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	if utf8.RuneCountInString(cleaned) <= maxRunes {
		return cleaned
	}
	return string([]rune(cleaned)[:maxRunes])
}

func ValidateScript(script string) error {
	length := utf8.RuneCountInString(script)
	switch {
	case length == 0:
		return ErrInvalidScript
	case length > maxScriptRunes:
		return ErrInvalidScript
	}
	return nil
}
