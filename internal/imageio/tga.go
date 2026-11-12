package imageio

import (
	"encoding/binary"
	"image"
)

// DecodeTGA decodes an uncompressed or RLE true-color TGA of 24 or 32 bits
// (loadTGA). Rows are stored bottom-up, as shStudio assumes for every TGA.
func DecodeTGA(data []byte) (*image.NRGBA, error) {
	if len(data) < 18 {
		return nil, ErrFormat
	}
	hd := data[:18]
	w := int(int16(binary.LittleEndian.Uint16(hd[12:])))
	h := int(int16(binary.LittleEndian.Uint16(hd[14:])))
	typ, bpp := hd[2], hd[16]
	switch typ {
	case 1, 2, 9, 10, 3, 11:
	default:
		return nil, ErrUnsupported
	}
	if hd[1] != 0 || typ&7 != 2 || bpp < 24 || w <= 0 || h <= 0 {
		return nil, ErrUnsupported
	}
	four := bpp > 24
	src := data[18:]
	if int(hd[0]) > len(src) {
		return nil, ErrFormat
	}
	src = src[hd[0]:]
	pix := make([]byte, 4*w*h) // BGRA, bottom-up
	at := 0
	next := func() byte {
		if at >= len(src) {
			panic(ErrFormat)
		}
		b := src[at]
		at++
		return b
	}
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = ErrFormat
			}
		}()
		switch {
		case typ == 2 && bpp == 24:
			for i := 0; i < w*h; i++ {
				p := pix[4*i:]
				p[0], p[1], p[2] = next(), next(), next()
			}
		case typ == 2 && bpp == 32:
			for i := 0; i < w*h; i++ {
				p := pix[4*i:]
				p[0], p[1], p[2], p[3] = next(), next(), next(), next()
			}
		case typ == 10:
			for p := 0; p < len(pix); {
				packet := next()
				count := int(packet&0x7F) + 1
				if packet&0x80 == 0 {
					for ; count > 0 && p < len(pix); count-- {
						pix[p], pix[p+1], pix[p+2] = next(), next(), next()
						if four {
							pix[p+3] = next()
						}
						p += 4
					}
				} else {
					b, g, r := next(), next(), next()
					var a byte
					if four {
						a = next()
					}
					for ; count > 0 && p < len(pix); count-- {
						pix[p], pix[p+1], pix[p+2], pix[p+3] = b, g, r, a
						p += 4
					}
				}
			}
		default:
			return ErrUnsupported
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		s := pix[4*w*(h-1-y):]
		d := out.Pix[y*out.Stride:]
		for x := 0; x < w; x++ {
			a := byte(0xFF)
			if four {
				a = s[4*x+3]
			}
			d[4*x], d[4*x+1], d[4*x+2], d[4*x+3] = s[4*x+2], s[4*x+1], s[4*x], a
		}
	}
	return out, nil
}
