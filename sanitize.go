package twitterapi

import (
	"strings"
	"unicode"
)

// sanitizeReplacer maps common rich-text Unicode that breaks Twitter's
// 280-char counter back to ASCII or strips it. Each entry uses an escape
// sequence so this source file contains no exotic bytes (Go rejects a
// literal BOM mid-file, and many editors corrupt zero-width chars).
var sanitizeReplacer = strings.NewReplacer(
	"\u2014", "-", // em dash
	"\u2013", "-", // en dash
	"\u2011", "-", // non-breaking hyphen
	"\u2212", "-", // minus sign
	"\u00A0", " ", // NBSP
	"\u2007", " ", // figure space
	"\u202F", " ", // narrow no-break space
	"\u201C", "\"", // left double quote
	"\u201D", "\"", // right double quote
	"\u2018", "'", // left single quote
	"\u2019", "'", // right single quote
	"\u200B", "", // zero-width space
	"\u200C", "", // zero-width non-joiner
	"\u200D", "", // zero-width joiner
	"\u2060", "", // word joiner
	"\uFEFF", "", // BOM
)

// SanitizeForTwitter normalizes punctuation and removes invisible/zero-width
// characters that often sneak in from rich-text editors and break Twitter's
// 280-char counting. It does NOT truncate; callers decide.
//
// Replacements:
//   - em/en dashes, non-breaking hyphen, minus sign  -> ASCII hyphen
//   - non-breaking / figure / narrow spaces          -> regular space
//   - smart quotes                                   -> ASCII quotes
//   - zero-width / word-joiner / BOM                 -> stripped
//
// All other Unicode control runes (except newline) are dropped.
func SanitizeForTwitter(input string) string {
	input = sanitizeReplacer.Replace(input)

	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if r == '\n' {
			b.WriteRune(r)
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
