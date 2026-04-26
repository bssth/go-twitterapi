package console

import "unicode/utf8"

func OneLine(s string, maxRunes ...int) string {
	limit := 100
	if len(maxRunes) > 0 && maxRunes[0] > 0 {
		limit = maxRunes[0]
	}

	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch r {
		case '\n', '\r', '\t':
			r = ' '
		}
		out = append(out, r)
	}
	if len(out) > limit {
		out = append(out[:limit], '…')
	}
	return string(out)
}

func RuneLen(s string) int { return utf8.RuneCountInString(s) }
