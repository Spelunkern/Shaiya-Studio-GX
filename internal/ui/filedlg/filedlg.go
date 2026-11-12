// Package filedlg is the file and folder picker of Shaiya Studio GX. It
// stands in for the Windows common dialogs (IFileDialog) that shStudio uses
// in FileDialogs and BrowseFoldersDialog, drawn with Dear ImGui so it is the
// same on every OS.
package filedlg

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/natural"
)

// Mode selects what the dialog picks.
type Mode int

const (
	OpenFile Mode = iota
	SaveFile
	SelectFolder
)

// Dialog is a modal picker. Create it with New, call Draw every frame until
// it returns done.
type Dialog struct {
	Title   string
	Mode    Mode
	Filters []Filter // for files; the first is the default
	// Accept, when set, validates a chosen folder (SelectFolder).
	Accept func(dir string) bool
	// Inner draws dialogs stacked above this one (inside its popup).
	Inner func()

	dir     string
	name    string
	filter  int
	entries []os.DirEntry
	err     string
	opened  bool
	result  string
	done    bool
	ok      bool
}

// Filter is a named list of patterns such as "*.saf".
type Filter struct {
	Name     string
	Patterns []string
}

// New returns a dialog starting at path (a file or folder).
func New(title string, mode Mode, path string, filters ...Filter) *Dialog {
	d := &Dialog{Title: title, Mode: mode, Filters: filters}
	d.dir, d.name = splitStart(path, mode)
	d.read()
	return d
}

func splitStart(path string, mode Mode) (dir, name string) {
	if path == "" {
		home, _ := os.UserHomeDir()
		return home, ""
	}
	path = filepath.FromSlash(strings.ReplaceAll(path, `\`, string(os.PathSeparator)))
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		return path, ""
	}
	dir = filepath.Dir(path)
	if mode != SelectFolder {
		name = filepath.Base(path)
	}
	if _, err := os.Stat(dir); err != nil {
		home, _ := os.UserHomeDir()
		dir = home
	}
	return dir, name
}

func (d *Dialog) read() {
	entries, err := os.ReadDir(d.dir)
	d.err = ""
	if err != nil {
		d.err = err.Error()
	}
	var keep []os.DirEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if !e.IsDir() && (d.Mode == SelectFolder || !d.matches(e.Name())) {
			continue
		}
		keep = append(keep, e)
	}
	sort.SliceStable(keep, func(i, j int) bool {
		if keep[i].IsDir() != keep[j].IsDir() {
			return keep[i].IsDir()
		}
		return natural.Compare(keep[i].Name(), keep[j].Name()) < 0
	})
	d.entries = keep
}

func (d *Dialog) matches(name string) bool {
	if len(d.Filters) == 0 {
		return true
	}
	for _, p := range d.Filters[d.filter].Patterns {
		if ok, _ := filepath.Match(strings.ToLower(p), strings.ToLower(name)); ok {
			return true
		}
	}
	return false
}

// Result returns the chosen path after Draw reported done with ok.
func (d *Dialog) Result() string { return d.result }

// Draw draws the dialog; it returns done=true once the user confirmed
// (ok=true) or cancelled.
func (d *Dialog) Draw() (done, ok bool) {
	if d.done {
		return true, d.ok
	}
	if !d.opened {
		imgui.OpenPopupStr(d.Title)
		d.opened = true
	}
	vp := imgui.MainViewport()
	imgui.SetNextWindowPosV(vp.Center(), imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
	imgui.SetNextWindowSizeV(imgui.NewVec2(640, 440), imgui.CondAppearing)
	if !imgui.BeginPopupModalV(d.Title, nil, 0) {
		return false, false
	}
	defer imgui.EndPopup()

	if imgui.Button("Up") {
		d.dir = filepath.Dir(d.dir)
		d.read()
	}
	imgui.SameLine()
	dir := d.dir
	imgui.SetNextItemWidth(-1)
	if imgui.InputTextWithHint("##dir", "folder", &dir, imgui.InputTextFlagsEnterReturnsTrue, nil) {
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			d.dir = dir
			d.read()
		}
	}
	h := imgui.ContentRegionAvail().Y - 2*imgui.FrameHeightWithSpacing()
	if imgui.BeginChildStrV("##list", imgui.NewVec2(0, h), imgui.ChildFlagsBorders, 0) {
		if d.err != "" {
			imgui.TextUnformatted(d.err)
		}
		for _, e := range d.entries {
			label := e.Name()
			if e.IsDir() {
				label = "[" + label + "]"
			}
			sel := !e.IsDir() && e.Name() == d.name
			if imgui.SelectableBoolV(label, sel, imgui.SelectableFlagsAllowDoubleClick, imgui.NewVec2(0, 0)) {
				if e.IsDir() {
					if imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
						d.dir = filepath.Join(d.dir, e.Name())
						d.read()
						break
					}
				} else {
					d.name = e.Name()
					if imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
						d.finish(filepath.Join(d.dir, d.name))
					}
				}
			}
		}
	}
	imgui.EndChild()

	if d.Mode != SelectFolder {
		imgui.SetNextItemWidth(imgui.ContentRegionAvail().X * 0.6)
		imgui.InputTextWithHint("##name", "file name", &d.name, 0, nil)
		if len(d.Filters) > 0 {
			imgui.SameLine()
			imgui.SetNextItemWidth(-1)
			if imgui.BeginCombo("##filter", d.Filters[d.filter].Name) {
				for i, f := range d.Filters {
					if imgui.SelectableBoolV(f.Name, i == d.filter, 0, imgui.NewVec2(0, 0)) {
						d.filter = i
						d.read()
					}
				}
				imgui.EndCombo()
			}
		}
	} else {
		imgui.TextUnformatted(d.dir)
	}
	label := "Open"
	switch d.Mode {
	case SaveFile:
		label = "Save"
	case SelectFolder:
		label = "Select Folder"
	}
	if imgui.Button(label) {
		switch d.Mode {
		case SelectFolder:
			if d.Accept == nil || d.Accept(d.dir) {
				d.finish(d.dir)
			}
		default:
			if d.name != "" {
				d.finish(filepath.Join(d.dir, d.name))
			}
		}
	}
	imgui.SameLine()
	if imgui.Button("Cancel") || imgui.IsKeyPressedBool(imgui.KeyEscape) {
		d.done, d.ok = true, false
		imgui.CloseCurrentPopup()
	}
	if d.Inner != nil {
		d.Inner()
	}
	if d.done {
		return true, d.ok
	}
	return false, false
}

func (d *Dialog) finish(path string) {
	d.result = path
	d.done, d.ok = true, true
	imgui.CloseCurrentPopup()
}
