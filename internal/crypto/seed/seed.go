// Package seed implements the SEED block cipher used by Shaiya to encrypt
// BinarySData data files (DB*Data.SData), ported from shStudio
// src/crypto/SEED.cpp.
package seed

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// DataFileSignature is the 40-byte ASCII signature at the start of an
// encrypted data file header.
const DataFileSignature = "0001CBCEBC5B2784D3FC9A2A9DB84D1C3FEB6E99"

// HeaderSize is the size of the encrypted data file header.
const HeaderSize = 0x40

// ErrBlockSize is returned when the input is not a multiple of 16 bytes.
var ErrBlockSize = errors.New("seed: input is not a multiple of the block size")

// Cipher is a SEED cipher keyed with the data file round keys.
type Cipher struct {
	keys [32]uint32
}

// NewDataFileCipher returns the cipher keyed for Shaiya data files.
func NewDataFileCipher() *Cipher {
	c := &Cipher{}
	for i := range c.keys {
		c.keys[i] = dataFileRoundKeys[i] ^ 0x55AA55AA
	}
	return c
}

func g(x uint32) uint32 {
	return ss0[x&0xFF] ^ ss1[(x>>8)&0xFF] ^ ss2[(x>>16)&0xFF] ^ ss3[x>>24]
}

func swap(x uint32) uint32 {
	return ((x<<8 | x>>24) & 0x00FF00FF) | ((x<<24 | x>>8) & 0xFF00FF00)
}

func round(l0, l1 *uint32, r0, r1 uint32, k []uint32) {
	t0 := r0 ^ k[0]
	t1 := r1 ^ k[1]
	t1 ^= t0
	t1 = g(t1)
	t0 += t1
	t0 = g(t0)
	t1 += t0
	t1 = g(t1)
	t0 += t1
	*l0 ^= t0
	*l1 ^= t1
}

func (c *Cipher) block(dst, src []byte, encrypt bool) {
	le := binary.LittleEndian
	l0, l1 := swap(le.Uint32(src[0:])), swap(le.Uint32(src[4:]))
	r0, r1 := swap(le.Uint32(src[8:])), swap(le.Uint32(src[12:]))
	if encrypt {
		for r := 0; r < 32; r += 4 {
			round(&l0, &l1, r0, r1, c.keys[r:])
			round(&r0, &r1, l0, l1, c.keys[r+2:])
		}
	} else {
		for r := 30; r >= 0; r -= 4 {
			round(&l0, &l1, r0, r1, c.keys[r:])
			round(&r0, &r1, l0, l1, c.keys[r-2:])
		}
	}
	le.PutUint32(dst[0:], swap(r0))
	le.PutUint32(dst[4:], swap(r1))
	le.PutUint32(dst[8:], swap(l0))
	le.PutUint32(dst[12:], swap(l1))
}

// Encrypt encrypts src into dst (which may alias src). len(src) must be a
// multiple of 16.
func (c *Cipher) Encrypt(dst, src []byte) error {
	if len(src)%16 != 0 {
		return ErrBlockSize
	}
	for i := 0; i < len(src); i += 16 {
		c.block(dst[i:], src[i:], true)
	}
	return nil
}

// Decrypt decrypts src into dst (which may alias src). len(src) must be a
// multiple of 16.
func (c *Cipher) Decrypt(dst, src []byte) error {
	if len(src)%16 != 0 {
		return ErrBlockSize
	}
	for i := 0; i < len(src); i += 16 {
		c.block(dst[i:], src[i:], false)
	}
	return nil
}

// IsEncrypted reports whether data starts with the data file signature
// (the check StudioFrame uses when opening files).
func IsEncrypted(data []byte) bool {
	return len(data) >= 16 && bytes.Equal(data[:16], []byte(DataFileSignature[:16]))
}

// Header is the 64-byte header of an encrypted data file.
type Header [HeaderSize]byte

func (h *Header) word(i int) uint32       { return binary.LittleEndian.Uint32(h[i*4:]) }
func (h *Header) setWord(i int, v uint32) { binary.LittleEndian.PutUint32(h[i*4:], v) }

// PlainSize returns the plain data size stored in the header.
func (h *Header) PlainSize() uint32 {
	if h.word(10) != 0 {
		return h.word(11)
	}
	return h.word(12)
}

// DecryptFile decrypts a data file image. It mirrors sub_418C00: when data is
// not an encrypted file (too short, not 16-byte aligned or without the
// signature) it returns data unchanged with encrypted=false. On success it
// returns the plain bytes and the header, which EncryptFile reuses.
func DecryptFile(data []byte) (plain []byte, hdr Header, encrypted bool, err error) {
	size := len(data)
	if size < HeaderSize || size%16 != 0 {
		return data, hdr, false, nil
	}
	copy(hdr[:], data[:HeaderSize])
	if !IsEncrypted(hdr[:]) {
		return data, hdr, false, nil
	}
	encSize := uint32(size - HeaderSize)
	plainSize := hdr.PlainSize()
	if plainSize < encSize && plainSize > encSize-16 {
		plainSize = encSize
	}
	if plainSize > encSize {
		return nil, hdr, true, errors.New("seed: plain size larger than encrypted body")
	}
	buf := make([]byte, encSize)
	if err := NewDataFileCipher().Decrypt(buf, data[HeaderSize:]); err != nil {
		return nil, hdr, true, err
	}
	return buf[:plainSize], hdr, true, nil
}

// EncryptFile encrypts plain using hdr (as returned by DecryptFile, or zero
// for a new file) and returns the updated header and the encrypted body. It
// mirrors sub_418E60; the caller writes hdr followed by the body.
func EncryptFile(plain []byte, hdr Header) (Header, []byte, error) {
	size := len(plain)
	if size == 0 {
		return hdr, nil, errors.New("seed: empty data")
	}
	crcAtSecond := hdr.word(10) == 0
	zero := true
	for _, b := range hdr[:40] {
		if b != 0 {
			zero = false
			break
		}
	}
	if zero {
		copy(hdr[:40], DataFileSignature)
	}
	if crcAtSecond {
		hdr.setWord(12, uint32(size))
	} else {
		hdr.setWord(11, uint32(size))
	}
	crc := ^uint32(0)
	for _, b := range plain {
		crc = dataFileCRC[byte(crc)^b] ^ (crc >> 8)
	}
	crc = ^crc
	if crcAtSecond {
		hdr.setWord(11, crc)
	} else {
		hdr.setWord(10, crc)
	}
	padded := size
	if size%16 != 0 {
		padded = size - size%16 + 16
	}
	buf := make([]byte, padded)
	copy(buf, plain)
	if err := NewDataFileCipher().Encrypt(buf, buf); err != nil {
		return hdr, nil, err
	}
	return hdr, buf, nil
}
