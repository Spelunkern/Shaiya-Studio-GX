package codepage

import "testing"

func TestRoundTrip(t *testing.T) {
	for _, c := range []struct {
		cp int
		s  string
	}{{949, "샤이야 data"}, {1361, "한글 abc"}, {1252, "Ångström"}, {1200, "UTF16 ü"}, {932, "日本"}} {
		b := Encode(c.s, c.cp)
		if got := Decode(b, c.cp); got != c.s {
			t.Errorf("cp %d: %q -> % x -> %q", c.cp, c.s, b, got)
		}
	}
	// Johab "가" is 0x8861.
	if b := Encode("가", 1361); len(b) != 2 || b[0] != 0x88 || b[1] != 0x61 {
		t.Errorf("johab 가 = % x", b)
	}
}
