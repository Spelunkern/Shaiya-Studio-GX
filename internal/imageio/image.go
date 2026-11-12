package imageio

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
)

// Decode decodes data named name: DDS (by magic or extension), TGA, or a
// format of the standard decoders (BMP, JPG, PNG, GIF), as ImageView and
// the icon loaders do. Client ".dds" files may hold TGA data.
func Decode(data []byte, name string) (*image.NRGBA, error) {
	ext := strings.ToLower(filepath.Ext(name))
	if HasDDSMagic(data) {
		return DecodeDDS(data)
	}
	if ext == ".tga" || ext == ".dds" {
		return DecodeTGA(data)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if n, ok := img.(*image.NRGBA); ok {
		return n, nil
	}
	out := image.NewNRGBA(img.Bounds())
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
	return out, nil
}
