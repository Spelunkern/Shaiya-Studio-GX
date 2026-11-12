package app

import (
	"image"
	"image/png"
	"os"
	"strconv"

	"github.com/go-gl/gl/v2.1/gl"
)

// Debug capture: SHSTUDIOGX_SCREENSHOT=file.png saves the main window after
// SHSTUDIOGX_FRAMES frames (default 30) and exits. SHSTUDIOGX_OPEN opens a
// workspace or archive at startup (with SHSTUDIOGX_CP as code page) and
// SHSTUDIOGX_DOC a file of it. Used to check the UI without a person at the
// screen.
var (
	shotPath   = os.Getenv("SHSTUDIOGX_SCREENSHOT")
	shotFrames = envInt("SHSTUDIOGX_FRAMES", 30)
	shotCount  int
	glReady    bool
)

func envInt(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil {
		return v
	}
	return def
}

func (a *App) afterRender() {
	if shotPath == "" {
		return
	}
	shotCount++
	if shotCount < shotFrames {
		return
	}
	if !glReady {
		if err := gl.Init(); err != nil {
			os.Exit(2)
		}
		glReady = true
	}
	w, h := a.be.DisplaySize()
	pix := make([]byte, int(w)*int(h)*4)
	gl.ReadBuffer(gl.BACK)
	gl.PixelStorei(gl.PACK_ALIGNMENT, 1)
	gl.ReadPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pix))
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	stride := int(w) * 4
	for y := 0; y < int(h); y++ {
		copy(img.Pix[y*stride:(y+1)*stride], pix[(int(h)-1-y)*stride:])
	}
	if f, err := os.Create(shotPath); err == nil {
		png.Encode(f, img)
		f.Close()
	}
	os.Exit(0)
}

// debugOpen opens SHSTUDIOGX_OPEN at startup.
func (a *App) debugOpen() {
	path := os.Getenv("SHSTUDIOGX_OPEN")
	if path == "" {
		return
	}
	a.openClient(path, envInt("SHSTUDIOGX_CP", 1252), true)
	if doc := os.Getenv("SHSTUDIOGX_DOC"); doc != "" {
		a.afterOpen = func() {
			if f := a.client.root.FindPath(doc); f != nil {
				a.reopenDocument(f)
			} else {
				a.status = "debug: " + doc + " not found"
			}
		}
	}
}
