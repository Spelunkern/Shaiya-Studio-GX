// Package codepage converts between Windows code pages and Go strings, the
// role WideCharToMultiByte/MultiByteToWideChar play in shStudio.
package codepage

import (
	"strings"
	"sync"
	"unicode/utf16"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// Special code page values used by shStudio.
const (
	ACP     = 0    // CP_ACP: the system ANSI code page
	OEMCP   = 1    // CP_OEMCP: the system OEM code page
	UTF16LE = 1200 // strings stored as UTF-16LE
	UTF8    = 65001
)

var (
	mu       sync.RWMutex
	ansiCP   = 1252
	oemCP    = 437
	encoders = map[int]encoding.Encoding{
		437:   charmap.CodePage437,
		850:   charmap.CodePage850,
		874:   charmap.Windows874,
		932:   japanese.ShiftJIS,
		936:   simplifiedchinese.GBK,
		949:   korean.EUCKR,
		950:   traditionalchinese.Big5,
		1250:  charmap.Windows1250,
		1251:  charmap.Windows1251,
		1252:  charmap.Windows1252,
		1253:  charmap.Windows1253,
		1254:  charmap.Windows1254,
		1255:  charmap.Windows1255,
		1256:  charmap.Windows1256,
		1257:  charmap.Windows1257,
		1258:  charmap.Windows1258,
		28591: charmap.ISO8859_1,
		28592: charmap.ISO8859_2,
		28593: charmap.ISO8859_3,
		28594: charmap.ISO8859_4,
		28595: charmap.ISO8859_5,
		28597: charmap.ISO8859_7,
		28598: charmap.ISO8859_8,
		28599: charmap.ISO8859_9,
		28605: charmap.ISO8859_15,
		1361:  johab{},
	}
)

// SetSystem sets the code pages that ACP and OEMCP resolve to. On Windows
// the application calls it with GetACP/GetOEMCP; elsewhere the defaults are
// 1252 and 437.
func SetSystem(ansi, oem int) {
	mu.Lock()
	ansiCP, oemCP = ansi, oem
	mu.Unlock()
}

// Resolve maps ACP/OEMCP to the concrete code page.
func Resolve(cp int) int {
	mu.RLock()
	defer mu.RUnlock()
	switch cp {
	case ACP:
		return ansiCP
	case OEMCP:
		return oemCP
	}
	return cp
}

func lookup(cp int) encoding.Encoding {
	cp = Resolve(cp)
	if e, ok := encoders[cp]; ok {
		return e
	}
	return charmap.Windows1252
}

// Decode converts bytes in code page cp to a string.
func Decode(b []byte, cp int) string {
	switch Resolve(cp) {
	case UTF8:
		return string(b)
	case UTF16LE:
		u := make([]uint16, len(b)/2)
		for i := range u {
			u[i] = uint16(b[2*i]) | uint16(b[2*i+1])<<8
		}
		return string(utf16.Decode(u))
	}
	out, err := lookup(cp).NewDecoder().Bytes(b)
	if err != nil {
		return string(b)
	}
	return string(out)
}

// Encode converts s to code page cp; characters that cannot be represented
// become '?', like WideCharToMultiByte with a default char.
func Encode(s string, cp int) []byte {
	switch Resolve(cp) {
	case UTF8:
		return []byte(s)
	case UTF16LE:
		u := utf16.Encode([]rune(s))
		b := make([]byte, 2*len(u))
		for i, v := range u {
			b[2*i], b[2*i+1] = byte(v), byte(v>>8)
		}
		return b
	}
	enc := encoding.ReplaceUnsupported(lookup(cp).NewEncoder())
	out, err := enc.Bytes([]byte(s))
	if err != nil {
		return []byte(strings.Map(func(r rune) rune {
			if r < 0x80 {
				return r
			}
			return '?'
		}, s))
	}
	return out
}

// Known reports whether cp is supported.
func Known(cp int) bool {
	switch Resolve(cp) {
	case UTF8, UTF16LE:
		return true
	}
	_, ok := encoders[Resolve(cp)]
	return ok
}
