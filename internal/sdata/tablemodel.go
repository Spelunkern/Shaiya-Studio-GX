package sdata

// TableModel is a generic editable table (TableModel): DB text tables and
// the simple SData tables shown in a grid.
type TableModel interface {
	RowCount() int
	FieldCount() int
	KeyCount() int
	IsNumericField(field int) bool
	FieldName(field int) string
	GetText(row, field int) string
	// SetText returns one of the Set* results.
	SetText(row, field int, text string) int
	InsertRow(index int) int
	DeleteRow(index int) bool
	StatusText() string
	TypeName() string
	Error() string
	NotEncodableText() string
}

// Results of TableModel.SetText.
const (
	SetUnchanged    = 0
	SetChanged      = 1
	SetBadNumber    = -1
	SetNotEncodable = -2
	SetBadCell      = -3
)

// IsKeyField reports whether field is one of the leading key fields.
func IsKeyField(m TableModel, field int) bool { return field >= 0 && field < m.KeyCount() }
