package sdata

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/codepage"
)

// DBText split modes: how the columns of a DB text table were divided
// between integer keys and strings.
const (
	SplitNone        = 0
	SplitByName      = 1
	SplitImplicitKey = 2
	SplitLeadingKeys = 3
)

type dbTextRow struct {
	keys []int64
	strs [][]byte
}

// DBTextTable is a generic DB*Text_<LANG>.SData table (DBTextTable): any
// number of 64-bit key columns followed by length-prefixed strings.
type DBTextTable struct {
	CodePage int
	prefix   []byte
	columns  []string
	keyCols  []int
	strCols  []int
	rows     []dbTextRow
	split    int
	loaded   bool
	err      string
}

// NewDBTextTable returns an empty table (code page 1252).
func NewDBTextTable() *DBTextTable { return &DBTextTable{CodePage: 1252} }

// DBTextCodePage guesses the code page of a DB text file from the language
// in its name (codePageForName).
func DBTextCodePage(fileName string) int {
	base := fileName[strings.LastIndexAny(fileName, `\/`)+1:]
	lang := ""
	lower := strings.ToLower(base)
	for i := 0; i+5 <= len(lower); i++ {
		if lower[i:i+5] == "text_" {
			lang = base[i+5:]
		}
	}
	if lang == "" {
		return 1252
	}
	if i := strings.IndexAny(lang, "_."); i >= 0 {
		lang = lang[:i]
	}
	switch strings.ToUpper(lang) {
	case "CHN", "CN":
		return 936
	case "KOR", "KR":
		return 949
	case "JPN", "JP":
		return 932
	case "TWN", "TW":
		return 950
	case "RUS", "RU":
		return 1251
	case "POL", "PL":
		return 1250
	case "TUR", "TR":
		return 1254
	}
	return 1252
}

// SetCodePage sets the code page (1252 when cp <= 0).
func (t *DBTextTable) SetCodePage(cp int) {
	if cp <= 0 {
		cp = 1252
	}
	t.CodePage = cp
}

// IsEmpty reports whether nothing is loaded.
func (t *DBTextTable) IsEmpty() bool { return !t.loaded }

// Clear empties the table.
func (t *DBTextTable) Clear() {
	cp := t.CodePage
	*t = DBTextTable{CodePage: cp}
}

func numericName(name string) bool {
	n := strings.ToLower(name)
	return strings.HasSuffix(n, "id") || strings.HasSuffix(n, "type") || strings.HasSuffix(n, "level") ||
		strings.HasSuffix(n, "code")
}

func (t *DBTextTable) tryLayout(data []byte, pos, nKeys, nStrs int) ([]dbTextRow, bool) {
	size := len(data)
	if nKeys+nStrs <= 0 || pos+4 > size {
		return nil, false
	}
	count := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	minRow := uint64(8*nKeys + 4*nStrs)
	if uint64(count)*minRow > uint64(size-pos) {
		return nil, false
	}
	rows := make([]dbTextRow, 0, count)
	for r := uint32(0); r < count; r++ {
		row := dbTextRow{keys: make([]int64, nKeys), strs: make([][]byte, nStrs)}
		for k := range row.keys {
			if pos+8 > size {
				return nil, false
			}
			row.keys[k] = int64(binary.LittleEndian.Uint64(data[pos:]))
			pos += 8
		}
		for s := range row.strs {
			if pos+4 > size {
				return nil, false
			}
			n := binary.LittleEndian.Uint32(data[pos:])
			pos += 4
			if n > 0x100000 || int(n) > size-pos {
				return nil, false
			}
			row.strs[s] = append([]byte(nil), data[pos:pos+int(n)]...)
			pos += int(n)
		}
		rows = append(rows, row)
	}
	return rows, pos == size
}

// Parse loads the table, trying the column splits in shStudio's order: by
// column name, one implicit key, then 0..n leading key columns.
func (t *DBTextTable) Parse(data []byte) bool {
	t.Clear()
	if len(data) < 0x80+8 {
		t.err = "The file is too small to be a DB text table."
		return false
	}
	pos := 0x80
	ncols := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	if ncols == 0 || ncols > 256 {
		t.err = "Invalid column count in the DB text table header."
		return false
	}
	var cols []string
	for c := uint32(0); c < ncols; c++ {
		if pos+1 > len(data) {
			t.err = "Truncated column list."
			return false
		}
		chars := int(data[pos])
		pos++
		if pos+2*chars > len(data) {
			t.err = "Truncated column list."
			return false
		}
		cols = append(cols, codepage.Decode(data[pos:pos+2*chars], codepage.UTF16LE))
		pos += 2 * chars
	}
	type cand struct {
		split      int
		keys, strs []int
	}
	var cands []cand
	byName := cand{split: SplitByName}
	for i, c := range cols {
		if numericName(c) {
			byName.keys = append(byName.keys, i)
		} else {
			byName.strs = append(byName.strs, i)
		}
	}
	cands = append(cands, byName)
	implicit := cand{split: SplitImplicitKey, keys: []int{-1}}
	for i := range cols {
		implicit.strs = append(implicit.strs, i)
	}
	cands = append(cands, implicit)
	for k := 0; k <= len(cols); k++ {
		c := cand{split: SplitLeadingKeys}
		for i := range cols {
			if i < k {
				c.keys = append(c.keys, i)
			} else {
				c.strs = append(c.strs, i)
			}
		}
		cands = append(cands, c)
	}
	for _, c := range cands {
		rows, ok := t.tryLayout(data, pos, len(c.keys), len(c.strs))
		if !ok {
			continue
		}
		t.prefix = append([]byte(nil), data[:pos]...)
		t.columns = cols
		t.keyCols, t.strCols = c.keys, c.strs
		t.rows = rows
		t.split = c.split
		t.loaded = true
		return true
	}
	t.err = "Unknown DB text table layout: no split of the columns into integer keys and strings matches the data."
	return false
}

// Read loads the table (read).
func (t *DBTextTable) Read(data []byte, cp int, f Format) error {
	if !t.Parse(data) {
		return fmt.Errorf("%s", t.err)
	}
	return nil
}

// Write serializes the table (write); nothing when not loaded.
func (t *DBTextTable) Write(cp int, f Format) ([]byte, error) {
	if !t.loaded {
		return nil, nil
	}
	out := append([]byte(nil), t.prefix...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(t.rows)))
	for _, r := range t.rows {
		for _, k := range r.keys {
			out = binary.LittleEndian.AppendUint64(out, uint64(k))
		}
		for _, s := range r.strs {
			out = binary.LittleEndian.AppendUint32(out, uint32(len(s)))
			out = append(out, s...)
		}
	}
	return out, nil
}

func (t *DBTextTable) RowCount() int                 { return len(t.rows) }
func (t *DBTextTable) FieldCount() int               { return len(t.keyCols) + len(t.strCols) }
func (t *DBTextTable) KeyCount() int                 { return len(t.keyCols) }
func (t *DBTextTable) IsNumericField(field int) bool { return IsKeyField(t, field) }
func (t *DBTextTable) TypeName() string              { return "a DB text table" }
func (t *DBTextTable) Error() string                 { return t.err }

// FieldName returns a column name ("(key)" for the implicit key).
func (t *DBTextTable) FieldName(field int) string {
	if field < 0 || field >= t.FieldCount() {
		return ""
	}
	col := 0
	if field < t.KeyCount() {
		col = t.keyCols[field]
	} else {
		col = t.strCols[field-t.KeyCount()]
	}
	if col < 0 {
		return "(key)"
	}
	return t.columns[col]
}

// SplitName names the split mode.
func (t *DBTextTable) SplitName() string {
	switch t.split {
	case SplitByName:
		return "by name"
	case SplitImplicitKey:
		return "implicit key"
	case SplitLeadingKeys:
		return "leading keys"
	}
	return "none"
}

// StatusText describes the table.
func (t *DBTextTable) StatusText() string {
	return fmt.Sprintf("%d rows, %d columns (%d integer keys, split %s), code page %d",
		t.RowCount(), t.FieldCount(), t.KeyCount(), t.SplitName(), t.CodePage)
}

// NotEncodableText is the message for text the code page can not store.
func (t *DBTextTable) NotEncodableText() string {
	return fmt.Sprintf("The text contains characters that code page %d cannot store. The cell was not changed.", t.CodePage)
}

func (t *DBTextTable) decode(b []byte) string { return codepage.Decode(b, t.CodePage) }

func (t *DBTextTable) encode(s string) ([]byte, bool) {
	if s == "" {
		return nil, true
	}
	b := codepage.Encode(s, t.CodePage)
	cp := codepage.Resolve(t.CodePage)
	if cp == codepage.UTF8 {
		return b, true
	}
	return b, t.decode(b) == s && utf8.ValidString(s)
}

// GetText returns a cell as text.
func (t *DBTextTable) GetText(row, field int) string {
	if row < 0 || row >= t.RowCount() || field < 0 || field >= t.FieldCount() {
		return ""
	}
	r := t.rows[row]
	if field < t.KeyCount() {
		return strconv.FormatInt(r.keys[field], 10)
	}
	return t.decode(r.strs[field-t.KeyCount()])
}

// Key returns a key cell.
func (t *DBTextTable) Key(row, field int) int64 {
	if row < 0 || row >= t.RowCount() || !IsKeyField(t, field) {
		return 0
	}
	return t.rows[row].keys[field]
}

// SetText sets a cell (setText).
func (t *DBTextTable) SetText(row, field int, text string) int {
	if row < 0 || row >= t.RowCount() || field < 0 || field >= t.FieldCount() {
		return SetBadCell
	}
	r := &t.rows[row]
	if field < t.KeyCount() {
		v, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err != nil || strings.TrimLeft(text, " ") == "" {
			return SetBadNumber
		}
		if r.keys[field] == v {
			return SetUnchanged
		}
		r.keys[field] = v
		return SetChanged
	}
	b, ok := t.encode(text)
	if !ok {
		return SetNotEncodable
	}
	cell := &r.strs[field-t.KeyCount()]
	if t.decode(*cell) == text {
		return SetUnchanged
	}
	*cell = b
	return SetChanged
}

// InsertRow inserts an empty row at index (or at the end); the keys copy
// the previous row with the last key incremented.
func (t *DBTextTable) InsertRow(index int) int {
	if !t.loaded {
		return -1
	}
	if index < 0 || index > t.RowCount() {
		index = t.RowCount()
	}
	row := dbTextRow{keys: make([]int64, len(t.keyCols)), strs: make([][]byte, len(t.strCols))}
	if index > 0 && len(row.keys) > 0 {
		copy(row.keys, t.rows[index-1].keys)
		row.keys[len(row.keys)-1]++
	}
	t.rows = append(t.rows[:index], append([]dbTextRow{row}, t.rows[index:]...)...)
	return index
}

// DeleteRow removes a row.
func (t *DBTextTable) DeleteRow(index int) bool {
	if index < 0 || index >= t.RowCount() {
		return false
	}
	t.rows = append(t.rows[:index], t.rows[index+1:]...)
	return true
}
