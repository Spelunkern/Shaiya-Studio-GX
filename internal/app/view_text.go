package app

import (
	"strings"
	"unicode/utf16"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/codepage"
)

// textView is a plain text document (TextView): .txt, .cfg, .csv, .ini,
// .sql, .xml and .lua files, or a text built by another view.
type textView struct {
	file     *archive.File
	text     string
	cp       int
	editable bool // m_readOnly in shStudio: the text comes from the file and can be saved
	modified bool
	status   string
	ready    bool
	title    string
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		switch ext {
		case ".txt", ".cfg", ".csv", ".ini", ".sql", ".xml", ".lua":
			return &textView{file: f, editable: true}
		}
		return nil
	})
}

// newTextDoc returns a read-only text window (StdTextView / TextView with
// a given text).
func newTextDoc(title, text string) *textView {
	f := archive.NewFile(nil, title, 0)
	f.Path = "\x00text:" + title
	return &textView{file: f, text: text, title: title, ready: true}
}

func (v *textView) Path() string { return v.file.Path }

func (v *textView) Title() string {
	t := v.file.Name
	if v.title != "" {
		t = v.title
	}
	if v.modified {
		t += " *"
	}
	return t
}

func (v *textView) setReady() { v.ready = true }

// load reads the file (TextView::loadText): UTF-16 with BOM, UTF-8 with BOM,
// or the code page of the file (asked when unknown). Unix line ends are
// converted to CR LF.
func (v *textView) load(a *App, done func(err error)) {
	if !v.editable {
		done(nil)
		return
	}
	data, err := readFile(v.file.Path, false)
	if err != nil {
		done(err)
		return
	}
	if len(data) < 2 {
		done(errEmpty)
		return
	}
	if data[0] == 0xFF && data[1] == 0xFE {
		v.cp = codepage.UTF16LE
		v.file.CodePage = v.cp
		v.text = trimNul(codepage.Decode(data[2:], codepage.UTF16LE))
		done(nil)
		return
	}
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
		v.cp = codepage.UTF8
		v.file.CodePage = v.cp
	}
	finish := func(cp int) {
		if cp == -1 {
			done(errAborted)
			return
		}
		v.cp = cp
		v.text = codepage.Decode(data, cp)
		if !strings.Contains(v.text, "\r") {
			v.text = strings.ReplaceAll(v.text, "\n", "\r\n")
			v.modified = true
			v.status = "Unix EOL converted to Windows EOL"
		}
		done(nil)
	}
	if v.cp == 0 {
		v.cp = v.file.CodePage
	}
	if v.cp == 0 {
		a.chooseCodepage(v.file.Name, finish)
		return
	}
	finish(v.cp)
}

func trimNul(s string) string { return strings.TrimRight(s, "\x00") }

// save writes the text back in its code page (TextView::saveText).
func (v *textView) save(a *App) bool {
	if !v.editable {
		return false
	}
	var out []byte
	switch v.cp {
	case codepage.UTF16LE:
		u := utf16.Encode([]rune(v.text))
		out = append(out, 0xFF, 0xFE)
		for _, c := range u {
			out = append(out, byte(c), byte(c>>8))
		}
	case codepage.UTF8:
		out = append([]byte{0xEF, 0xBB, 0xBF}, v.text...)
	default:
		out = codepage.Encode(v.text, v.cp)
	}
	if err := writeFile(v.file.Path, out, false); err != nil {
		a.alert("The application failed to save the document "+v.file.Name+".\nThe error was: "+err.Error(), iconError, nil)
		return false
	}
	v.modified = false
	return true
}

func (v *textView) Close(a *App) bool {
	if !v.modified || !v.editable {
		v.file.State = 0
		return true
	}
	a.askSave(v.file.Path, func(answer int) {
		if answer == answerCancel || answer == answerYes && !v.save(a) {
			return
		}
		v.modified = false
		v.file.State = 0
		a.closeDocument(v)
	})
	return false
}

func (v *textView) Draw(a *App) bool {
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
			imgui.Separator()
			if imgui.MenuItemBoolV("Save", "Ctrl+S", false, v.editable && v.modified) {
				v.save(a)
			}
			imgui.EndMenu()
		}
		imgui.EndMenuBar()
	}
	if imgui.IsWindowFocusedV(imgui.FocusedFlagsChildWindows) {
		io := imgui.CurrentIO()
		if io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyS) && v.editable && v.modified {
			v.save(a)
		}
		if io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyW) {
			keep = false
		}
	}
	h := imgui.ContentRegionAvail().Y - imgui.FrameHeightWithSpacing()
	flags := imgui.InputTextFlagsAllowTabInput
	if !v.editable {
		flags |= imgui.InputTextFlagsReadOnly
	}
	// ImGui edits LF text; CR LF is restored when saving.
	text := strings.ReplaceAll(v.text, "\r\n", "\n")
	if imgui.InputTextMultiline("##text", &text, imgui.NewVec2(-1, h), flags, nil) {
		v.text = strings.ReplaceAll(text, "\n", "\r\n")
		v.modified = true
	}
	imgui.TextUnformatted(v.status)
	return keep
}
