package app

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/imgui"
	"golang.org/x/image/bmp"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/imageio"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/filedlg"
)

var (
	errEmpty   = errors.New("empty file")
	errAborted = errors.New("aborted")
)

// imageView shows a picture (ImageView): .bmp, .dds, .gif, .jpg, .png and
// .tga files.
type imageView struct {
	file  *archive.File
	data  []byte
	ext   string
	img   *image.NRGBA
	tex   *backend.Texture
	ready bool
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		switch ext {
		case ".bmp", ".dds", ".gif", ".jpg", ".png", ".tga":
			return &imageView{file: f}
		}
		return nil
	})
}

func (v *imageView) Path() string  { return v.file.Path }
func (v *imageView) Title() string { return v.file.Name }
func (v *imageView) setReady()     { v.ready = true }
func (v *imageView) Close(*App) bool {
	if v.tex != nil {
		v.tex.Release()
	}
	v.file.State = 0
	return true
}

// load reads the picture; a missing .tga is looked up as .dds
// (ImageView::onCreate).
func (v *imageView) load(a *App, done func(err error)) {
	path := v.file.Path
	data, err := readFile(path, false)
	if err != nil && strings.HasSuffix(strings.ToLower(path), ".tga") {
		path = strings.TrimSuffix(path, filepath.Ext(path)) + ".dds"
		data, err = readFile(path, false)
	}
	if err != nil {
		done(err)
		return
	}
	v.ext = strings.ToLower(filepath.Ext(path))
	img, err := imageio.Decode(data, path)
	if err != nil {
		done(err)
		return
	}
	v.data, v.img = data, img
	done(nil)
}

func (v *imageView) Draw(a *App) bool {
	if !v.ready {
		imgui.TextUnformatted("Loading data...")
		return true
	}
	keep := true
	if imgui.BeginMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolV("Close", "Ctrl+W", false, true) {
				keep = false
			}
			if imgui.MenuItemBool("Save As …") {
				v.saveAs(a)
			}
			imgui.Separator()
			imgui.MenuItemBoolV("Information", "Ctrl+I", false, false)
			imgui.EndMenu()
		}
		imgui.EndMenuBar()
	}
	if imgui.IsWindowFocusedV(imgui.FocusedFlagsChildWindows) && imgui.CurrentIO().KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyW) {
		keep = false
	}
	if v.tex == nil {
		v.tex = backend.NewTextureFromRgba(&image.RGBA{Pix: v.img.Pix, Stride: v.img.Stride, Rect: v.img.Rect})
	}
	b := v.img.Bounds()
	if imgui.BeginChildStrV("##pict", imgui.NewVec2(0, 0), 0, imgui.WindowFlagsHorizontalScrollbar) {
		imgui.Image(v.tex.ID, imgui.NewVec2(float32(b.Dx()), float32(b.Dy())))
	}
	imgui.EndChild()
	return keep
}

// saveAs saves the original bytes, or a BMP (ImageView "Save As").
func (v *imageView) saveAs(a *App) {
	filters := []filedlg.Filter{{Name: "Bitmap File", Patterns: []string{"*.bmp"}}}
	switch v.ext {
	case ".dds":
		filters = append([]filedlg.Filter{{Name: "DirectDraw Surface File", Patterns: []string{"*.dds"}}}, filters...)
	case ".tga":
		filters = append([]filedlg.Filter{{Name: "Targa File", Patterns: []string{"*.tga"}}}, filters...)
	}
	name := strings.TrimSuffix(v.file.Name, filepath.Ext(v.file.Name)) + filepath.Ext(filters[0].Patterns[0])
	a.pickFile(filedlg.New("Save As", filedlg.SaveFile, name, filters...), func(path string) {
		if strings.HasSuffix(strings.ToLower(path), ".bmp") {
			fh, err := os.Create(path)
			if err == nil {
				bmp.Encode(fh, v.img)
				fh.Close()
			}
			return
		}
		os.WriteFile(path, v.data, 0o644)
	})
}
