package services

import (
	"regexp"
	"strings"

	cc "fairchild_be/internal/constants"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// NormalizeMobileNumber converts common Philippine local formats
// (e.g. "09171234567", "9171234567", "639171234567") into E.164
// ("+639171234567"), which is what Movider requires. A number already in
// E.164 form is returned unchanged. Anything else is rejected rather than
// guessed at.
func NormalizeMobileNumber(raw string) (string, error) {
	digits := stripNonDigits(raw)

	switch {
	case strings.HasPrefix(strings.TrimSpace(raw), "+"):
		candidate := "+" + digits
		if !e164Pattern.MatchString(candidate) {
			return "", cc.ErrInvalidMobileNumber
		}
		return candidate, nil
	case strings.HasPrefix(digits, "63") && len(digits) == 12:
		return "+" + digits, nil
	case strings.HasPrefix(digits, "0") && len(digits) == 11:
		return "+63" + digits[1:], nil
	case len(digits) == 10:
		return "+63" + digits, nil
	default:
		return "", cc.ErrInvalidMobileNumber
	}
}

func stripNonDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
