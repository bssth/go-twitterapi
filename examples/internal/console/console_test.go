package console

import "testing"

func TestOneLine(t *testing.T) {
	cases := []struct {
		in    string
		limit int
		want  string
	}{
		{"plain", 0, "plain"},
		{"a\nb\rc\td", 0, "a b c d"},
		{"long string that exceeds the limit", 5, "long …"},
		{"short", 100, "short"},
	}
	for _, c := range cases {
		var got string
		if c.limit == 0 {
			got = OneLine(c.in)
		} else {
			got = OneLine(c.in, c.limit)
		}
		if got != c.want {
			t.Errorf("OneLine(%q, %d) = %q; want %q", c.in, c.limit, got, c.want)
		}
	}
}

func TestRuneLen(t *testing.T) {
	if RuneLen("hello") != 5 {
		t.Fatal()
	}
	if RuneLen("привет") != 6 {
		t.Fatalf("got %d", RuneLen("привет"))
	}
}
