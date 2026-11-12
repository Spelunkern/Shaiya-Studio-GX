// Package binstream provides the binary reader and writer used by every
// file format, ported from shStudio's iStream/oStream (src/core).
//
// Like the C++ streams, which throw on malformed input, Reader panics with an
// *Error when it runs out of data; parsers wrap their work in Try to turn
// that into a returned error.
package binstream

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/codepage"
)

// Error is the panic value raised by Reader on malformed input.
type Error struct{ Msg string }

func (e *Error) Error() string { return e.Msg }

// Fail raises a stream error; format parsers use it like a C++ throw.
func Fail(format string, args ...any) {
	panic(&Error{Msg: fmt.Sprintf(format, args...)})
}

// Try runs f and converts a stream panic into an error. Other panics are
// re-raised.
func Try(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var se *Error
			if e, ok := r.(error); ok && errors.As(e, &se) {
				err = se
				return
			}
			panic(r)
		}
	}()
	f()
	return nil
}

// MaxString is the longest string a stream accepts (String::read).
const MaxString = 0x4000

// Reader reads little-endian (or byte-swapped) values from a byte slice.
type Reader struct {
	Data     []byte
	Pos      int
	Swap     bool
	CodePage int
}

// NewReader returns a Reader over data with the given code page.
func NewReader(data []byte, cp int) *Reader { return &Reader{Data: data, CodePage: cp} }

// Len returns the number of unread bytes.
func (r *Reader) Len() int { return len(r.Data) - r.Pos }

// Size returns the total size.
func (r *Reader) Size() int { return len(r.Data) }

func (r *Reader) need(n int) {
	if n < 0 || r.Len() < n {
		Fail("array index out of bounds: %d available, %d needed", r.Len(), n)
	}
}

// Skip advances the position by n bytes.
func (r *Reader) Skip(n int) {
	r.need(n)
	r.Pos += n
}

// Byte reads one byte.
func (r *Reader) Byte() byte {
	r.need(1)
	v := r.Data[r.Pos]
	r.Pos++
	return v
}

// Int8 reads a signed byte.
func (r *Reader) Int8() int8 { return int8(r.Byte()) }

// Uint16 reads a 16-bit value.
func (r *Reader) Uint16() uint16 {
	r.need(2)
	v := binary.LittleEndian.Uint16(r.Data[r.Pos:])
	if r.Swap {
		v = v>>8 | v<<8
	}
	r.Pos += 2
	return v
}

// Int16 reads a signed 16-bit value.
func (r *Reader) Int16() int16 { return int16(r.Uint16()) }

// PeekUint32 reads a 32-bit value without advancing.
func (r *Reader) PeekUint32() uint32 {
	r.need(4)
	v := binary.LittleEndian.Uint32(r.Data[r.Pos:])
	if r.Swap {
		v = swap32(v)
	}
	return v
}

// Uint32 reads a 32-bit value.
func (r *Reader) Uint32() uint32 {
	v := r.PeekUint32()
	r.Pos += 4
	return v
}

// Int32 reads a signed 32-bit value.
func (r *Reader) Int32() int32 { return int32(r.Uint32()) }

// Int reads a 32-bit value as int.
func (r *Reader) Int() int { return int(int32(r.Uint32())) }

// Int64 reads a 64-bit value.
func (r *Reader) Int64() int64 {
	r.need(8)
	v := binary.LittleEndian.Uint64(r.Data[r.Pos:])
	if r.Swap {
		v = swap64(v)
	}
	r.Pos += 8
	return int64(v)
}

// Float32 reads a float (never byte-swapped, like iStream::readFloat).
func (r *Reader) Float32() float32 {
	r.need(4)
	v := math.Float32frombits(binary.LittleEndian.Uint32(r.Data[r.Pos:]))
	r.Pos += 4
	return v
}

// Bytes reads n bytes; the result aliases the input.
func (r *Reader) Bytes(n int) []byte {
	r.need(n)
	b := r.Data[r.Pos : r.Pos+n : r.Pos+n]
	r.Pos += n
	return b
}

// Copy reads n bytes into a new slice.
func (r *Reader) Copy(n int) []byte { return append([]byte(nil), r.Bytes(n)...) }

// Read fills dst.
func (r *Reader) Read(dst []byte) { copy(dst, r.Bytes(len(dst))) }

// String reads a length-prefixed string (String::read). lengthSize is 1 or
// 4; cp -1 means the reader's code page. Trailing NULs are dropped.
func (r *Reader) String(lengthSize, cp int) string {
	if cp == -1 {
		cp = r.CodePage
	}
	var n int
	switch lengthSize {
	case 1:
		n = int(r.Byte())
	case 4:
		n = int(r.Uint32())
	default:
		Fail("Invalid size of string's length")
	}
	if n < 0 || n > MaxString {
		Fail("Invalid stream format")
	}
	if n == 0 {
		return ""
	}
	if cp == codepage.UTF16LE {
		b := r.Bytes(2 * n)
		return trimNul(codepage.Decode(b, cp))
	}
	b := r.Bytes(n)
	for n > 0 && b[n-1] == 0 {
		n--
	}
	return codepage.Decode(b[:n], cp)
}

// FixedString reads a fixed-size NUL-padded string field.
func (r *Reader) FixedString(size, cp int) string {
	if cp == -1 {
		cp = r.CodePage
	}
	return CString(r.Bytes(size), cp)
}

// CString decodes a NUL-terminated byte field.
func CString(b []byte, cp int) string {
	for i, c := range b {
		if c == 0 {
			b = b[:i]
			break
		}
	}
	return codepage.Decode(b, cp)
}

func trimNul(s string) string {
	for len(s) > 0 && s[len(s)-1] == 0 {
		s = s[:len(s)-1]
	}
	return s
}

func swap32(v uint32) uint32 {
	return v>>24 | (v>>8)&0xFF00 | (v<<8)&0xFF0000 | v<<24
}

func swap64(v uint64) uint64 {
	return uint64(swap32(uint32(v)))<<32 | uint64(swap32(uint32(v>>32)))
}

// Writer builds a byte buffer (oStream).
type Writer struct {
	Buf      []byte
	Swap     bool
	CodePage int
	// WriteNul appends a NUL to strings (oStream default: true).
	WriteNul bool
}

// NewWriter returns a Writer with the oStream defaults.
func NewWriter(cp int) *Writer { return &Writer{CodePage: cp, WriteNul: true} }

// Len returns the number of bytes written.
func (w *Writer) Len() int { return len(w.Buf) }

// Bytes returns the written bytes.
func (w *Writer) Bytes() []byte { return w.Buf }

// Put appends raw bytes.
func (w *Writer) Put(b []byte) { w.Buf = append(w.Buf, b...) }

// Byte appends one byte.
func (w *Writer) Byte(v byte) { w.Buf = append(w.Buf, v) }

// Uint16 appends a 16-bit value.
func (w *Writer) Uint16(v uint16) {
	if w.Swap {
		v = v>>8 | v<<8
	}
	w.Buf = binary.LittleEndian.AppendUint16(w.Buf, v)
}

// Int16 appends a signed 16-bit value.
func (w *Writer) Int16(v int16) { w.Uint16(uint16(v)) }

// Uint32 appends a 32-bit value.
func (w *Writer) Uint32(v uint32) {
	if w.Swap {
		v = swap32(v)
	}
	w.Buf = binary.LittleEndian.AppendUint32(w.Buf, v)
}

// Int32 appends a signed 32-bit value.
func (w *Writer) Int32(v int32) { w.Uint32(uint32(v)) }

// Int appends an int as 32 bits.
func (w *Writer) Int(v int) { w.Uint32(uint32(int32(v))) }

// Int64 appends a 64-bit value.
func (w *Writer) Int64(v int64) {
	u := uint64(v)
	if w.Swap {
		u = swap64(u)
	}
	w.Buf = binary.LittleEndian.AppendUint64(w.Buf, u)
}

// Float32 appends a float.
func (w *Writer) Float32(v float32) {
	w.Buf = binary.LittleEndian.AppendUint32(w.Buf, math.Float32bits(v))
}

func (w *Writer) putLength(n uint32, lengthSize int) {
	if w.Swap {
		n = swap32(n)
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], n)
	w.Buf = append(w.Buf, b[:lengthSize]...)
}

// String appends a length-prefixed string (String::write). cp -1 means the
// writer's code page.
func (w *Writer) String(s string, lengthSize, cp int) {
	if cp == -1 {
		cp = w.CodePage
	}
	if s == "" {
		if w.WriteNul {
			w.putLength(1, lengthSize)
			if cp == codepage.UTF16LE {
				w.Buf = append(w.Buf, 0, 0)
			} else {
				w.Buf = append(w.Buf, 0)
			}
		} else {
			w.putLength(0, lengthSize)
		}
		return
	}
	b := codepage.Encode(s, cp)
	if cp == codepage.UTF16LE {
		n := len(b) / 2
		if w.WriteNul {
			b = append(b, 0, 0)
			n++
		}
		w.putLength(uint32(n), lengthSize)
		w.Put(b)
		return
	}
	if w.WriteNul {
		b = append(b, 0)
	}
	w.putLength(uint32(len(b)), lengthSize)
	w.Put(b)
}

// FixedString appends s encoded and NUL-padded (or truncated) to size bytes.
func (w *Writer) FixedString(s string, size, cp int) {
	if cp == -1 {
		cp = w.CodePage
	}
	b := make([]byte, size)
	copy(b, codepage.Encode(s, cp))
	w.Put(b)
}
