package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
)

// Document is an open file window (Doc / Document in shStudio). Each one is
// an ImGui window that can be docked or dragged out of the main window.
type Document interface {
	// Path identifies the document (findDocument).
	Path() string
	// Title is the window title.
	Title() string
	// Draw draws the window content, including its menu bar; it returns
	// false when the document asks to close.
	Draw(a *App) bool
	// Close is called when the window is closed; returning false keeps it
	// open (unsaved changes).
	Close(a *App) bool
}

// viewFactory builds the view for a file (createView). It returns nil when
// the file is not supported.
type viewFactory func(a *App, ext string, f *archive.File, reload bool) Document

var viewFactories []viewFactory

// registerView adds a factory; views register themselves from their files.
func registerView(f viewFactory) { viewFactories = append(viewFactories, f) }

type openDoc struct {
	doc   Document
	open  bool
	focus bool
}

func (a *App) findDocument(path string) *openDoc {
	for _, d := range a.docs {
		if strings.EqualFold(d.doc.Path(), path) {
			return d
		}
	}
	return nil
}

// reopenDocument opens a tree file (Application::reopenDocument).
func (a *App) reopenDocument(f *archive.File) bool { return a.openDocument(f, f.Path, true) }

// openFile opens a disk file (Application::openFile).
func (a *App) openFile(path string) bool { return a.openDocument(nil, path, false) }

// openDocument is Application::openDocument.
func (a *App) openDocument(f *archive.File, path string, reload bool) bool {
	a.status = ""
	if path == "" {
		return false
	}
	if d := a.findDocument(path); d != nil {
		d.focus = true
		return false
	}
	name := baseName(path)
	ext := strings.ToLower(filepath.Ext(name))
	if f == nil {
		f = archive.NewFile(nil, name, 0)
		f.Path = path
	}
	var doc Document
	switch ext {
	case ".wav", ".mp3":
		openExternal(a, f)
		return true
	default:
		for _, factory := range viewFactories {
			if doc = factory(a, ext, f, reload); doc != nil {
				break
			}
		}
	}
	if doc == nil {
		a.status = `The file "` + name + `" is not supported.`
		return false
	}
	a.docs = append(a.docs, &openDoc{doc: doc, open: true, focus: true})
	return true
}

// openExternal hands a sound file to the system player (ShellExecute
// "open"). Archive files are copied to a temporary file first.
func openExternal(a *App, f *archive.File) {
	path := f.Path
	if p := content.Current; p != nil && p.IsArchive() {
		data, err := p.ReadFile(f, false)
		if err != nil {
			a.status = err.Error()
			return
		}
		path = filepath.Join(os.TempDir(), f.Name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			a.status = err.Error()
			return
		}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		a.status = err.Error()
	}
}

func baseName(p string) string {
	i := strings.LastIndexAny(p, `\/`)
	return p[i+1:]
}

// drawDocs draws every open document window.
func (a *App) drawDocs() {
	kept := a.docs[:0]
	for _, d := range a.docs {
		if d.focus {
			imgui.SetNextWindowFocus()
			d.focus = false
		}
		imgui.SetNextWindowSizeV(imgui.NewVec2(800, 560), imgui.CondFirstUseEver)
		open := true
		if imgui.BeginV(d.doc.Title()+"###"+d.doc.Path(), &open, imgui.WindowFlagsMenuBar) {
			if !d.doc.Draw(a) {
				open = false
			}
		}
		imgui.End()
		if !open && !d.doc.Close(a) {
			open = true
		}
		if open {
			kept = append(kept, d)
		}
	}
	a.docs = kept
}

// closeAll closes every document (Application::closeAll then
// StudioFrame::closeAll); it returns false when a document refuses.
func (a *App) closeAll() bool {
	kept := a.docs[:0]
	for _, d := range a.docs {
		if !d.doc.Close(a) {
			kept = append(kept, d)
		}
	}
	a.docs = kept
	if len(a.docs) > 0 {
		return false
	}
	a.closeClient()
	return true
}
