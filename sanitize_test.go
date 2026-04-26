package twitterapi

import "testing"

func TestSanitizeForTwitter(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"\u2014em\u2013en\u2011nbh\u2212minus", "-em-en-nbh-minus"},
		{"a\u00A0b\u2007c\u202Fd", "a b c d"},
		{"\u201Chi\u201D \u2018x\u2019", "\"hi\" 'x'"},
		{"a\u200Bb\u200Cc\u200Dd\u2060e\uFEFFf", "abcdef"},
		{"  trim  ", "trim"},
		{"line1\nline2", "line1\nline2"},
		{"\x00ctrl\x07rune", "ctrlrune"},
	}
	for _, c := range cases {
		got := SanitizeForTwitter(c.in)
		if got != c.want {
			t.Errorf("SanitizeForTwitter(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}
