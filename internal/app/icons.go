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

// iconTexRef returns the texture of the sheet holding id (or of the first
// sheet) for image buttons.
func (a *App) iconTexRef(set *icons.Set, id int) imgui.TextureRef {
	sh, _, ok := set.Find(id)
	if !ok {
		if set.Empty() {
			return imgui.TextureRef{}
		}
		sh = set.Sheets[0]
	}
	return a.sheetTexture(sh).ID
}

// iconUV returns the top-left (or bottom-right) UV of id; a missing icon
// maps to an empty area.
func (a *App) iconUV(set *icons.Set, id int, max bool) imgui.Vec2 {
	sh, r, ok := set.Find(id)
	if !ok {
		return imgui.NewVec2(0, 0)
	}
	w, h := float32(sh.Image.Rect.Dx()), float32(sh.Image.Rect.Dy())
	if max {
		return imgui.NewVec2(float32(r.Max.X)/w, float32(r.Max.Y)/h)
	}
	return imgui.NewVec2(float32(r.Min.X)/w, float32(r.Min.Y)/h)
}

// pickIcon draws a grid of the icons of set (the IconPicker drop-down);
// it returns the clicked id.
func (a *App) pickIcon(set *icons.Set) (int, bool) {
	n := 0
	for _, sh := range set.Sheets {
		for i := 0; i < sh.Count(); i++ {
			id := sh.FirstID + i
			if _, ok := sh.Rect(id); !ok {
				continue
			}
			if n%16 != 0 {
				imgui.SameLine()
			}
			n++
			imgui.PushIDInt(int32(id))
			clicked := imgui.ImageButtonV("##i", a.sheetTexture(sh).ID, imgui.NewVec2(32, 32),
				a.iconUV(set, id, false), a.iconUV(set, id, true), imgui.NewVec4(0, 0, 0, 0), imgui.NewVec4(1, 1, 1, 1))
			imgui.PopID()
			if clicked {
				return id, true
			}
		}
	}
	return 0, false
}
