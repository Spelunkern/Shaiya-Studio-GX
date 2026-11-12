package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// ItemSet is a row of SetItem.SData / DBSetItemData (Data::ItemSet): a set
// of up to 13 items and the bonus text of each count.
type ItemSet struct {
	Index int16
	Name  string
	Types [13]int16
	IDs   [13]int16
	Bonus [13]string
}

// NewItemSet returns an empty set.
func NewItemSet(index int16) *ItemSet { return &ItemSet{Index: index} }

// Reset clears the set but keeps its index.
func (s *ItemSet) Reset() { *s = ItemSet{Index: s.Index} }

func setSlot(tag Tag) (kind byte, i int, ok bool) {
	if len(tag) != 4 {
		return 0, 0, false
	}
	var n int
	if _, err := fmt.Sscanf(string(tag[2:]), "%02d", &n); err != nil || n < 1 || n > 13 {
		return 0, 0, false
	}
	switch tag[:2] {
	case "tp":
		return 't', n - 1, true
	case "id":
		return 'i', n - 1, true
	case "bn":
		return 'b', n - 1, true
	}
	return 0, 0, false
}

// Get returns "indx", "tpNN" or "idNN".
func (s *ItemSet) Get(tag Tag) int {
	if tag == "indx" {
		return int(s.Index)
	}
	switch k, i, ok := setSlot(tag); {
	case ok && k == 't':
		return int(s.Types[i])
	case ok && k == 'i':
		return int(s.IDs[i])
	}
	return 0
}

// Set sets "tpNN" or "idNN".
func (s *ItemSet) Set(tag Tag, v int) bool {
	switch k, i, ok := setSlot(tag); {
	case ok && k == 't':
		s.Types[i] = int16(v)
	case ok && k == 'i':
		s.IDs[i] = int16(v)
	default:
		return false
	}
	return true
}

// Text returns "name" or "bnNN".
func (s *ItemSet) Text(tag Tag) string {
	if tag == "name" {
		return s.Name
	}
	if k, i, ok := setSlot(tag); ok && k == 'b' {
		return s.Bonus[i]
	}
	return ""
}

// SetText sets "name" or "bnNN".
func (s *ItemSet) SetText(tag Tag, v string) bool {
	if tag == "name" {
		s.Name = v
		return true
	}
	if k, i, ok := setSlot(tag); ok && k == 'b' {
		s.Bonus[i] = v
		return true
	}
	return false
}

// Compare compares on "indx" or "name".
func (s *ItemSet) Compare(o *ItemSet, tag Tag) int {
	switch tag {
	case "indx":
		return cmp(s.Index, o.Index)
	case "name":
		return compareLogical(s.Name, o.Name)
	}
	return 0
}

// Assign copies everything but the index.
func (s *ItemSet) Assign(src *ItemSet) bool {
	idx := s.Index
	*s = *src
	s.Index = idx
	return true
}

// Line exports the set as "index;name;type;id;bonus;..." (exportText).
func (s *ItemSet) Line() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d;%s", s.Index, s.Name)
	for i := range s.Types {
		fmt.Fprintf(&b, ";%d;%d;%s", s.Types[i], s.IDs[i], s.Bonus[i])
	}
	return b.String()
}

// Read reads a row (Data::ItemSet::read).
func (s *ItemSet) Read(r *binstream.Reader, f Format) {
	switch f {
	case File:
		s.Index = r.Int16()
		s.Name = r.String(4, r.CodePage)
		for i := range s.Types {
			s.Types[i] = r.Int16()
			s.IDs[i] = r.Int16()
		}
		for i := range s.Bonus {
			s.Bonus[i] = lfToCrLf(r.String(4, r.CodePage))
		}
	case Flnt:
		s.Index = int16(r.Int64())
		for i := range s.Types {
			s.Types[i] = int16(r.Int64())
			s.IDs[i] = int16(r.Int64())
		}
	case Ltxt:
		s.Name = r.String(4, -1)
		for i := range s.Bonus {
			s.Bonus[i] = lfToCrLf(r.String(4, -1))
		}
	}
}

// Write writes a row (Data::ItemSet::write).
func (s *ItemSet) Write(w *binstream.Writer, f Format) {
	switch f {
	case File:
		w.Int16(s.Index)
		w.String(s.Name, 4, w.CodePage)
		for i := range s.Types {
			w.Int16(s.Types[i])
			w.Int16(s.IDs[i])
		}
		for i := range s.Bonus {
			w.String(crLfToLf(s.Bonus[i]), 4, w.CodePage)
		}
	case Flnt:
		w.Int64(int64(s.Index))
		for i := range s.Types {
			w.Int64(int64(s.Types[i]))
			w.Int64(int64(s.IDs[i]))
		}
	case Ltxt:
		w.Int64(int64(s.Index))
		w.String(s.Name, 4, -1)
		for i := range s.Bonus {
			w.String(crLfToLf(s.Bonus[i]), 4, -1)
		}
	}
}

// ItemSets is SetItem.SData / DBSetItemData / DBSetItemText (Data::ItemSets).
type ItemSets struct {
	Sets     []*ItemSet
	DataHead []byte
	TextHead []byte
}

// FilePrefix is the prefix of the DB text files.
func (s *ItemSets) FilePrefix() string { return `BinarySData\DBSetItemText_` }

// Read loads the table (Data::ItemSets::read).
func (s *ItemSets) Read(data []byte, cp int, f Format) error {
	return binstream.Try(func() {
		r := binstream.NewReader(data, cp)
		if f == Ltxt {
			s.TextHead = r.Copy(128)
			if r.Uint32() != 15 {
				failDB()
			}
			checkSignature(r, "id")
			checkSignature(r, "name")
			for i := 1; i <= 13; i++ {
				checkSignature(r, fmt.Sprintf("seteff%d", i))
			}
			n := int(r.Uint32())
			for ; n > 0; n-- {
				id := int(r.Int64())
				var found *ItemSet
				for _, x := range s.Sets {
					if int(x.Index) == id {
						found = x
						break
					}
				}
				if found == nil {
					found = NewItemSet(0)
				}
				found.Read(r, Ltxt)
			}
			return
		}
		s.Sets = nil
		if f == Flnt {
			s.DataHead = r.Copy(128)
			if r.Uint32() != 27 {
				failDB()
			}
			checkSignature(r, "id")
			for j := 1; j <= 13; j++ {
				checkSignature(r, fmt.Sprintf("category%d", j))
				checkSignature(r, fmt.Sprintf("number%d", j))
			}
		}
		n := r.Int()
		for k := 1; k <= n; k++ {
			set := NewItemSet(int16(k))
			s.Sets = append(s.Sets, set)
			set.Read(r, f)
		}
	})
}

// Write serializes the table (Data::ItemSets::write).
func (s *ItemSets) Write(cp int, f Format) ([]byte, error) {
	w := binstream.NewWriter(cp)
	switch f {
	case Ltxt:
		w.Put(head(s.TextHead))
		w.Uint32(15)
		writeColumn(w, "id")
		writeColumn(w, "name")
		for i := 1; i <= 13; i++ {
			writeColumn(w, fmt.Sprintf("seteff%d", i))
		}
	case Flnt:
		w.Put(head(s.DataHead))
		w.Uint32(27)
		writeColumn(w, "id")
		for j := 1; j <= 13; j++ {
			writeColumn(w, fmt.Sprintf("category%d", j))
			writeColumn(w, fmt.Sprintf("number%d", j))
		}
	}
	w.Uint32(uint32(len(s.Sets)))
	for _, set := range s.Sets {
		set.Write(w, f)
	}
	return w.Bytes(), nil
}

// Text exports the table (exportText '----').
func (s *ItemSets) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d\r\n", len(s.Sets))
	for _, set := range s.Sets {
		b.WriteString(set.Line())
		b.WriteString("\r\n")
	}
	return b.String()
}

// head returns a 128-byte DB header (zeroes when the table had none; the
// C++ code would write uninitialized memory).
func head(h []byte) []byte {
	if len(h) == 128 {
		return h
	}
	return make([]byte, 128)
}
