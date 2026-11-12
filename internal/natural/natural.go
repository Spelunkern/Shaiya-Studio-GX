// Package natural implements the "logical" string order of StrCmpLogicalW,
// which shStudio uses to sort archive folders and files: comparison is
// case-insensitive and runs of digits compare by numeric value.
package natural

import "unicode"

// Compare returns -1, 0 or 1.
func Compare(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	i, j := 0, 0
	for i < len(ra) && j < len(rb) {
		ca, cb := ra[i], rb[j]
		if isDigit(ca) && isDigit(cb) {
			si := i
			for i < len(ra) && ra[i] == '0' {
				i++
			}
			sj := j
			for j < len(rb) && rb[j] == '0' {
				j++
			}
			ei, ej := i, j
			for ei < len(ra) && isDigit(ra[ei]) {
				ei++
			}
			for ej < len(rb) && isDigit(rb[ej]) {
				ej++
			}
			if ei-i != ej-j {
				return sign(ei - i - (ej - j))
			}
			for k := 0; k < ei-i; k++ {
				if ra[i+k] != rb[j+k] {
					return sign(int(ra[i+k]) - int(rb[j+k]))
				}
			}
			// Equal values: more leading zeros sort first.
			if (i - si) != (j - sj) {
				return sign((j - sj) - (i - si))
			}
			i, j = ei, ej
			continue
		}
		la, lb := unicode.ToLower(ca), unicode.ToLower(cb)
		if la != lb {
			return sign(int(la) - int(lb))
		}
		i++
		j++
	}
	return sign((len(ra) - i) - (len(rb) - j))
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}
