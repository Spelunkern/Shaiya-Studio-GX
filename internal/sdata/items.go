package sdata

import (
	"errors"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// Items is Item.SData / DBItemData.SData / DBItemText_<LANG>.SData
// (Data::Items).
type Items struct {
	Items    []*Item
	MaxType  int
	Version  int    // m_value30
	DataHead []byte // 128-byte DB data header (m_buffer1)
	TextHead []byte // 128-byte DB text header (m_buffer2)
	columns  []byte // column block of the DB text file
}

// NewItems returns an empty table (version 0x500).
func NewItems() *Items { return &Items{Version: 0x500} }

// FilePrefix is the prefix of the DB text files.
func (s *Items) FilePrefix() string { return `BinarySData\DBItemText_` }

// Find returns the item type/typeID (sub_42FC00).
func (s *Items) Find(t, id int) *Item {
	for _, it := range s.Items {
		if int(it.Type) == t && int(it.TypeID) == id {
			return it
		}
	}
	return nil
}

// ByGrade returns the items of a grade (sub_42FC50).
func (s *Items) ByGrade(grade int) []*Item {
	var out []*Item
	for _, it := range s.Items {
		if int(it.Grade) == grade {
			out = append(out, it)
		}
	}
	return out
}

// SetVersion sets the file version of the table and its rows
// (setCodePage).
func (s *Items) SetVersion(v int) {
	s.Version = v
	for _, it := range s.Items {
		it.Version = v
	}
}

// readFile reads the .SData format with version v (sub_4301C0).
func (s *Items) readFile(data []byte, cp, v int) bool {
	s.Items = nil
	s.Version = v
	err := binstream.Try(func() {
		r := binstream.NewReader(data, cp)
		index := 1
		s.MaxType = r.Int()
		for t := 1; t <= s.MaxType; t++ {
			n := r.Int()
			for id := 1; id <= n; id++ {
				it := NewItem(index, v)
				index++
				it.Read(r, File)
				if it.Type == 0 {
					it.Type = uint8(t)
					it.TypeID = uint8(id)
					it.Name = "???" + itoa(id)
				}
				s.Items = append(s.Items, it)
			}
		}
		SortList(s.Items, "type", "tpid", true)
	})
	if err != nil {
		s.Items = nil
		return false
	}
	return true
}

func itoa(v int) string { return i32s(int32(v)) }

// ErrFormat is returned when a file is not in a known format.
var ErrFormat = errors.New("sdata: invalid or not supported format")

// itemDBColumns are the 69 columns of DBItemData.
var itemDBColumns = []string{
	"itemtype", "itemtypeid", "image", "icon", "level", "country", "attackfighter", "defensefighter",
	"patrolrogue", "shootrogue", "attackmage", "defensemage", "grow", "str", "dex", "rec", "int",
	"wis", "luc", "vg", "og", "ig", "range", "attacktime", "attrib", "special", "slot", "quality",
	"effect1", "effect2", "effect3", "effect4", "consthp", "constsp", "constmp", "conststr",
	"constdex", "constrec", "constint", "constwis", "constluc", "speed", "exp", "buy", "sell",
	"grade", "drop", "server", "count", "duration", "extduration", "secoption", "optionrate",
	"buymethod", "maxlevel", "weaponpart", "dyeingtype", "arg3", "arg4", "usecontype", "useconvar",
	"moneytype", "itemskill", "itemupgrade", "arg10", "genecount", "arg12", "spellbookexp",
	"spellbookdurability",
}

// Read loads the table from data in format f (Data::Items::read). For
// File the three versions are tried in turn (EP5, EP6, EP6.2). Ltxt merges
// names and descriptions into the loaded rows.
func (s *Items) Read(data []byte, cp int, f Format) error {
	switch f {
	case Ltxt:
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			s.TextHead = r.Copy(128)
			if r.Uint32() != 4 {
				failDB()
			}
			start := r.Pos
			names := []string{"itemtype", "itemtypeid", "itemname", "text"}
			seen := make([]bool, 4)
			for c := 0; c < 4; c++ {
				name := r.String(1, 1200)
				k := 0
				for k < 4 && (seen[k] || !strings.EqualFold(name, names[k])) {
					k++
				}
				if k == 4 {
					failDB()
				}
				seen[k] = true
			}
			s.columns = append([]byte(nil), r.Data[start:r.Pos]...)
			n := r.Int()
			for ; n > 0; n-- {
				t := int(r.Int64())
				id := int(r.Int64())
				if it := s.Find(t, id); it != nil {
					it.Read(r, Ltxt)
				} else {
					r.String(4, -1)
					r.String(4, -1)
				}
			}
		})
	case Flnt:
		s.Items = nil
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			s.DataHead = r.Copy(128)
			columns := int(r.Uint32())
			if columns < 69 {
				failDB()
			}
			s.Version = 0x1600
			for _, c := range itemDBColumns {
				checkSignature(r, c)
			}
			for k := columns - 69; k > 0; k-- {
				r.String(1, 1200)
			}
			n := int(r.Uint32())
			for i := 1; i <= n; i++ {
				it := NewItem(i, 0x1620)
				s.Items = append(s.Items, it)
				it.Read(r, Flnt)
				for k := columns - 69; k > 0; k-- {
					r.Int64()
				}
			}
		})
	case File:
		for _, v := range []int{0x500, 0x600, 0x620} {
			if s.readFile(data, cp, v) {
				return nil
			}
		}
		return ErrFormat
	}
	return ErrFormat
}

// Write serializes the table in format f (Data::Items::write). The rows
// are sorted by type and type id first.
func (s *Items) Write(cp int, f Format) ([]byte, error) {
	SortList(s.Items, "type", "tpid", true)
	w := binstream.NewWriter(cp)
	switch f {
	case Ltxt:
		if s.TextHead == nil {
			return nil, errors.New("Error: textHeader does not exist")
		}
		w.Put(s.TextHead)
		w.Uint32(4)
		if len(s.columns) > 0 {
			w.Put(s.columns)
		} else {
			for _, c := range []string{"itemtype", "itemtypeid", "itemname", "text"} {
				writeColumn(w, c)
			}
		}
		w.Uint32(uint32(len(s.Items)))
		for _, it := range s.Items {
			it.Write(w, Ltxt)
		}
	case Flnt:
		if s.DataHead == nil {
			return nil, errors.New("Error: dataHeader does not exist")
		}
		w.Put(s.DataHead)
		w.Uint32(69)
		for _, c := range itemDBColumns {
			writeColumn(w, c)
		}
		w.Uint32(uint32(len(s.Items)))
		for _, it := range s.Items {
			it.Write(w, Flnt)
		}
	case File:
		maxType := 0
		for _, it := range s.Items {
			if int(it.Type) > maxType {
				maxType = int(it.Type)
			}
		}
		s.MaxType = maxType
		w.Uint32(uint32(maxType))
		index := 0
		for t := 1; t <= maxType; t++ {
			n := 0
			for _, it := range s.Items {
				if int(it.Type) == t {
					n++
				}
			}
			w.Uint32(uint32(n))
			for ; n > 0; n-- {
				s.Items[index].Write(w, File)
				index++
			}
		}
	}
	return w.Bytes(), nil
}

// Text exports the table as CSV or SQL (toText).
func (s *Items) Text(format string) string {
	var b strings.Builder
	if format == TextSQL {
		b.WriteString("TRUNCATE TABLE [PS_GameDefs].[dbo].[Items];\r\n")
		for i, it := range s.Items {
			if i%100 == 0 {
				b.WriteString("GO\r\n")
			}
			line, _ := it.Line(TextSQL)
			b.WriteString(line)
		}
		b.WriteString("GO\r\n")
		return b.String()
	}
	for _, it := range s.Items {
		if line, ok := it.Line(format); ok {
			b.WriteString(line)
		}
	}
	return b.String()
}
