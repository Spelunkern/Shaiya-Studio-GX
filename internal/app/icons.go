package app

import (
	"image"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/icons"
)

// iconLoader returns the icon loader of the open workspace (the icon part
// of ClientData), creating it on first use.
func (a *App) iconLoader() *icons.Loader {
	if a.icons == nil {
		a.icons = icons.NewLoader(func(path string) []byte {
			p := content.Current
			if p == nil {
				return nil
			}
			f := p.Root().FindPath(path)
			if f == nil {
				return nil
			}
			data, err := p.ReadFile(f, false)
			if err != nil {
				return nil
			}
			return data
		})
	}
	return a.icons
}

// releaseIcons frees the icon textures (ClientData::clear).
func (a *App) releaseIcons() {
	for _, t := range a.textures {
		t.Release()
	}
	a.textures = nil
	a.icons = nil
}

// sheetTexture uploads a sheet once.
func (a *App) sheetTexture(s *icons.Sheet) *backend.Texture {
	if t, ok := s.Tex.(*backend.Texture); ok {
		return t
	}
	rgba := &image.RGBA{Pix: s.Image.Pix, Stride: s.Image.Stride, Rect: s.Image.Rect}
	t := backend.NewTextureFromRgba(rgba)
	s.Tex = t
	a.textures = append(a.textures, t)
	return t
}

// drawIcon draws icon id of set at size; it reserves the space when the
// icon is missing so that rows stay aligned.
func (a *App) drawIcon(set *icons.Set, id int, size float32) {
	sh, r, ok := set.Find(id)
	if !ok {
		imgui.Dummy(imgui.NewVec2(size, size))
		return
	}
	t := a.sheetTexture(sh)
	w, h := float32(sh.Image.Rect.Dx()), float32(sh.Image.Rect.Dy())
	imgui.ImageV(t.ID, imgui.NewVec2(size, size),
		imgui.NewVec2(float32(r.Min.X)/w, float32(r.Min.Y)/h), imgui.NewVec2(float32(r.Max.X)/w, float32(r.Max.Y)/h))
}
