package imageio

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
)

// TestArchiveImages decodes a few client images from SHSTUDIO_SAF and,
// with SHSTUDIO_PNG_DIR set, writes them as PNG for a visual check.
func TestArchiveImages(t *testing.T) {
	saf := os.Getenv("SHSTUDIO_SAF")
	if saf == "" {
		t.Skip("SHSTUDIO_SAF not set")
	}
	a := archive.New(949)
	if err := a.Open(saf); err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.ReadIndex(nil); err != nil {
		t.Fatal(err)
	}
	n := 0
	var visit func(d *archive.Folder)
	counts := map[string][2]int{}
	visit = func(d *archive.Folder) {
		for _, f := range d.Files {
			ext := filepath.Ext(f.Name)
			if ext != ".dds" && ext != ".DDS" && ext != ".tga" && ext != ".TGA" {
				continue
			}
			data, err := a.ReadFile(f)
			if err != nil {
				continue
			}
			img, err := Decode(data, f.Name)
			c := counts[ext]
			if err == nil {
				c[0]++
			} else {
				c[1]++
			}
			counts[ext] = c
			if dir := os.Getenv("SHSTUDIO_PNG_DIR"); dir != "" && err == nil && n < 6 && (f.Name == "icon_somo2.dds" || f.Name == "icon_skill.dds" || f.Name == "icon_rapis.dds") {
				fh, _ := os.Create(filepath.Join(dir, f.Name+".png"))
				png.Encode(fh, img)
				fh.Close()
				n++
			}
		}
		for _, s := range d.Folders {
			visit(s)
		}
	}
	visit(&a.Root.Folder)
	t.Logf("decoded/failed per extension: %v", counts)
}
