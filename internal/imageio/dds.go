// Package imageio decodes the image formats of the client (DDS, TGA, and
// BMP/JPG/PNG/GIF through the standard decoders), ported from shStudio
// src/core/image/imageio.cpp. Images are returned as *image.NRGBA.
package imageio

import (
	"encoding/binary"
	"errors"
	"image"
)

// Errors.
var (
	ErrFormat      = errors.New("imageio: invalid or unsupported image")
	ErrUnsupported = errors.New("imageio: unsupported image kind")
)

// ddsHeader is the 128-byte DDS header (DdsHeader).
type ddsHeader struct {
	magic, size, flags, height, width, pitchOrLinearSize, depth, mipMapCount uint32
	pfSize, pfFlags, fourCC, rgbBitCount, rMask, gMask, bMask, aMask         uint32
	caps, caps2                                                              uint32
}

func parseDDSHeader(b []byte) ddsHeader {
	u := func(o int) uint32 { return binary.LittleEndian.Uint32(b[o:]) }
	return ddsHeader{
		magic: u(0), size: u(4), flags: u(8), height: u(12), width: u(16), pitchOrLinearSize: u(20),
		depth: u(24), mipMapCount: u(28), pfSize: u(76), pfFlags: u(80), fourCC: u(84),
		rgbBitCount: u(88), rMask: u(92), gMask: u(96), bMask: u(100), aMask: u(104),
		caps: u(108), caps2: u(112),
	}
}

// format classifies the pixel format (DdsHeader::getFormat).
func (h *ddsHeader) format() int {
	var format int
	var size uint32
	pf := h.pfFlags
	if pf&4 == 0 {
		switch {
		case pf&0x20000 != 0:
			format = 9
			if pf&1 != 0 {
				format = 10
			}
		case pf&1 == 0:
			format = 1
		default:
			format = 0
		}
		size = h.height * h.width * h.depth * (h.rgbBitCount >> 3)
	} else {
		blocks := h.depth * (h.width + 3) / 4 * ((h.height + 3) / 4)
		pixels := h.width * h.height * h.depth
		switch h.fourCC {
		case 0x31545844: // DXT1
			format, size = 2, 8*blocks
		case 0x32545844: // DXT2
			format, size = 3, 16*blocks
		case 0x33545844: // DXT3
			format, size = 4, 16*blocks
		case 0x34545844: // DXT4
			format, size = 5, 16*blocks
		case 0x35545844: // DXT5
			format, size = 6, 16*blocks
		case 0x32495441: // ATI2
			format, size = 7, 16*blocks
		case 0x31495441: // ATI1
			format, size = 8, 8*blocks
		case 0x42475852: // RXGB
			format, size = 11, 16*blocks
		case 36:
			format, size = 12, 8*pixels
		case 111:
			format, size = 13, 2*pixels
		case 112:
			format, size = 14, 4*pixels
		case 113:
			format, size = 15, 8*pixels
		case 114:
			format, size = 16, 4*pixels
		case 115:
			format, size = 17, 8*pixels
		case 116:
			format, size = 18, 16*pixels
		default:
			format, size = 255, 16*blocks
		}
	}
	if h.flags&0x80008 == 0 || h.pitchOrLinearSize == 0 {
		h.flags |= 0x80000
		h.pitchOrLinearSize = size
	}
	return format
}

func (h *ddsHeader) bytesPerPixel(format int) uint32 {
	switch format {
	case 9, 10, 0:
		return h.rgbBitCount >> 3
	case 1, 7, 11:
		return 3
	case 8:
		return 1
	case 12, 15, 17:
		return 8
	case 18:
		return 16
	}
	return 4
}

// adjustSize recomputes the data size of volume textures
// (DdsHeader::adjustSize).
func (h *ddsHeader) adjustSize(format int) {
	if int32(h.depth) <= 1 {
		return
	}
	if h.caps2&0x200000 == 0 {
		h.depth = 1
	}
	w, hh := max(int32(h.width), 1), max(int32(h.height), 1)
	switch format {
	case 0, 1, 9, 10:
		h.pitchOrLinearSize = (h.rgbBitCount >> 3) * uint32(w) * uint32(hh)
	case 2, 8:
		h.pitchOrLinearSize = 8 * (h.height + 3) / 4 * ((h.width + 3) / 4)
	case 3, 4, 5, 6, 7, 11:
		h.pitchOrLinearSize = 16 * (h.height + 3) / 4 * ((h.width + 3) / 4)
	case 12, 13, 14, 15, 16, 17, 18:
		h.pitchOrLinearSize = uint32(w) * (uint32(hh) * h.bytesPerPixel(format))
	}
	h.flags |= 0x80000
	h.pitchOrLinearSize *= h.depth
}

func (h *ddsHeader) readData(b []byte) []byte {
	if h.flags&0x80000 != 0 {
		if uint32(len(b)) < h.pitchOrLinearSize {
			return nil
		}
		return b[:h.pitchOrLinearSize]
	}
	n := h.height * ((h.width * h.rgbBitCount) >> 3) * h.depth
	if uint32(len(b)) < n {
		return nil
	}
	return b[:n]
}

// ddsImage is the decoded BGRA(+) pixel buffer (DdsImage).
type ddsImage struct {
	width, height, depth uint32
	components           uint32
	rowSize, sliceSize   uint32
	data                 []byte
	packed10             bool
}

func componentCount(format int) uint32 {
	switch format {
	case 1, 7, 11, 14, 17, 16, 13:
		return 3
	case 9, 8:
		return 1
	case 10:
		return 2
	}
	return 4
}

// init allocates the buffer like DdsImage::init; only 8-bit layouts are
// decoded by shStudio, the others fail later in decode.
func (im *ddsImage) init(h *ddsHeader, format int) bool {
	count := uint32(4)
	size := uint32(1)
	switch format {
	case 0:
		count = 4
		if im.packed10 {
			size = 2
		}
	case 1:
		count = 4
	case 7:
		count = 3
	case 8, 9:
		count = 1
		if format == 9 && h.rgbBitCount == 16 && h.rMask == 0xFFFF {
			size = 2
		}
	case 10:
		count = 2
	case 12:
		count, size = componentCount(12), 2
	case 13, 14, 15, 16, 17, 18:
		count, size = componentCount(format), 4
	case 11:
		count = 3
	}
	if h.width == 0 {
		h.width = 1
	}
	if h.height == 0 {
		h.height = 1
	}
	if h.depth == 0 {
		h.depth = 1
	}
	im.width, im.height, im.depth = h.width, h.height, h.depth
	im.components = count
	im.rowSize = h.width * count * size
	im.sliceSize = im.rowSize * h.height
	if uint64(im.sliceSize)*uint64(im.depth) > 1<<30 {
		return false
	}
	im.data = make([]byte, im.sliceSize*im.depth)
	return true
}

func maskShift(mask uint32) (shift, bits uint32) {
	if mask == 0 {
		return 0, 0
	}
	var i uint32
	for ; i < 32 && mask&1 == 0; i++ {
		mask >>= 1
	}
	var j uint32
	for ; j < 8 && mask&1 != 0; j++ {
		mask >>= 1
	}
	return i, 8 - j
}

func channel(v, mask, shift, bits uint32) byte { return byte(byte((v&mask)>>shift) << bits) }

func (im *ddsImage) convertMasked(h *ddsHeader, src []byte, format int) bool {
	if format == 9 && h.rgbBitCount == 16 && h.rMask == 0xFFFF {
		copy(im.data, src)
		return true
	}
	rS, rB := maskShift(h.rMask)
	gS, gB := maskShift(h.gMask)
	bS, bB := maskShift(h.bMask)
	aS, aB := maskShift(h.aMask)
	bpp := h.rgbBitCount >> 3
	var v uint32
	at := 0
	get := func(i int) uint32 {
		if at+i < len(src) {
			return uint32(src[at+i])
		}
		return 0
	}
	n := uint32(len(im.data))
	for pos := uint32(0); pos < n; pos += im.components {
		switch {
		case n-pos >= 4:
			v = get(0) | get(1)<<8 | get(2)<<16 | get(3)<<24
		case bpp == 3:
			v = get(0) | get(1)<<8 | get(2)<<16
		case bpp == 1:
			v = get(0)
		case bpp == 2:
			v = get(0) | get(1)<<8
		}
		at += int(bpp)
		d := im.data
		d[pos] = channel(v, h.bMask, bS, bB)
		switch {
		case im.components >= 3:
			d[pos+1] = channel(v, h.gMask, gS, gB)
			d[pos+2] = channel(v, h.rMask, rS, rB)
			if im.components == 4 {
				d[pos+3] = channel(v, h.aMask, aS, aB)
				if aB >= 7 {
					if d[pos+3] != 0 {
						d[pos+3] = 0xFF
					}
				} else if aB >= 4 {
					d[pos+3] |= d[pos+3] >> 4
				}
			}
		case im.components == 2:
			d[pos+1] = channel(v, h.aMask, aS, aB)
			if aB >= 7 {
				if d[pos+1] != 0 {
					d[pos+1] = 0xFF
				}
			} else if aB >= 4 && pos+3 < n {
				d[pos+1] |= d[pos+3] >> 4
			}
		}
	}
	return true
}

// expandColors builds the 4-color palette of a DXT3/DXT5 block.
func expandColors(p *[16]byte, lo0, hi0, lo1, hi1 byte) {
	g0 := lo0>>5 | (hi0&7)<<3
	g1 := lo1>>5 | (hi1&7)<<3
	p[0] = (lo0&0x1F)<<3 | (lo0&0x1F)>>2
	p[1] = g0<<2 | g0>>3
	p[2] = (hi0>>3)<<3 | hi0>>5
	p[4] = (lo1&0x1F)<<3 | (lo1&0x1F)>>2
	p[5] = g1<<2 | g1>>3
	p[6] = (hi1>>3)<<3 | hi1>>5
	p[8] = byte((int(p[4]) + 2*int(p[0]) + 1) / 3)
	p[9] = byte((int(p[5]) + 2*int(p[1]) + 1) / 3)
	p[10] = byte((int(p[6]) + 2*int(p[2]) + 1) / 3)
	p[12] = byte((int(p[0]) + 2*int(p[4]) + 1) / 3)
	p[13] = byte((int(p[1]) + 2*int(p[5]) + 1) / 3)
	p[14] = byte((int(p[2]) + 2*int(p[6]) + 1) / 3)
}

func (im *ddsImage) put(z, row, px uint32, rgb []byte, n int) {
	if px < im.width && row < im.height {
		o := z*im.sliceSize + row*im.rowSize + px*im.components
		copy(im.data[o:o+uint32(n)], rgb[:n])
	}
}

func (im *ddsImage) colorBlock(z, y, x uint32, p *[16]byte, bits uint32, n int) {
	row := y
	for shift := 4; shift < 36; shift, row = shift+8, row+1 {
		for k := 0; k < 4; k++ {
			e := p[4*((bits>>(shift-4+2*k))&3):]
			im.put(z, row, x+uint32(k), e[:], n)
		}
	}
}

func (im *ddsImage) decodeDXT1(src []byte) bool {
	var p [16]byte
	p[3], p[7], p[11] = 0xFF, 0xFF, 0xFF
	at := 0
	for z := uint32(0); z < im.depth; z++ {
		for y := uint32(0); y < im.height; y += 4 {
			for x := uint32(0); x < im.width; x += 4 {
				if at+8 > len(src) {
					return true
				}
				c0 := binary.LittleEndian.Uint16(src[at:])
				c1 := binary.LittleEndian.Uint16(src[at+2:])
				bits := binary.LittleEndian.Uint32(src[at+4:])
				at += 8
				r0, r1 := byte((c0>>11)&0x1F), byte((c1>>11)&0x1F)
				g0, g1 := byte((c0>>5)&0x3F), byte((c1>>5)&0x3F)
				p[2] = r0<<3 | r0>>2
				p[1] = g0<<2 | g0>>3
				p[0] = byte(c0)<<3 | r0>>2
				p[6] = r1<<3 | r1>>2
				p[5] = g1<<2 | g1>>3
				p[4] = byte(c1)<<3 | r1>>2
				if c0 <= c1 {
					p[10] = byte((int(p[2]) + int(p[6])) / 2)
					p[9] = byte((int(p[1]) + int(p[5])) / 2)
					p[8] = byte((int(p[0]) + int(p[4])) / 2)
					p[15] = 0
				} else {
					p[10] = byte((int(p[6]) + 2*int(p[2]) + 1) / 3)
					p[9] = byte((int(p[5]) + 2*int(p[1]) + 1) / 3)
					p[8] = byte((int(p[4]) + 2*int(p[0]) + 1) / 3)
					p[15] = 0xFF
				}
				p[14] = byte((int(p[2]) + 2*int(p[6]) + 1) / 3)
				p[13] = byte((int(p[1]) + 2*int(p[5]) + 1) / 3)
				p[12] = byte((int(p[0]) + 2*int(p[4]) + 1) / 3)
				im.colorBlock(z, y, x, &p, bits, 4)
			}
		}
	}
	return true
}

func (im *ddsImage) decodeDXT3(src []byte) bool {
	var p [16]byte
	at := 0
	for z := uint32(0); z < im.depth; z++ {
		for y := uint32(0); y < im.height; y += 4 {
			for x := uint32(0); x < im.width; x += 4 {
				if at+16 > len(src) {
					return true
				}
				b := src[at : at+16]
				at += 16
				expandColors(&p, b[8], b[9], b[10], b[11])
				im.colorBlock(z, y, x, &p, binary.LittleEndian.Uint32(b[12:]), 3)
				row := y
				for r := 0; r < 4; r, row = r+1, row+1 {
					w := uint32(b[2*r]) | uint32(b[2*r+1])<<8
					alpha := [4]byte{b[2*r] & 0xF, byte((w >> 4) & 0xF), byte((w >> 8) & 0xF), byte(w >> 12)}
					for k := 0; k < 4; k++ {
						px := x + uint32(k)
						if px < im.width && row < im.height {
							o := row*im.rowSize + px*im.components + z*im.sliceSize + 3
							im.data[o] = alpha[k] | alpha[k]<<4
						}
					}
				}
			}
		}
	}
	return true
}

func (im *ddsImage) decodeDXT5(src []byte) bool {
	var p [16]byte
	var alpha [8]byte
	at := 0
	for z := uint32(0); z < im.depth; z++ {
		for y := uint32(0); y < im.height; y += 4 {
			for x := uint32(0); x < im.width; x += 4 {
				if at+16 > len(src) {
					return true
				}
				b := src[at : at+16]
				at += 16
				a0, a1 := int(b[0]), int(b[1])
				alpha[0], alpha[1] = b[0], b[1]
				expandColors(&p, b[8], b[9], b[10], b[11])
				im.colorBlock(z, y, x, &p, binary.LittleEndian.Uint32(b[12:]), 3)
				if a0 <= a1 {
					alpha[2] = byte((a1 + 4*a0 + 2) / 5)
					alpha[6] = 0
					alpha[7] = 0xFF
					alpha[3] = byte((a0 + 2*(a0+1) + 2*a1) / 5)
					alpha[4] = byte((a1 + 2*(a1+1) + 2*a0) / 5)
					alpha[5] = byte((a0 + 4*a1 + 2) / 5)
				} else {
					alpha[2] = byte((a1 + 6*a0 + 3) / 7)
					alpha[3] = byte((a0 + 4*a0 + 3 + 2*a1) / 7)
					alpha[4] = byte((3*(a1+1) + 4*a0) / 7)
					alpha[5] = byte((3*(a0+1) + 4*a1) / 7)
					alpha[6] = byte((a1 + 4*a1 + 3 + 2*a0) / 7)
					alpha[7] = byte((a0 + 6*a1 + 3) / 7)
				}
				for half := 0; half < 2; half++ {
					ab := b[2+3*half:]
					idx := uint32(ab[0]) | uint32(binary.LittleEndian.Uint16(ab[1:]))<<8
					row := y + uint32(2*half)
					for r := 0; r < 2; r, row = r+1, row+1 {
						for k := 0; k < 4; k++ {
							px := x + uint32(k)
							if px < im.width && row < im.height {
								im.data[z*im.sliceSize+row*im.rowSize+px*im.components+3] = alpha[idx&7]
							}
							idx >>= 3
						}
					}
				}
			}
		}
	}
	return true
}

func (im *ddsImage) unpremultiply() {
	for i := 0; i+3 < len(im.data); i += 4 {
		if a := uint32(im.data[i+3]); a != 0 {
			im.data[i] = byte((uint32(im.data[i]) << 8) / a)
			im.data[i+1] = byte((uint32(im.data[i+1]) << 8) / a)
			im.data[i+2] = byte((uint32(im.data[i+2]) << 8) / a)
		}
	}
}

func (im *ddsImage) decode(h *ddsHeader, src []byte, format int) bool {
	switch format {
	case 0, 1, 9, 10:
		if im.packed10 {
			return false
		}
		return im.convertMasked(h, src, format)
	case 2:
		return im.decodeDXT1(src)
	case 3:
		im.decodeDXT3(src)
		im.unpremultiply()
		return true
	case 4:
		return im.decodeDXT3(src)
	case 5:
		im.decodeDXT5(src)
		im.unpremultiply()
		return true
	case 6:
		return im.decodeDXT5(src)
	case 12:
		copy(im.data, src)
		return true
	}
	return false
}

// HasDDSMagic reports whether data starts with "DDS ".
func HasDDSMagic(data []byte) bool {
	return len(data) >= 4 && binary.LittleEndian.Uint32(data) == 0x20534444
}

// DecodeDDS decodes the first surface of a DDS file (loadDDS). Like
// shStudio, the alpha of formats without alpha is 0xFF only for DXT1 and
// masked formats with an alpha mask; it is kept as decoded.
func DecodeDDS(data []byte) (*image.NRGBA, error) {
	if len(data) < 128 {
		return nil, ErrFormat
	}
	h := parseDDSHeader(data)
	if h.depth == 0 {
		h.depth = 1
	}
	if h.magic != 0x20534444 || (h.size != 124 && h.size != 0x20534444) || h.pfSize != 32 || h.width == 0 || h.height == 0 {
		return nil, ErrFormat
	}
	format := h.format()
	if format == 255 {
		return nil, ErrFormat
	}
	packed := h.rMask == 0x3FF00000 && h.gMask == 0xFFC00 && h.bMask == 0x3FF && h.aMask == 0xC0000000 ||
		h.rMask == 0x3FF && h.gMask == 0xFFC00 && h.bMask == 0x3FF00000 && h.aMask == 0xC0000000
	if h.caps&8 != 0 && h.caps2&0x200 != 0 {
		return nil, ErrUnsupported // cube maps
	}
	im := &ddsImage{packed10: packed}
	h.adjustSize(format)
	raw := h.readData(data[128:])
	if raw == nil || !im.init(&h, format) || !im.decode(&h, raw, format) {
		return nil, ErrFormat
	}
	return im.toNRGBA(), nil
}

// toNRGBA converts the first slice of the BGRA(+) buffer, like the 32-bit
// DIB that toBitmap builds (only 4-component buffers keep their alpha).
func (im *ddsImage) toNRGBA() *image.NRGBA {
	w, h := int(im.width), int(im.height)
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	c := int(im.components)
	for y := 0; y < h; y++ {
		src := im.data[y*int(im.rowSize):]
		dst := out.Pix[y*out.Stride:]
		for x := 0; x < w; x++ {
			s := src[x*c:]
			d := dst[x*4:]
			switch c {
			case 4:
				d[0], d[1], d[2], d[3] = s[2], s[1], s[0], s[3]
			case 3:
				d[0], d[1], d[2], d[3] = s[2], s[1], s[0], 0xFF
			case 2:
				d[0], d[1], d[2], d[3] = s[0], s[0], s[0], s[1]
			default:
				d[0], d[1], d[2], d[3] = s[0], s[0], s[0], 0xFF
			}
		}
	}
	return out
}
