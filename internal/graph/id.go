package graph

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// NormalizeID produces a deterministic slug per spec §6:
// NFC normalize, casefold (lowercase), collapse every run of non-word
// characters to a single '-', and trim leading/trailing '-'.
func NormalizeID(s string) string {
	s = norm.NFC.String(s)
	s = strings.ToLower(s)

	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		if isWordRune(r) {
			b.WriteRune(r)
			prevDash = false
		} else {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// isWordRune reports whether r is a letter or digit (kept verbatim) for ID purposes.
func isWordRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r > 0x7F:
		return unicodeIsLetterOrDigit(r)
	default:
		return false
	}
}

// NodeID builds a deterministic, file-qualified node id: "<file>:<label>", both
// normalized. File qualification prevents collisions between identically named
// symbols in different files.
func NodeID(file, label string) string {
	return NormalizeID(file) + ":" + NormalizeID(label)
}
