package codepage

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// johab implements code page 1361 (Korean Johab) for ASCII and the 11,172
// modern Hangul syllables, which are encoded algorithmically. Johab symbols
// and Hanja are not supported and decode/encode as '?'.
type johab struct{}

var (
	johabCho  = [32]int{-1, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	johabJung = [32]int{-1, -1, -1, 0, 1, 2, 3, 4, -1, -1, 5, 6, 7, 8, 9, 10, -1, -1, 11, 12, 13, 14, 15, 16, -1, -1, 17, 18, 19, 20, -1, -1}
	johabJong = [32]int{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, -1, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, -1, -1}
)

func (johab) NewDecoder() *encoding.Decoder { return &encoding.Decoder{Transformer: johabDec{}} }
func (johab) NewEncoder() *encoding.Encoder { return &encoding.Encoder{Transformer: johabEnc{}} }

type johabDec struct{ transform.NopResetter }

func (johabDec) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		b := src[nSrc]
		if b < 0x80 {
			if nDst >= len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			dst[nDst] = b
			nDst++
			nSrc++
			continue
		}
		if nSrc+1 >= len(src) {
			if !atEOF {
				return nDst, nSrc, transform.ErrShortSrc
			}
			if nDst >= len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			dst[nDst] = '?'
			return nDst + 1, nSrc + 1, nil
		}
		code := int(b)<<8 | int(src[nSrc+1])
		r := '?'
		if b >= 0x84 && b <= 0xD3 {
			cho, jung, jong := johabCho[(code>>10)&31], johabJung[(code>>5)&31], johabJong[code&31]
			if cho >= 0 && jung >= 0 && jong >= 0 {
				r = rune(0xAC00 + (cho*21+jung)*28 + jong)
			}
		}
		if nDst+utf8.RuneLen(r) > len(dst) {
			return nDst, nSrc, transform.ErrShortDst
		}
		nDst += utf8.EncodeRune(dst[nDst:], r)
		nSrc += 2
	}
	return nDst, nSrc, nil
}

type johabEnc struct{ transform.NopResetter }

func index(table *[32]int, v int) int {
	for i, x := range table {
		if x == v {
			return i
		}
	}
	return 0
}

func (johabEnc) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		r, size := utf8.DecodeRune(src[nSrc:])
		if r == utf8.RuneError && size == 1 && !atEOF && !utf8.FullRune(src[nSrc:]) {
			return nDst, nSrc, transform.ErrShortSrc
		}
		switch {
		case r < 0x80:
			if nDst >= len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			dst[nDst] = byte(r)
			nDst++
		case r >= 0xAC00 && r <= 0xD7A3:
			if nDst+2 > len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			s := int(r - 0xAC00)
			code := 0x8000 | index(&johabCho, s/(21*28))<<10 | index(&johabJung, s/28%21)<<5 | index(&johabJong, s%28)
			dst[nDst], dst[nDst+1] = byte(code>>8), byte(code)
			nDst += 2
		default:
			if nDst >= len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			dst[nDst] = '?'
			nDst++
		}
		nSrc += size
	}
	return nDst, nSrc, nil
}
