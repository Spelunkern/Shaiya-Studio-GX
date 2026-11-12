// Package sdata implements the SData tables of the client and their
// BinarySData (DB*Data / DB*Text) counterparts, ported from shStudio
// src/data. Every table reads and writes its formats exactly like the C++
// code, so that loading and saving an untouched file gives the same bytes.
package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/natural"
)

// Format selects a serialization (the DWORD fmt of Streamable::read/write).
type Format int

const (
	// File is the client .SData format ('file').
	File Format = iota
	// Flnt is the BinarySData DB*Data format: a 128-byte header, column
	// names and rows of 64-bit cells ('flnt').
	Flnt
	// Ltxt is the BinarySData DB*Text_<LANG> format ('ltxt').
	Ltxt
)

// Text formats of toText.
const (
	TextCSV  = "csv "
	TextCSVE = "csve" // CSV including invalid rows
	TextSQL  = "sql "
)

// Tag is a four-character field identifier ('type', 'lvl ', ...), used for
// columns, sorting and generic get/set.
type Tag string

// Errors raised by the DB formats.
const errDBFormat = "Invalid or not supported DB format"

func failDB() { binstream.Fail(errDBFormat) }

// checkSignature reads a UTF-16 column name with a 1-byte length and checks
// it (Serializable::checkSignature).
func checkSignature(r *binstream.Reader, expected string) {
	s := r.String(1, 1200)
	if !strings.EqualFold(s, expected) {
		failDB()
	}
}

// writeColumn writes a column name like oStream << const wchar_t*: a 1-byte
// length and the UTF-16 characters, without NUL.
func writeColumn(w *binstream.Writer, name string) {
	w.Byte(byte(len(name)))
	for _, c := range name {
		w.Uint16(uint16(c))
	}
}

// unescape converts "\n", "\r" (both to CR LF) and "\t" in text read from a
// file (String::unescape).
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	r := []rune(s)
	for i := 0; i < len(r); i++ {
		c := r[i]
		if c != '\\' {
			b.WriteRune(c)
			continue
		}
		i++
		var next rune
		if i < len(r) {
			next = r[i]
		}
		switch next {
		case 'n', 'r':
			b.WriteString("\r\n")
		case 't':
			b.WriteRune('\t')
		default:
			b.WriteRune('\\')
			if i < len(r) {
				b.WriteRune(next)
			}
		}
	}
	return b.String()
}

// escape converts CR LF to "\n" (String::escape).
func escape(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}
	var b strings.Builder
	r := []rune(s)
	for i := 0; i < len(r); i++ {
		c := r[i]
		if c != '\r' {
			b.WriteRune(c)
			continue
		}
		i++
		if i < len(r) && r[i] == '\n' {
			b.WriteString(`\n`)
		} else {
			b.WriteRune('\r')
			if i < len(r) {
				b.WriteRune(r[i])
			}
		}
	}
	return b.String()
}

// sqlQuote quotes text for SQL, or returns null (String::sqlQuote).
func sqlQuote(s string) string {
	if s == "" {
		return "null"
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// compareLogical compares two strings like String::compareLogical.
func compareLogical(a, b string) int { return natural.Compare(a, b) }

func cmp[T int | int8 | uint8 | int16 | uint16 | int32 | uint32](a, b T) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	}
	return 1
}

// flntRow keeps the original bytes of a DB row so that cells whose
// canonical encoding differs (FlntRows in shStudio) are written back
// unchanged.
type flntRow struct {
	original, canon []byte
}

func rememberRow(raw, canon []byte) *flntRow {
	if len(raw) != len(canon) || len(raw) == 0 || len(raw)%8 != 0 || string(raw) == string(canon) {
		return nil
	}
	return &flntRow{original: append([]byte(nil), raw...), canon: append([]byte(nil), canon...)}
}

func (f *flntRow) restore(out []byte) {
	if f == nil || len(f.canon) != len(out) {
		return
	}
	for c := 0; c < len(out); c += 8 {
		if string(out[c:c+8]) == string(f.canon[c:c+8]) {
			copy(out[c:c+8], f.original[c:c+8])
		}
	}
}

// itoa helpers matching String << (type) conversions.
func u8s(v uint8) string   { return fmt.Sprint(v) }
func i16s(v int16) string  { return fmt.Sprint(v) }
func u16s(v uint16) string { return fmt.Sprint(v) }
func i32s(v int32) string  { return fmt.Sprint(v) }
