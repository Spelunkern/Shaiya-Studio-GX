package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/crypto/seed"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/filedlg"
)

// tableData is an SData table (Serializable in shStudio).
type tableData interface {
	Read(data []byte, cp int, f sdata.Format) error
	Write(cp int, f sdata.Format) ([]byte, error)
}

// sdataView is the part of an SData document that depends on the table:
// its lists, menus and exports (the SDataView subclasses).
type sdataView interface {
	// loaded is called once the data is read (and after a merge).
	loaded(d *sdataDoc)
	// draw draws the content below the menu bar.
	draw(d *sdataDoc)
	// editMenu and toolsMenu draw the view's items of these menus.
	editMenu(d *sdataDoc)
	toolsMenu(d *sdataDoc)
	// export returns the table as CSV ("csv ") or SQL ("sql ").
	export(format string) (string, bool)
}

// sdataDoc is a document showing an SData table (Doc + DocView +
// SDataView).
type sdataDoc struct {
	file     *archive.File
	data     tableData
	view     sdataView
	format   sdata.Format // ftCodepage: the format the file was read in
	codePage int
	// langFile is the language file merged into the data (m_fileName).
	langFile string
	// options.
	hasTranslation bool // DocView::hasTranslation
	detectCP       bool // DocView::detectsCodepage
	decrypt        bool // m_decrypt
	// prefix of the DB text files ("BinarySData\\DBItemText_").
	textPrefix string

	modified bool
	status   string
	ready    bool
	failed   bool
}

func (d *sdataDoc) Path() string { return d.file.Path }

func (d *sdataDoc) Title() string {
	t := d.file.Name
	if d.modified {
		t += " *"
	}
	return t
}

func (d *sdataDoc) setModified(on bool) {
	d.modified = on
	if on {
		d.file.State |= archive.StateDirty
	} else {
		d.file.State &^= archive.StateDirty
	}
}

// readFile reads a file of the workspace or of the disk
// (StudioFrame::loadFromFile).
func readFile(path string, decrypt bool) ([]byte, error) {
	p := content.Current
	if p != nil {
		if f := p.Root().FindPath(path); f != nil {
			return p.ReadFile(f, decrypt)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || !decrypt {
		return data, err
	}
	plain, _, _, err := seed.DecryptFile(data)
	return plain, err
}

// writeFile stores data (StudioFrame::saveToFile): files that were
// encrypted are encrypted again.
func writeFile(path string, data []byte, encrypt bool) error {
	p := content.Current
	var f *archive.File
	if p != nil {
		f = p.Root().FindPath(path)
	}
	if encrypt {
		if f != nil {
			encrypt = seed.IsEncrypted(f.Key[:])
		} else {
			encrypt = fileIsEncrypted(path)
		}
	}
	if p != nil && f != nil {
		return p.WriteFile(f, data, encrypt)
	}
	if encrypt {
		var key [64]byte
		out, err := content.EncryptDataFile(data, &key)
		if err != nil {
			return err
		}
		return os.WriteFile(path, out, 0o644)
	}
	return os.WriteFile(path, data, 0o644)
}

func fileIsEncrypted(path string) bool {
	fh, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fh.Close()
	st, err := fh.Stat()
	if err != nil || st.Size() < 0x40 || st.Size()&0xF != 0 {
		return false
	}
	var h [0x40]byte
	if _, err := fh.Read(h[:]); err != nil {
		return false
	}
	return seed.IsEncrypted(h[:])
}

// fileExists checks a workspace path that may hold a '*' pattern
// (StudioFrame::resolvePath); it returns the resolved path.
func fileExists(path string) (string, bool) {
	p := content.Current
	if p != nil {
		if !strings.Contains(path, "*") {
			return path, p.Root().FindPath(path) != nil
		}
		dir, spec := path[:strings.LastIndex(path, `\`)+1], path[strings.LastIndex(path, `\`)+1:]
		if folder := p.Root().FindFolderPath(dir); folder != nil {
			for _, f := range folder.Files {
				if ok, _ := filepath.Match(strings.ToLower(spec), strings.ToLower(f.Name)); ok {
					return dir + f.Name, true
				}
			}
			return path, false
		}
	}
	matches, _ := filepath.Glob(path)
	if len(matches) > 0 {
		return matches[0], true
	}
	_, err := os.Stat(path)
	return path, err == nil
}

// load reads the document (DocView::load). It may ask for a code page;
// done receives the result.
func (d *sdataDoc) load(a *App, done func(err error)) {
	data, err := readFile(d.file.Path, d.decrypt)
	if err != nil {
		done(err)
		return
	}
	dir := d.file.Path[:strings.LastIndexAny(d.file.Path, `\/`)+1]
	name := d.file.Name
	format := sdata.File
	langPath := ""
	switch {
	case strings.HasPrefix(name, "DB"):
		format = sdata.Flnt
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if len(base) >= 4 {
			base = base[:len(base)-4]
		}
		langPath = dir + base + "Text_"
	case strings.HasPrefix(name, "NpcQuest"):
		langPath = dir + strings.TrimSuffix(name, filepath.Ext(name)) + "Trans_"
		if _, ok := fileExists(langPath + "*.sdata"); ok {
			format = sdata.Flnt
		}
	}
	language := prefs.Default.Int("SData", "language", 840)
	if format != sdata.File && d.hasTranslation {
		suffix := sdata.CountryCode(language)
		if strings.HasPrefix(name, "NpcQuest") {
			suffix = sdata.CountryName(language)
		}
		lf := langPath + suffix + ".SData"
		if _, ok := fileExists(lf); !ok {
			a.alert("The file "+baseName(lf)+" which is required to load the file "+name+
				" does not exist. You must choose another language in the settings or have that file in the repository and reopen it to finalize the action.", iconError, nil)
			done(errors.New("language file missing"))
			return
		}
		langPath = lf
	}
	d.format = format
	finish := func(cp int) {
		if err := d.data.Read(data, cp, format); err != nil {
			done(err)
			return
		}
		if format == sdata.Flnt && d.hasTranslation {
			text, err := readFile(langPath, false)
			if err != nil {
				a.alert("The file "+langPath+" cannot be read.\nVerify it is not locked by another process and try again.", iconError, nil)
			} else if err := d.data.Read(text, cp, sdata.Ltxt); err != nil {
				a.alert(err.Error(), iconError, nil)
			} else {
				d.langFile = langPath
			}
		}
		d.codePage = cp
		d.file.CodePage = cp
		done(nil)
	}
	cp := d.file.CodePage
	if !d.detectCP {
		finish(cp)
		return
	}
	if cp == 0 {
		detect := true
		if format == sdata.Flnt {
			switch language {
			case 410:
				cp, detect = 949, false
			case 156:
				cp, detect = 936, false
			case 250, 276, 380, 840:
				cp, detect = 1252, false
			case 616, 792:
				cp, detect = 1250, false
			}
		}
		if detect {
			a.chooseCodepage(d.file.Name, func(c int) {
				if c <= 0 {
					done(errors.New("aborted"))
					return
				}
				finish(c)
			})
			return
		}
	} else if cp < 0 {
		done(errors.New("aborted"))
		return
	}
	finish(cp)
}

// store writes the document back (DocView::store).
func (d *sdataDoc) store(a *App) error {
	cp := 0
	if p := content.Current; p != nil {
		cp = p.CodePage()
	}
	if cp <= 0 {
		cp = d.file.CodePage
	}
	if cp <= 0 {
		cp = d.codePage
	}
	out, err := d.data.Write(cp, d.format)
	if err != nil {
		return err
	}
	if err := writeFile(d.file.Path, out, d.decrypt); err != nil {
		return err
	}
	if d.format == sdata.Flnt && d.langFile != "" {
		text, err := d.data.Write(cp, sdata.Ltxt)
		if err != nil {
			return err
		}
		return writeFile(d.langFile, text, false)
	}
	return nil
}

// save is Doc::save.
func (d *sdataDoc) save(a *App) bool {
	if err := d.store(a); err != nil {
		a.alert("The application failed to save the document "+d.file.Name+".\nThe error was: "+err.Error(), iconError, nil)
		return false
	}
	d.setModified(false)
	return true
}

// Close asks to save a modified document (Doc::onClose).
func (d *sdataDoc) Close(a *App) bool {
	if !d.modified {
		d.file.State = 0
		return true
	}
	a.askSave(d.file.Path, func(answer int) {
		switch answer {
		case answerYes:
			if !d.save(a) {
				return
			}
		case answerCancel:
			return
		}
		d.modified = false
		d.file.State = 0
		a.closeDocument(d)
	})
	return false
}

// Draw draws the document window content.
func (d *sdataDoc) Draw(a *App) bool {
	if !d.ready {
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
			if imgui.MenuItemBoolV("Save", "Ctrl+S", false, d.modified) {
				d.save(a)
			}
			if imgui.MenuItemBoolV("Export...", "Ctrl+E", false, true) {
				d.exportFile(a)
			}
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Edit") {
			imgui.MenuItemBoolV("Undo", "Ctrl+Z", false, false)
			imgui.MenuItemBoolV("Redo", "Ctrl+Y", false, false)
			imgui.Separator()
			d.view.editMenu(d)
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Tools") {
			d.view.toolsMenu(d)
			if imgui.BeginMenu("Scripts") {
				if imgui.MenuItemBool("Import text fields from a language file") {
					d.mergeLanguageFile(a)
				}
				imgui.EndMenu()
			}
			imgui.EndMenu()
		}
		imgui.EndMenuBar()
	}
	if imgui.IsWindowFocusedV(imgui.FocusedFlagsChildWindows) {
		io := imgui.CurrentIO()
		if io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyS) && d.modified {
			d.save(a)
		}
		if io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyE) {
			d.exportFile(a)
		}
		if io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyW) {
			keep = false
		}
	}
	h := imgui.ContentRegionAvail().Y - imgui.FrameHeightWithSpacing()
	if imgui.BeginChildStrV("##content", imgui.NewVec2(0, h), 0, 0) {
		d.view.draw(d)
	}
	imgui.EndChild()
	imgui.Separator()
	imgui.TextUnformatted(d.status)
	return keep
}

// exportFile saves the table as CSV or SQL (SDataView::exportFile).
func (d *sdataDoc) exportFile(a *App) {
	name := strings.TrimSuffix(d.file.Name, filepath.Ext(d.file.Name))
	dlg := filedlg.New("Export", filedlg.SaveFile, name+".csv",
		filedlg.Filter{Name: "Comma-Separated Values File", Patterns: []string{"*.csv"}},
		filedlg.Filter{Name: "Structured Query Language File", Patterns: []string{"*.sql"}})
	a.pickFile(dlg, func(path string) {
		var format string
		switch strings.ToLower(filepath.Ext(path)) {
		case ".csv":
			format = sdata.TextCSV
		case ".sql":
			format = sdata.TextSQL
		default:
			return
		}
		text, ok := d.view.export(format)
		if !ok {
			return
		}
		// UTF-16LE with a BOM, as written by shStudio.
		u := utf16.Encode([]rune(text))
		out := make([]byte, 2, 2+2*len(u))
		out[0], out[1] = 0xFF, 0xFE
		for _, c := range u {
			out = append(out, byte(c), byte(c>>8))
		}
		if len(u) == 0 {
			out = nil
		}
		if err := os.WriteFile(path, out, 0o644); err == nil {
			d.status = "Data successfully exported."
		}
	})
}

// mergeLanguageFile loads a language file into the data
// (SDataView::mergeLanguageFile).
func (d *sdataDoc) mergeLanguageFile(a *App) {
	dlg := filedlg.New("Select a language file to merge", filedlg.OpenFile, "",
		filedlg.Filter{Name: "SData File", Patterns: []string{"*.sdata"}})
	a.pickFile(dlg, func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			a.alert("The selected file cannot be loaded.", iconError, nil)
			return
		}
		d.status = "Updating text fields..."
		if err := d.data.Read(data, d.codePage, sdata.Ltxt); err != nil {
			a.alert("The selected file is not a valid names file.\n"+err.Error(), iconError, nil)
		} else {
			d.langFile = path
		}
		d.view.loaded(d)
		d.status = ""
	})
}

func (d *sdataDoc) setReady() {
	d.ready = true
	d.view.loaded(d)
}

func itoa(i int) string { return sdata.FormatInt(i) }

func pad2(i int) string {
	s := itoa(i)
	if len(s) < 2 {
		s = "0" + s
	}
	return s
}
