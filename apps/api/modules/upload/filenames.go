package upload

import "strings"

// SanitizeFilenamePart collapses any run of non-alphanumeric characters into a
// single dash so a name is safe to embed in a download filename. Empty or
// entirely-punctuation input falls back to fallback so the caller never builds
// a filename with a missing segment.
func SanitizeFilenamePart(s, fallback string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	part := strings.Trim(b.String(), "-")
	if part == "" {
		return fallback
	}
	return part
}
