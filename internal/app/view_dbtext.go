package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
)

// dbTextView edits a DB*Text_<LANG>.SData table cell by cell (DBTextView).
type dbTextView struct {
	app      *App
	doc      *sdataDoc
	table    *sdata.DBTextTable
	selected map[int]bool
	focus    int
	field    int // current column (m_curField)
	editRow  int
	editCol  int
	editText string
	editNew  bool
	find     string
	findOn   bool
	findMsg  string
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		if ok, _ := filepath.Match("db*text_*.sdata", strings.ToLower(f.Name)); !ok {
			return nil
		}
		t := sdata.NewDBTextTable()
		t.SetCodePage(sdata.DBTextCodePage(f.Name))
		v := &dbTextView{app: a, table: t, selected: map[int]bool{}, focus: -1, editRow: -1}
		d := &sdataDoc{file: f, data: t, view: v}
		v.doc = d
		return d
	})
}

func (v *dbTextView) loaded(d *sdataDoc) {
	v.field = 0
	if v.table.KeyCount() < v.table.FieldCount() {
		v.field = v.table.KeyCount()
	}
	d.status = v.table.StatusText()
}

func (v *dbTextView) beginEdit(row, col int) {
	if row < 0 || row >= v.table.RowCount() || col < 0 || col >= v.table.FieldCount() {
		return
	}
	v.endEdit(true, 0)
	v.field = col
	v.editRow, v.editCol = row, col
	v.editText = v.table.GetText(row, col)
	v.editNew = true
}

// endEdit stores the edited cell (DBTextView::endEdit); move goes to the
// previous or next column.
func (v *dbTextView) endEdit(commit bool, move int) {
	if v.editRow < 0 {
		return
	}
	row, col := v.editRow, v.editCol
	v.editRow = -1
	if commit {
		text := v.editText
		orig := v.table.GetText(row, col)
		if !strings.Contains(orig, "\r\n") && strings.Contains(orig, "\n") {
			text = strings.ReplaceAll(text, "\r", "")
		} else {
			text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\n", "\r\n")
		}
		switch v.table.SetText(row, col, text) {
		case sdata.SetChanged:
			v.doc.setModified(true)
		case sdata.SetBadNumber:
			msg := "This column holds an integer: enter a decimal number."
			if sdata.IsKeyField(v.table, col) {
				msg = "This column holds an integer key: enter a decimal number."
			}
			v.app.alert(msg, iconError, nil)
		case sdata.SetNotEncodable:
			v.app.alert(v.table.NotEncodableText(), iconError, nil)
		}
	}
	if move != 0 {
		v.beginEdit(row, col+move)
	}
}

func (v *dbTextView) selectRow(r int) {
	v.selected = map[int]bool{r: true}
	v.focus = r
}

func (v *dbTextView) insertRow() {
	at := v.table.RowCount()
	if v.focus >= 0 {
		at = v.focus + 1
	}
	if at = v.table.InsertRow(at); at >= 0 {
		v.doc.setModified(true)
		v.selectRow(at)
		v.doc.status = v.table.StatusText()
	}
}

func (v *dbTextView) deleteSelected() {
	var rows []int
	for r := range v.selected {
		rows = append(rows, r)
	}
	if len(rows) == 0 {
		return
	}
	msg := "Delete the selected row?"
	if len(rows) > 1 {
		msg = fmt.Sprintf("Delete the %d selected rows?", len(rows))
	}
	v.app.confirm(msg, func(yes bool) {
		if !yes {
			return
		}
		first := -1
		for i := v.table.RowCount() - 1; i >= 0; i-- {
			if v.selected[i] {
				v.table.DeleteRow(i)
				first = i
			}
		}
		v.selected = map[int]bool{}
		v.doc.setModified(true)
		if first >= v.table.RowCount() {
			first = v.table.RowCount() - 1
		}
		if first >= 0 {
			v.selectRow(first)
		}
		v.doc.status = v.table.StatusText()
	})
}

func (v *dbTextView) eraseSelected() {
	changed := false
	for r := range v.selected {
		for f := v.table.KeyCount(); f < v.table.FieldCount(); f++ {
			if !v.table.IsNumericField(f) && v.table.SetText(r, f, "") == sdata.SetChanged {
				changed = true
			}
		}
	}
	if changed {
		v.doc.setModified(true)
	}
}

func (v *dbTextView) copySelected() {
	var b strings.Builder
	for r := 0; r < v.table.RowCount(); r++ {
		if !v.selected[r] {
			continue
		}
		for f := 0; f < v.table.FieldCount(); f++ {
			if f > 0 {
				b.WriteByte('\t')
			}
			b.WriteString(strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(v.table.GetText(r, f)))
		}
		b.WriteString("\r\n")
	}
	if b.Len() > 0 {
		imgui.SetClipboardText(b.String())
	}
}

func (v *dbTextView) findNext() {
	needle := strings.ToLower(v.find)
	n := v.table.RowCount()
	if needle == "" || n == 0 {
		return
	}
	for step := 1; step <= n; step++ {
		row := ((v.focus+step)%n + n) % n
		for f := 0; f < v.table.FieldCount(); f++ {
			if strings.Contains(strings.ToLower(v.table.GetText(row, f)), needle) {
				v.field = f
				v.selectRow(row)
				v.findMsg = ""
				return
			}
		}
	}
	v.findMsg = "Not found"
}

func (v *dbTextView) draw(d *sdataDoc) {
	t := v.table
	n := t.FieldCount()
	focused := imgui.IsWindowFocusedV(imgui.FocusedFlagsChildWindows)
	if focused && v.editRow < 0 {
		io := imgui.CurrentIO()
		switch {
		case io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyF):
			v.findOn = true
		case io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyG), imgui.IsKeyPressedBool(imgui.KeyF3):
			v.findNext()
		case imgui.IsKeyPressedBool(imgui.KeyF2), imgui.IsKeyPressedBool(imgui.KeyEnter):
			v.beginEdit(v.focus, v.field)
		case imgui.IsKeyPressedBool(imgui.KeyDelete) && !io.KeyShift():
			v.deleteSelected()
		case imgui.IsKeyPressedBool(imgui.KeyInsert):
			v.insertRow()
		case io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyLeftArrow) && v.field > 0:
			v.field--
		case io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyRightArrow) && v.field+1 < n:
			v.field++
		case io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyC):
			v.copySelected()
		case io.KeyCtrl() && imgui.IsKeyPressedBool(imgui.KeyA):
			for r := 0; r < t.RowCount(); r++ {
				v.selected[r] = true
			}
		case imgui.IsKeyPressedBool(imgui.KeyUpArrow) && v.focus > 0:
			v.selectRow(v.focus - 1)
		case imgui.IsKeyPressedBool(imgui.KeyDownArrow) && v.focus < t.RowCount()-1:
			v.selectRow(v.focus + 1)
		}
	}
	h := float32(0)
	if v.findOn {
		h = -imgui.FrameHeightWithSpacing()
	}
	flags := imgui.TableFlagsBorders | imgui.TableFlagsRowBg | imgui.TableFlagsScrollY | imgui.TableFlagsScrollX |
		imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit
	if n > 0 && imgui.BeginTableV("dbtext", int32(n), flags, imgui.NewVec2(0, h), 0) {
		imgui.TableSetupScrollFreeze(0, 1)
		for f := 0; f < n; f++ {
			w := float32(220)
			name := t.FieldName(f)
			switch {
			case t.IsNumericField(f):
				w = max(80, imgui.CalcTextSize(name).X+24)
			case strings.EqualFold(name, "text"):
				w = 420
			}
			imgui.TableSetupColumnV(name, imgui.TableColumnFlagsWidthFixed, w, 0)
		}
		imgui.TableHeadersRow()
		clip := imgui.NewListClipper()
		clip.Begin(int32(t.RowCount()))
		for clip.Step() {
			for r := int(clip.DisplayStart()); r < int(clip.DisplayEnd()); r++ {
				imgui.TableNextRow()
				imgui.PushIDInt(int32(r))
				for f := 0; f < n; f++ {
					imgui.TableSetColumnIndex(int32(f))
					if r == v.editRow && f == v.editCol {
						imgui.SetNextItemWidth(-1)
						if v.editNew {
							imgui.SetKeyboardFocusHere()
							v.editNew = false
						}
						imgui.InputTextWithHint("##edit", "", &v.editText, 0, nil)
						switch {
						case imgui.IsKeyPressedBool(imgui.KeyEscape):
							v.endEdit(false, 0)
						case imgui.IsKeyPressedBool(imgui.KeyTab):
							move := 1
							if imgui.CurrentIO().KeyShift() {
								move = -1
							}
							v.endEdit(true, move)
						case imgui.IsKeyPressedBool(imgui.KeyEnter) || imgui.IsItemDeactivated():
							v.endEdit(true, 0)
						}
						continue
					}
					text := strings.NewReplacer("\r", " ", "\t", " ", "\n", "¶").Replace(t.GetText(r, f))
					if f == 0 {
						if imgui.SelectableBoolV(text+"##r", v.selected[r], imgui.SelectableFlagsSpanAllColumns|imgui.SelectableFlagsAllowDoubleClick, imgui.NewVec2(0, 0)) {
							if imgui.CurrentIO().KeyCtrl() {
								v.selected[r] = !v.selected[r]
								v.focus = r
							} else {
								v.selectRow(r)
							}
						}
					} else {
						if t.IsNumericField(f) {
							if off := imgui.ContentRegionAvail().X - imgui.CalcTextSize(text).X; off > 0 {
								imgui.SetCursorPosX(imgui.CursorPosX() + off)
							}
						}
						imgui.TextUnformatted(text)
					}
					if imgui.IsItemHovered() && imgui.IsMouseClickedBool(imgui.MouseButtonLeft) {
						v.field = f
					}
					if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
						v.selectRow(r)
						v.beginEdit(r, f)
					}
				}
				imgui.PopID()
			}
		}
		clip.End()
		clip.Destroy()
		imgui.EndTable()
	}
	if v.findOn {
		imgui.SetNextItemWidth(240)
		if imgui.InputTextWithHint("##find", "Find", &v.find, imgui.InputTextFlagsEnterReturnsTrue, nil) {
			v.findNext()
		}
		imgui.SameLine()
		imgui.TextUnformatted(v.findMsg)
		imgui.SameLine()
		if imgui.Button("x") || imgui.IsKeyPressedBool(imgui.KeyEscape) {
			v.findOn = false
		}
	}
}

func (v *dbTextView) editMenu(d *sdataDoc) {
	has := len(v.selected) > 0
	if imgui.MenuItemBoolV("Copy", "Ctrl+C", false, has) {
		v.copySelected()
	}
	if imgui.MenuItemBoolV("Select All", "Ctrl+A", false, true) {
		for r := 0; r < v.table.RowCount(); r++ {
			v.selected[r] = true
		}
	}
	imgui.Separator()
	if imgui.MenuItemBoolV("Find …", "Ctrl+F", false, true) {
		v.findOn = true
	}
	if imgui.MenuItemBoolV("Find Next", "Ctrl+G", false, v.find != "") {
		v.findNext()
	}
	imgui.Separator()
	if imgui.MenuItemBoolV("Create item", "Ins", false, true) {
		v.insertRow()
	}
	if imgui.MenuItemBoolV("Delete item", "Del", false, has) {
		v.deleteSelected()
	}
	if imgui.MenuItemBoolV("Erase item", "Shift+Del", false, has) {
		v.eraseSelected()
	}
}

func (v *dbTextView) toolsMenu(d *sdataDoc) {}

func (v *dbTextView) export(format string) (string, bool) { return "", false }
