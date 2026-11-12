// Package table is the list control of Shaiya Studio GX: the ImGui
// counterpart of shStudio's ListCtrl / ListCtrlOD (report-style lists of
// nodes with tagged columns, sorting, multi-selection, insert, delete,
// clear, copy and paste).
package table

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
)

// Column format flags (LCF_* and LVCFMT_* in shStudio): the low two bits
// align the text, the others select how the value is shown.
const (
	AlignLeft   = 0x0000
	AlignRight  = 0x0001
	AlignCenter = 0x0002
	Signed      = 0x0010 // getInt as %d
	Unsigned    = 0x0020 // getInt as %u
	Hex         = 0x0040 // getInt as 0x%X
	Float       = 0x0080 // getInt bits as a float
	None        = 0x0100 // nothing
	Bold        = 0x1000
	Owner       = 0x2000 // the list's own text (CellText)
)

// Column is a list column.
type Column struct {
	Title  string
	Width  float32
	Format int
	Tag    sdata.Tag
}

// C builds a column like ListCtrl::addColumn(title, width, format, id).
func C(title string, width int, format int, tag sdata.Tag) Column {
	return Column{Title: title, Width: float32(width), Format: format, Tag: tag}
}

// Node is a row of a list.
type Node interface {
	Get(tag sdata.Tag) int
	Text(tag sdata.Tag) string
}

// List is a report list. The owner fills the callbacks; unset ones disable
// the matching commands.
type List struct {
	ID      string
	Columns []Column

	// Rows returns the number of rows; Row returns one.
	Rows func() int
	Row  func(i int) Node
	// CellText returns the text of an Owner column.
	CellText func(n Node, tag sdata.Tag) string
	// Sort sorts on key1 then key2 and reports whether the order changed.
	Sort func(key1, key2 sdata.Tag, ascending bool) bool

	// Commands (ListCtrlOD). Insert returns the index of the new row or -1.
	Insert   func(index int) int
	Remove   func(index int) bool
	Clear    func(index int) bool
	MoveUp   func(index int) bool
	MoveDown func(index int) bool
	Activate func(index int)
	// Copy returns the clipboard text of rows; Paste adds or updates rows
	// from the internal clipboard and reports whether it changed anything.
	Copy  func(rows []int) string
	Paste func(selected []int) bool
	// CanPaste reports whether the internal clipboard holds rows of this
	// list.
	CanPaste func() bool
	// Changed is called after any change (onChanged).
	Changed func()
	// Confirm asks a yes/no question and calls then with the answer.
	Confirm func(text string, then func(yes bool))
	// ContextMenu draws extra context menu items (filters, columns, ...).
	ContextMenu func()
	// Icon draws the icon of a row in the first column, if any.
	Icon func(n Node, size float32)

	SortColumn     sdata.Tag
	PrevSortColumn sdata.Tag
	Ascending      bool
	SingleSel      bool

	selected map[int]bool
	anchor   int
	focus    int
	scrollTo int
}

// New returns a list sorted ascending.
func New(id string, cols ...Column) *List {
	return &List{ID: id, Columns: cols, Ascending: true, selected: map[int]bool{}, anchor: -1, focus: -1, scrollTo: -1}
}

// FormatValue formats a cell (ListCtrl::formatValue).
func (l *List) FormatValue(n Node, c Column) string {
	if c.Format&0xF000 == Owner {
		if l.CellText != nil {
			return l.CellText(n, c.Tag)
		}
		return n.Text(c.Tag)
	}
	switch c.Format & 0x01F0 {
	case 0:
		return n.Text(c.Tag)
	case None:
		return ""
	case Signed:
		return strconv.Itoa(int(int32(n.Get(c.Tag))))
	case Unsigned:
		return strconv.FormatUint(uint64(uint32(n.Get(c.Tag))), 10)
	case Hex:
		return fmt.Sprintf("0x%X", uint32(n.Get(c.Tag)))
	case Float:
		v := math.Float32frombits(uint32(n.Get(c.Tag)))
		m := math.Abs(float64(v))
		switch {
		case m < 0.00001:
			return "0.0"
		case m < 1:
			return fmt.Sprintf("%.5f", v)
		case m < 10:
			return fmt.Sprintf("%.4f", v)
		case m < 100:
			return fmt.Sprintf("%.3f", v)
		}
		return fmt.Sprintf("%.2f", v)
	}
	return ""
}

// Selected returns the selected rows in ascending order.
func (l *List) Selected() []int {
	var out []int
	n := l.count()
	for i := 0; i < n; i++ {
		if l.selected[i] {
			out = append(out, i)
		}
	}
	return out
}

// SelectedCount returns the number of selected rows.
func (l *List) SelectedCount() int { return len(l.Selected()) }

// Select selects only row i (-1 clears the selection) and scrolls to it.
func (l *List) Select(i int) {
	l.selected = map[int]bool{}
	if i >= 0 && i < l.count() {
		l.selected[i] = true
		l.focus, l.anchor, l.scrollTo = i, i, i
	}
}

// SelectAll selects every row (multi-selection lists only).
func (l *List) SelectAll() bool {
	if l.SingleSel {
		return false
	}
	for i := 0; i < l.count(); i++ {
		l.selected[i] = true
	}
	return true
}

// Reset clears the selection (after the rows were reloaded).
func (l *List) Reset() {
	l.selected = map[int]bool{}
	l.focus, l.anchor = -1, -1
}

func (l *List) count() int {
	if l.Rows == nil {
		return 0
	}
	return l.Rows()
}

func (l *List) changed() {
	if l.Changed != nil {
		l.Changed()
	}
}

// ClickColumn sorts on a column like a header click (onColumnClick): the
// same column flips the order, another one becomes the first key and the
// old one the second.
func (l *List) ClickColumn(tag sdata.Tag) {
	if tag == l.SortColumn {
		l.Ascending = !l.Ascending
	} else {
		l.PrevSortColumn = l.SortColumn
		l.SortColumn = tag
	}
	l.Resort()
}

// Resort sorts again with the current keys.
func (l *List) Resort() {
	if l.Sort != nil {
		sel := l.selectedNodes()
		l.Sort(l.SortColumn, l.PrevSortColumn, l.Ascending)
		l.reselect(sel)
	}
}

func (l *List) selectedNodes() []Node {
	var out []Node
	for _, i := range l.Selected() {
		out = append(out, l.Row(i))
	}
	return out
}

func (l *List) reselect(nodes []Node) {
	l.selected = map[int]bool{}
	if len(nodes) == 0 {
		return
	}
	want := map[Node]bool{}
	for _, n := range nodes {
		want[n] = true
	}
	for i := 0; i < l.count(); i++ {
		if want[l.Row(i)] {
			l.selected[i] = true
		}
	}
}

// InsertNew inserts a new row after the selection (insertNew).
func (l *List) InsertNew(index int) bool {
	if l.Insert == nil {
		return false
	}
	at := l.Insert(index)
	if at < 0 {
		return false
	}
	l.Select(at)
	l.changed()
	return true
}

func (l *List) confirm(text string, then func()) {
	if l.Confirm == nil {
		then()
		return
	}
	l.Confirm(text, func(yes bool) {
		if yes {
			then()
		}
	})
}

// DeleteSelected removes the selected rows after a confirmation
// (deleteSelected); verb is "to delete" or "to cut".
func (l *List) DeleteSelected(verb string) {
	sel := l.Selected()
	if len(sel) == 0 || l.Remove == nil {
		return
	}
	text := "Are you sure you want " + verb + " this item ?"
	if len(sel) > 1 {
		text = fmt.Sprintf("Are you sure you want %s these %d items ?", verb, len(sel))
	}
	l.confirm(text, func() {
		for k := len(sel) - 1; k >= 0; k-- {
			if !l.Remove(sel[k]) {
				break
			}
		}
		if len(sel) == 1 {
			l.Select(min(sel[0], l.count()-1))
		} else {
			l.Reset()
		}
		l.changed()
	})
}

// ClearSelected resets the selected rows (clearSelected, "Erase item").
func (l *List) ClearSelected() {
	sel := l.Selected()
	if len(sel) == 0 || l.Clear == nil {
		return
	}
	run := func() {
		any := false
		for k := len(sel) - 1; k >= 0; k-- {
			if !l.Clear(sel[k]) {
				break
			}
			any = true
		}
		if any {
			l.changed()
		}
	}
	if len(sel) == 1 {
		run()
		return
	}
	l.confirm(fmt.Sprintf("Are you sure you want to clear these %d items ?", len(sel)), run)
}

// CopySelected copies the selected rows (copy).
func (l *List) CopySelected() bool {
	sel := l.Selected()
	if len(sel) == 0 || l.Copy == nil {
		return false
	}
	if text := l.Copy(sel); text != "" {
		imgui.SetClipboardText(text)
	}
	return true
}

// CutSelected copies then deletes (cut).
func (l *List) CutSelected() {
	if l.CopySelected() {
		l.DeleteSelected("to cut")
	}
}

// PasteRows pastes the internal clipboard (paste).
func (l *List) PasteRows() {
	if l.Paste != nil && l.Paste(l.Selected()) {
		l.changed()
	}
}

func (l *List) canPaste() bool { return l.CanPaste != nil && l.CanPaste() }

// Keys handles the keyboard while the list has focus (ListCtrlOD::onKey).
func (l *List) Keys() {
	io := imgui.CurrentIO()
	ctrl, shift := io.KeyCtrl(), io.KeyShift()
	pressed := func(k imgui.Key) bool { return imgui.IsKeyPressedBool(k) }
	n := l.count()
	switch {
	case ctrl && pressed(imgui.KeyD):
		l.Reset()
	case ctrl && pressed(imgui.KeyV), shift && pressed(imgui.KeyInsert):
		l.PasteRows()
	case ctrl && pressed(imgui.KeyX):
		l.CutSelected()
	case ctrl && (pressed(imgui.KeyC) || pressed(imgui.KeyInsert)):
		l.CopySelected()
	case ctrl && pressed(imgui.KeyA):
		l.SelectAll()
	case ctrl && pressed(imgui.KeyUpArrow):
		if sel := l.Selected(); len(sel) == 1 && sel[0] > 0 && l.MoveUp != nil && l.MoveUp(sel[0]) {
			l.Select(sel[0] - 1)
		}
	case ctrl && pressed(imgui.KeyDownArrow):
		if sel := l.Selected(); len(sel) == 1 && sel[0] < n-1 && l.MoveDown != nil && l.MoveDown(sel[0]) {
			l.Select(sel[0] + 1)
		}
	case shift && pressed(imgui.KeyDelete):
		l.ClearSelected()
	case !ctrl && !shift && pressed(imgui.KeyInsert):
		at := -1
		if sel := l.Selected(); len(sel) > 0 {
			at = sel[0]
		}
		l.InsertNew(at)
	case !ctrl && pressed(imgui.KeyDelete):
		l.DeleteSelected("to delete")
	case pressed(imgui.KeyEnter) || pressed(imgui.KeyKeypadEnter):
		if l.focus >= 0 && l.focus < n && l.Activate != nil {
			l.Activate(l.focus)
		}
	case pressed(imgui.KeyUpArrow) && l.focus > 0:
		l.Select(l.focus - 1)
	case pressed(imgui.KeyDownArrow) && l.focus < n-1:
		l.Select(l.focus + 1)
	case pressed(imgui.KeyHome) && n > 0:
		l.Select(0)
	case pressed(imgui.KeyEnd) && n > 0:
		l.Select(n - 1)
	}
}

// EditMenu draws the Cut/Copy/Paste/Create/Delete/Erase items, enabled
// like ListCtrlOD::onInitMenuPopup.
func (l *List) EditMenu() {
	has := l.SelectedCount() > 0
	if imgui.MenuItemBoolV("Cut", "Ctrl+X", false, has && l.Remove != nil && l.Copy != nil) {
		l.CutSelected()
	}
	if imgui.MenuItemBoolV("Copy", "Ctrl+C", false, has && l.Copy != nil) {
		l.CopySelected()
	}
	if imgui.MenuItemBoolV("Paste", "Ctrl+V", false, l.canPaste()) {
		l.PasteRows()
	}
	if !l.SingleSel {
		if imgui.MenuItemBoolV("Select All", "Ctrl+A", false, true) {
			l.SelectAll()
		}
	}
	imgui.Separator()
	if imgui.MenuItemBoolV("Create item", "Ins", false, l.Insert != nil) {
		at := -1
		if sel := l.Selected(); len(sel) > 0 {
			at = sel[0]
		}
		l.InsertNew(at)
	}
	if imgui.MenuItemBoolV("Delete item", "Del", false, has && l.Remove != nil) {
		l.DeleteSelected("to delete")
	}
	if imgui.MenuItemBoolV("Erase item", "Shift+Del", false, has && l.Clear != nil) {
		l.ClearSelected()
	}
}

// Draw draws the list filling the available space.
func (l *List) Draw() {
	n := l.count()
	flags := imgui.TableFlagsBorders | imgui.TableFlagsRowBg | imgui.TableFlagsScrollY | imgui.TableFlagsScrollX |
		imgui.TableFlagsResizable | imgui.TableFlagsReorderable | imgui.TableFlagsHideable | imgui.TableFlagsSizingFixedFit
	if len(l.Columns) == 0 {
		return
	}
	if !imgui.BeginTableV(l.ID, int32(len(l.Columns)), flags, imgui.NewVec2(0, 0), 0) {
		return
	}
	imgui.TableSetupScrollFreeze(1, 1)
	for _, c := range l.Columns {
		w := c.Width
		if w <= 0 {
			w = 60
		}
		imgui.TableSetupColumnV(c.Title+"##"+string(c.Tag), imgui.TableColumnFlagsWidthFixed, w, 0)
	}
	// Header row with click-to-sort.
	imgui.TableNextRowV(imgui.TableRowFlagsHeaders, 0)
	for i, c := range l.Columns {
		if !imgui.TableSetColumnIndex(int32(i)) {
			continue
		}
		label := c.Title
		if c.Tag == l.SortColumn {
			if l.Ascending {
				label += " ^"
			} else {
				label += " v"
			}
		}
		imgui.TableHeader(label + "##h" + string(c.Tag))
		if imgui.IsItemClicked() && c.Tag != "" {
			l.ClickColumn(c.Tag)
		}
	}
	if imgui.IsWindowFocused() {
		l.Keys()
	}
	clipper := imgui.NewListClipper()
	clipper.Begin(int32(n))
	if l.scrollTo >= 0 && l.scrollTo < n {
		clipper.IncludeItemByIndex(int32(l.scrollTo))
	}
	io := imgui.CurrentIO()
	for clipper.Step() {
		for row := int(clipper.DisplayStart()); row < int(clipper.DisplayEnd()); row++ {
			node := l.Row(row)
			imgui.TableNextRow()
			imgui.PushIDInt(int32(row))
			for ci, c := range l.Columns {
				if !imgui.TableSetColumnIndex(int32(ci)) {
					continue
				}
				if ci == 0 {
					if l.Icon != nil {
						l.Icon(node, imgui.TextLineHeight())
						imgui.SameLine()
					}
					sel := l.selected[row]
					if imgui.SelectableBoolV(l.FormatValue(node, c)+"##row", sel,
						imgui.SelectableFlagsSpanAllColumns|imgui.SelectableFlagsAllowDoubleClick|imgui.SelectableFlagsAllowOverlap,
						imgui.NewVec2(0, 0)) {
						l.click(row, io.KeyCtrl(), io.KeyShift())
						if imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) && l.Activate != nil {
							l.Activate(row)
						}
					}
					if imgui.IsItemClickedV(imgui.MouseButtonRight) && !l.selected[row] {
						l.Select(row)
					}
					if l.scrollTo == row {
						imgui.SetScrollHereYV(0.5)
						l.scrollTo = -1
					}
					continue
				}
				text := l.FormatValue(node, c)
				if align := c.Format & 3; align != AlignLeft {
					w := imgui.CalcTextSize(text).X
					avail := imgui.ContentRegionAvail().X
					off := avail - w
					if align == AlignCenter {
						off /= 2
					}
					if off > 0 {
						imgui.SetCursorPosX(imgui.CursorPosX() + off)
					}
				}
				imgui.TextUnformatted(text)
			}
			imgui.PopID()
		}
	}
	clipper.End()
	clipper.Destroy()
	if imgui.BeginPopupContextWindowV("##ctx"+l.ID, imgui.PopupFlagsMouseButtonRight) {
		if l.ContextMenu != nil {
			l.ContextMenu()
		} else {
			l.EditMenu()
		}
		imgui.EndPopup()
	}
	imgui.EndTable()
}

func (l *List) click(row int, ctrl, shift bool) {
	switch {
	case shift && !l.SingleSel && l.anchor >= 0:
		if !ctrl {
			l.selected = map[int]bool{}
		}
		a, b := l.anchor, row
		if a > b {
			a, b = b, a
		}
		for i := a; i <= b; i++ {
			l.selected[i] = true
		}
	case ctrl && !l.SingleSel:
		l.selected[row] = !l.selected[row]
		l.anchor = row
	default:
		l.selected = map[int]bool{row: true}
		l.anchor = row
	}
	l.focus = row
}
