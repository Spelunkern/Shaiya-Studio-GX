package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// Item is a row of Item.SData / DBItemData.SData (Data::Item).
type Item struct {
	Index   int
	Version int // 0x500 (EP5), 0x600 (EP6), 0x620 (EP6.2); 0x1620 for DB files

	Name, Desc                     string
	Type, TypeID, Model, Icon      uint8
	Level                          int16
	Country                        uint8
	Class                          [6]uint8
	Grow, Type2, Type3             uint8
	ReqStr, ReqDex, ReqRec         uint16
	ReqInt, ReqWis, ReqLuc         uint16
	ReqVg                          int16
	Unk72                          uint16
	ReqOg, ReqIg                   uint8
	Range                          int16
	AttackTime, Attrib, Special    uint8
	Slot                           uint8
	Quality, Attack, AttackAdd     uint16
	Defense, Resist                uint16
	HP, SP, MP                     uint16
	Str, Dex, Rec, Int, Wis, Luc   uint16
	Speed, Exp                     uint8
	Buy, Sell                      int32
	Grade, Drop                    int16
	Server, Count                  uint8
	Duration                       int32
	ExtDuration, SecOption         uint8
	OptionRate, BuyMethod          uint8
	MaxLevel, WeaponPart           uint8
	DyeingType                     uint8
	Arg3, Arg4                     int32
	UseConType, UseConVar          int32
	MoneyType, ItemSkill           int32
	ItemUpgrade, Arg10             int32
	GeneCount, Arg12, SpellBookExp uint8
	SpellBookDurability            int32
	Valid                          bool

	flnt *flntRow
}

// Column names and tags of the item tables (g_itemColumnNames/Tags).
var (
	ItemColumnNames = []string{
		"Type", "TypeID", "Name", "Description", "Model", "Icon", "Level", "Country",
		"Class1", "Class2", "Class3", "Class4", "Class5", "Class6", "Grow", "Type2", "Type3",
		"ReqStr", "ReqDex", "ReqRec", "ReqInt", "ReqWis", "ReqLuc", "ReqVg", "ReqOg", "ReqIg",
		"Range", "AttackTime", "Attrib", "Special", "Slot", "Quality", "Attack", "Atk Add", "Def",
		"Resist", "Str", "Dex", "Rec", "Int", "Wis", "Luc", "HP", "SP", "MP", "Speed", "Exp",
		"Buy", "Sell", "Grade", "Drop", "Server", "Count",
	}
	ItemColumnTags = []Tag{
		"type", "tpid", "name", "desc", "mdel", "icon", "lvl ", "cntr", "cla1", "cla2", "cla3", "cla4",
		"cla5", "cla6", "grow", "typ2", "typ3", "rStr", "rDex", "rRec", "rInt", "rWis", "rLuc", "rVg ",
		"rOg ", "rIg ", "rang", "atkt", "attr", "spec", "slot", "qual", "attk", "atpl", "defp", "resm",
		"cStr", "cDex", "cRec", "cInt", "cWis", "cLuc", "cHP ", "cSP ", "cMP ", "sped", "exp ", "buy ",
		"sell", "grde", "drop", "srvr", "cnt ",
	}
	ItemColumnNamesEp6 = []string{
		"Duration", "Ext. Duration", "Sec. Option", "Option Rate", "Buy Method", "Max Level", "Weapon Part",
		"Dyeing Type", "Arg. 3", "Arg. 4", "UseConType", "UseConVar", "Money Type", "Item Skill", "Item Upgrade",
		"Arg. 10", "Gene Count", "Arg. 12", "Spell Exp", "Spell Duration",
	}
	ItemColumnTagsEp6 = []Tag{
		"dura", "edur", "sopt", "orat", "bmth", "mxlv", "wprt", "dyty", "arg3", "arg4",
		"ucty", "ucvr", "mnty", "itsk", "itup", "argA", "gnct", "argC", "spex", "spdu",
	}
)

// NewItem returns a reset item (Data::Item constructor).
func NewItem(index, version int) *Item {
	it := &Item{Index: index, Version: version}
	it.Reset()
	return it
}

// Reset clears the item; its name becomes "???" + the 3-digit type id.
func (it *Item) Reset() {
	t, id, idx, ver := it.Type, it.TypeID, it.Index, it.Version
	*it = Item{Index: idx, Version: ver, Type: t, TypeID: id}
	it.Name = fmt.Sprintf("???%03d", id)
	it.Type2, it.Type3 = t, t
}

// ptr returns a pointer to the field of tag, or nil.
func (it *Item) ptr(tag Tag) any {
	switch tag {
	case "type":
		return &it.Type
	case "tpid":
		return &it.TypeID
	case "mdel":
		return &it.Model
	case "icon":
		return &it.Icon
	case "lvl ":
		return &it.Level
	case "cntr":
		return &it.Country
	case "cla1":
		return &it.Class[0]
	case "cla2":
		return &it.Class[1]
	case "cla3":
		return &it.Class[2]
	case "cla4":
		return &it.Class[3]
	case "cla5":
		return &it.Class[4]
	case "cla6":
		return &it.Class[5]
	case "grow":
		return &it.Grow
	case "typ2":
		return &it.Type2
	case "typ3":
		return &it.Type3
	case "rStr":
		return &it.ReqStr
	case "rDex":
		return &it.ReqDex
	case "rRec":
		return &it.ReqRec
	case "rInt":
		return &it.ReqInt
	case "rWis":
		return &it.ReqWis
	case "rLuc":
		return &it.ReqLuc
	case "rVg ":
		return &it.ReqVg
	case "rOg ":
		return &it.ReqOg
	case "rIg ":
		return &it.ReqIg
	case "rang":
		return &it.Range
	case "atkt":
		return &it.AttackTime
	case "attr":
		return &it.Attrib
	case "spec":
		return &it.Special
	case "slot":
		return &it.Slot
	case "qual":
		return &it.Quality
	case "attk":
		return &it.Attack
	case "atpl":
		return &it.AttackAdd
	case "defp":
		return &it.Defense
	case "resm":
		return &it.Resist
	case "cHP ":
		return &it.HP
	case "cSP ":
		return &it.SP
	case "cMP ":
		return &it.MP
	case "cStr":
		return &it.Str
	case "cDex":
		return &it.Dex
	case "cRec":
		return &it.Rec
	case "cInt":
		return &it.Int
	case "cWis":
		return &it.Wis
	case "cLuc":
		return &it.Luc
	case "sped":
		return &it.Speed
	case "exp ":
		return &it.Exp
	case "buy ":
		return &it.Buy
	case "sell":
		return &it.Sell
	case "grde":
		return &it.Grade
	case "drop":
		return &it.Drop
	case "srvr":
		return &it.Server
	case "cnt ":
		return &it.Count
	case "dura":
		return &it.Duration
	case "edur":
		return &it.ExtDuration
	case "sopt":
		return &it.SecOption
	case "orat":
		return &it.OptionRate
	case "bmth":
		return &it.BuyMethod
	case "mxlv":
		return &it.MaxLevel
	case "wprt":
		return &it.WeaponPart
	case "dyty":
		return &it.DyeingType
	case "arg3":
		return &it.Arg3
	case "arg4":
		return &it.Arg4
	case "ucty":
		return &it.UseConType
	case "ucvr":
		return &it.UseConVar
	case "mnty":
		return &it.MoneyType
	case "itsk":
		return &it.ItemSkill
	case "itup":
		return &it.ItemUpgrade
	case "argA":
		return &it.Arg10
	case "gnct":
		return &it.GeneCount
	case "argC":
		return &it.Arg12
	case "spex":
		return &it.SpellBookExp
	case "spdu":
		return &it.SpellBookDurability
	}
	return nil
}

// Get returns the value of a numeric field (getInt).
func (it *Item) Get(tag Tag) int { return getInt(it.ptr(tag)) }

// Set sets a numeric field, truncated to its size (setInt).
func (it *Item) Set(tag Tag, v int) bool { return setInt(it.ptr(tag), v) }

// Text returns a text field (getString).
func (it *Item) Text(tag Tag) string {
	switch tag {
	case "name":
		return it.Name
	case "desc":
		return it.Desc
	case "cnts":
		if it.Type == 30 {
			return "All"
		}
		switch it.Country {
		case 0:
			return "Hum"
		case 1:
			return "Elf"
		case 2:
			return "Light"
		case 3:
			return "Dth"
		case 4:
			return "Vail"
		case 5:
			return "Dark"
		case 6:
			return "All"
		}
		return "--"
	case "typS":
		return fmt.Sprintf("%d-%d", it.Type, it.TypeID)
	case "dtls":
		var parts []string
		for _, s := range []struct {
			n string
			v uint16
		}{{"Str", it.Str}, {"Dex", it.Dex}, {"Rec", it.Rec}, {"Int", it.Int}, {"Wis", it.Wis},
			{"Luc", it.Luc}, {"HP", it.HP}, {"SP", it.SP}, {"MP", it.MP}} {
			if s.v != 0 {
				parts = append(parts, fmt.Sprintf("%s:%d", s.n, s.v))
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// SetStr sets the name or the description (setString).
func (it *Item) SetStr(tag Tag, v string) bool {
	switch tag {
	case "name":
		it.Name = v
	case "desc":
		it.Desc = v
	default:
		return false
	}
	return true
}

// IconID returns the icon index in the item icon sheets (getId).
func (it *Item) IconID() int { return ItemIconBase(it.Type) + int(it.Icon) }

// Compare compares two items on tag (compare). The name and description
// use the logical order; "cntr" and unknown tags compare equal.
func (it *Item) Compare(o *Item, tag Tag) int {
	switch tag {
	case "name":
		return compareLogical(it.Name, o.Name)
	case "desc":
		return compareLogical(it.Desc, o.Desc)
	case "cntr":
		return 0
	}
	a, b := it.ptr(tag), o.ptr(tag)
	if a == nil {
		return 0
	}
	return cmp(getInt(a), getInt(b))
}

// Compare3 compares on up to three keys (compare3).
func (it *Item) Compare3(o *Item, k1, k2, k3 Tag) int {
	r := it.Compare(o, k1)
	if r == 0 {
		if k2 != "" {
			r = it.Compare(o, k2)
		}
		if r == 0 && k3 != "" {
			return it.Compare(o, k3)
		}
	}
	return r
}

// Assign copies src into it when both have the same type (and type id,
// for a valid item); it reports whether it did (assign).
func (it *Item) Assign(src *Item) bool {
	if src.Type != it.Type || it.Valid && src.TypeID != it.TypeID {
		return false
	}
	idx, ver, id, unk, flnt := it.Index, it.Version, it.TypeID, it.Unk72, it.flnt
	*it = *src
	it.Index, it.Version, it.TypeID, it.Unk72, it.flnt = idx, ver, id, unk, flnt
	return true
}

// Diff lists the numeric columns that differ from other, as
// "Name:value/other" separated by tabs (diff).
func (it *Item) Diff(o *Item) string {
	var parts []string
	for k := 4; k < len(ItemColumnTags); k++ {
		tag := ItemColumnTags[k]
		if it.Compare(o, tag) != 0 {
			parts = append(parts, fmt.Sprintf("%s:%d/%d", ItemColumnNames[k], it.Get(tag), o.Get(tag)))
		}
	}
	return strings.Join(parts, "\t")
}

// Read reads the item in format f (Data::Item::read).
func (it *Item) Read(r *binstream.Reader, f Format) {
	start := r.Pos
	switch f {
	case Ltxt:
		it.Name = r.String(4, -1)
		it.Desc = unescape(r.String(4, -1))
	case File:
		it.Name = r.String(4, r.CodePage)
		it.Desc = unescape(r.String(4, r.CodePage))
		it.Type = r.Byte()
		it.TypeID = r.Byte()
		it.Model = r.Byte()
		it.Icon = r.Byte()
		it.Level = r.Int16()
		it.Country = r.Byte()
		for i := range it.Class {
			it.Class[i] = r.Byte()
		}
		it.Grow = r.Byte()
		it.Type2 = r.Byte()
		it.Type3 = r.Byte()
		it.ReqStr = r.Uint16()
		it.ReqDex = r.Uint16()
		it.ReqRec = r.Uint16()
		it.ReqInt = r.Uint16()
		it.ReqWis = r.Uint16()
		it.ReqLuc = r.Uint16()
		it.ReqVg = r.Int16()
		if it.Version >= 0x620 {
			it.Unk72 = r.Uint16()
		}
		it.ReqOg = r.Byte()
		it.ReqIg = r.Byte()
		if it.Version >= 0x600 {
			it.Range = r.Int16()
		} else {
			rng := r.Byte()
			it.Range = int16(rng)
			if it.Type == 44 || it.Type == 100 {
				if rng != 0 && rng < 45 {
					it.Range = int16(rng) + 256
				} else if rng == 0 && it.Type == 100 && (it.TypeID == 95 || it.TypeID == 96 || it.TypeID == 97) {
					it.Range = 256
				}
			}
		}
		it.AttackTime = r.Byte()
		it.Attrib = r.Byte()
		it.Special = r.Byte()
		it.Slot = r.Byte()
		it.Quality = r.Uint16()
		it.Attack = r.Uint16()
		it.AttackAdd = r.Uint16()
		it.Defense = r.Uint16()
		it.Resist = r.Uint16()
		it.HP = r.Uint16()
		it.SP = r.Uint16()
		it.MP = r.Uint16()
		it.Str = r.Uint16()
		it.Dex = r.Uint16()
		it.Rec = r.Uint16()
		it.Int = r.Uint16()
		it.Wis = r.Uint16()
		it.Luc = r.Uint16()
		it.Speed = r.Byte()
		it.Exp = r.Byte()
		it.Buy = r.Int32()
		it.Sell = r.Int32()
		it.Grade = r.Int16()
		it.Drop = r.Int16()
		it.Server = r.Byte()
		it.Count = r.Byte()
		if it.Version >= 0x600 {
			it.Duration = r.Int32()
			it.ExtDuration = r.Byte()
			it.SecOption = r.Byte()
			it.OptionRate = r.Byte()
			it.BuyMethod = r.Byte()
			it.MaxLevel = r.Byte()
			it.WeaponPart = r.Byte()
			it.DyeingType = r.Byte()
			it.Arg3 = r.Int32()
			it.Arg4 = r.Int32()
			it.UseConType = r.Int32()
			it.UseConVar = r.Int32()
			it.MoneyType = r.Int32()
			it.ItemSkill = r.Int32()
			it.ItemUpgrade = r.Int32()
			it.Arg10 = r.Int32()
			it.GeneCount = r.Byte()
			it.Arg12 = r.Byte()
			it.SpellBookExp = r.Byte()
			it.SpellBookDurability = r.Int32()
		}
	case Flnt:
		for _, p := range it.flntFields() {
			setInt(p, int(r.Int64()))
		}
		w := binstream.NewWriter(0)
		it.writeFlnt(w)
		it.flnt = rememberRow(r.Data[start:r.Pos], w.Bytes())
		return
	default:
		return
	}
	it.Valid = !(it.Name != "" && it.Name[0] == '?')
}

// flntFields lists the fields of a DB row, in order.
func (it *Item) flntFields() []any {
	f := []any{&it.Type, &it.TypeID, &it.Model, &it.Icon, &it.Level, &it.Country}
	for i := range it.Class {
		f = append(f, &it.Class[i])
	}
	return append(f, &it.Grow, &it.ReqStr, &it.ReqDex, &it.ReqRec, &it.ReqInt, &it.ReqWis, &it.ReqLuc,
		&it.ReqVg, &it.ReqOg, &it.ReqIg, &it.Range, &it.AttackTime, &it.Attrib, &it.Special, &it.Slot,
		&it.Quality, &it.Attack, &it.AttackAdd, &it.Defense, &it.Resist, &it.HP, &it.SP, &it.MP,
		&it.Str, &it.Dex, &it.Rec, &it.Int, &it.Wis, &it.Luc, &it.Speed, &it.Exp, &it.Buy, &it.Sell,
		&it.Grade, &it.Drop, &it.Server, &it.Count, &it.Duration, &it.ExtDuration, &it.SecOption,
		&it.OptionRate, &it.BuyMethod, &it.MaxLevel, &it.WeaponPart, &it.DyeingType, &it.Arg3, &it.Arg4,
		&it.UseConType, &it.UseConVar, &it.MoneyType, &it.ItemSkill, &it.ItemUpgrade, &it.Arg10,
		&it.GeneCount, &it.Arg12, &it.SpellBookExp, &it.SpellBookDurability)
}

func (it *Item) writeFlnt(w *binstream.Writer) {
	for _, p := range it.flntFields() {
		w.Int64(int64(getInt(p)))
	}
}

// Write writes the item in format f (Data::Item::write).
func (it *Item) Write(w *binstream.Writer, f Format) {
	switch f {
	case Ltxt:
		w.Int64(int64(it.Type))
		w.Int64(int64(it.TypeID))
		w.String(it.Name, 4, -1)
		w.String(escape(it.Desc), 4, -1)
	case File:
		w.String(it.Name, 4, w.CodePage)
		w.String(escape(it.Desc), 4, w.CodePage)
		w.Put([]byte{it.Type, it.TypeID, it.Model, it.Icon})
		w.Int16(it.Level)
		w.Byte(it.Country)
		w.Put(it.Class[:])
		w.Put([]byte{it.Grow, it.Type2, it.Type3})
		for _, v := range []uint16{it.ReqStr, it.ReqDex, it.ReqRec, it.ReqInt, it.ReqWis, it.ReqLuc, uint16(it.ReqVg)} {
			w.Uint16(v)
		}
		if it.Version >= 0x620 {
			w.Uint16(it.Unk72)
		}
		w.Put([]byte{it.ReqOg, it.ReqIg})
		if it.Version >= 0x600 {
			w.Int16(it.Range)
		} else {
			w.Byte(byte(it.Range))
		}
		w.Put([]byte{it.AttackTime, it.Attrib, it.Special, it.Slot})
		for _, v := range []uint16{it.Quality, it.Attack, it.AttackAdd, it.Defense, it.Resist, it.HP, it.SP,
			it.MP, it.Str, it.Dex, it.Rec, it.Int, it.Wis, it.Luc} {
			w.Uint16(v)
		}
		w.Put([]byte{it.Speed, it.Exp})
		w.Int32(it.Buy)
		w.Int32(it.Sell)
		w.Int16(it.Grade)
		w.Int16(it.Drop)
		w.Put([]byte{it.Server, it.Count})
		if it.Version >= 0x600 {
			w.Int32(it.Duration)
			w.Put([]byte{it.ExtDuration, it.SecOption, it.OptionRate, it.BuyMethod, it.MaxLevel, it.WeaponPart, it.DyeingType})
			for _, v := range []int32{it.Arg3, it.Arg4, it.UseConType, it.UseConVar, it.MoneyType, it.ItemSkill, it.ItemUpgrade, it.Arg10} {
				w.Int32(v)
			}
			w.Put([]byte{it.GeneCount, it.Arg12, it.SpellBookExp})
			w.Int32(it.SpellBookDurability)
		}
	case Flnt:
		start := w.Len()
		it.writeFlnt(w)
		it.flnt.restore(w.Buf[start:])
	}
}

// Line returns the item as a CSV or SQL line (toText); ok is false for an
// invalid item in CSV.
func (it *Item) Line(format string) (string, bool) {
	if format == TextSQL {
		v := []string{i32s(int32(it.TypeID) + 1000*int32(it.Type)), sqlQuote(it.Name), u8s(it.Type), u8s(it.TypeID),
			i16s(it.Level), u8s(it.Country)}
		for _, c := range it.Class {
			v = append(v, u8s(c))
		}
		// SF_USHORT_SIGNED: unsigned shorts are printed as signed values.
		s16 := func(x uint16) string { return i16s(int16(x)) }
		v = append(v, u8s(it.Grow), s16(it.ReqStr), s16(it.ReqDex), s16(it.ReqRec), s16(it.ReqInt), s16(it.ReqWis),
			s16(it.ReqLuc), i16s(it.ReqVg), u8s(it.ReqOg), u8s(it.ReqIg), i16s(it.Range), u8s(it.AttackTime),
			u8s(it.Attrib), u8s(it.Special), u8s(it.Slot), s16(it.Quality), s16(it.Attack), s16(it.AttackAdd),
			s16(it.Defense), s16(it.Resist), s16(it.HP), s16(it.SP), s16(it.MP), s16(it.Str), s16(it.Dex),
			s16(it.Rec), s16(it.Int), s16(it.Wis), s16(it.Luc), u8s(it.Speed), u8s(it.Exp), i32s(it.Buy),
			i32s(it.Sell), i16s(it.Grade), i16s(it.Drop), u8s(it.Server), u8s(it.Count))
		return "INSERT INTO [PS_GameDefs].[dbo].[Items] " +
			"([ItemID],[ItemName],[Type],[TypeID],[Reqlevel],[Country],[Attackfighter],[Defensefighter],[Patrolrogue],[Shootrog" +
			"ue],[Attackmage],[Defensemage],[Grow],[ReqStr],[ReqDex],[ReqRec],[ReqInt],[ReqWis],[Reqluc],[ReqVg],[ReqOg],[ReqI" +
			"g],[Range],[AttackTime],[Attrib],[Special],[Slot],[Quality],[Effect1],[Effect2],[Effect3],[Effect4],[ConstHP],[Co" +
			"nstSP],[ConstMP],[ConstStr],[ConstDex],[ConstRec],[ConstInt],[ConstWis],[ConstLuc],[Speed],[Exp],[Buy],[Sell],[Gr" +
			"ade],[Drop],[Server],[Count]) values (" + strings.Join(v, ",") + ");\r\n", true
	}
	if !it.Valid && format != TextCSVE {
		return "", false
	}
	v := []string{u8s(it.Type), u8s(it.TypeID), `"` + it.Name + `"`, `"` + escape(it.Desc) + `"`,
		u8s(it.Model), u8s(it.Icon), i16s(it.Level), u8s(it.Country)}
	for _, c := range it.Class {
		v = append(v, u8s(c))
	}
	v = append(v, u8s(it.Grow), u8s(it.Type2), u8s(it.Type3), u16s(it.ReqStr), u16s(it.ReqDex), u16s(it.ReqRec),
		u16s(it.ReqInt), u16s(it.ReqWis), u16s(it.ReqLuc), i16s(it.ReqVg), u8s(it.ReqOg), u8s(it.ReqIg),
		i16s(it.Range), u8s(it.AttackTime), u8s(it.Attrib), u8s(it.Special), u8s(it.Slot), u16s(it.Quality),
		u16s(it.Attack), u16s(it.AttackAdd), u16s(it.Defense), u16s(it.Resist), u16s(it.HP), u16s(it.SP),
		u16s(it.MP), u16s(it.Str), u16s(it.Dex), u16s(it.Rec), u16s(it.Int), u16s(it.Wis), u16s(it.Luc),
		u8s(it.Speed), u8s(it.Exp), i32s(it.Buy), i32s(it.Sell), i16s(it.Grade), i16s(it.Drop),
		u8s(it.Server), u8s(it.Count))
	return strings.Join(v, ";") + "\r\n", true
}

// ClassFromFlags maps class flags to a class index (classFromFlags).
func ClassFromFlags(flags int16) int {
	switch {
	case flags&1 != 0 || flags&2 != 0:
		return 0
	case flags&4 != 0 || flags&8 != 0:
		return 2
	case flags&0x20 != 0:
		return 3
	case flags&0x10 != 0:
		return 1
	case flags&0x100 != 0 || flags&0x200 != 0:
		return 4
	case flags&0x400 != 0:
		return 6
	case flags&0x800 != 0:
		return 5
	case flags&0x2000 != 0:
		return 7
	case flags&0x1000 != 0:
		return 7
	}
	return -2
}

// BaseType maps the faction/episode variants of a weapon type to its base
// type (baseType).
func BaseType(t uint8) uint8 {
	switch t {
	case 1, 45:
		return 1
	case 2, 46:
		return 2
	case 3, 47:
		return 3
	case 4, 48:
		return 4
	case 5, 49, 50:
		return 5
	case 6, 51, 52:
		return 6
	case 7, 53, 54:
		return 7
	case 8, 55, 56:
		return 8
	case 9, 57:
		return 9
	case 10, 58:
		return 10
	case 11, 59:
		return 11
	case 12, 60, 61:
		return 12
	case 13, 62, 63:
		return 13
	case 14, 64:
		return 14
	case 15, 65:
		return 15
	case 19, 69, 75:
		return 19
	case 34, 84, 90:
		return 34
	}
	return t
}

// IsGearType reports whether t is an armor piece of the light (second) or
// fury side (isGearType).
func IsGearType(t uint8, second bool) bool {
	in := func(base uint8, skip uint8) bool { return t-base <= 5 && t != skip }
	if second {
		return in(16, 19) || in(66, 69) || in(72, 75)
	}
	return in(31, 34) || in(81, 84) || in(87, 90)
}

// MatchesType reports whether the item is of the armor kind t in any
// episode variant (matchesType).
func MatchesType(it *Item, t uint8) bool {
	x := it.Type
	switch t {
	case 16:
		return x == 16 || x == 66 || x == 72
	case 17:
		return x == 17 || x == 67 || x == 73
	case 18:
		return x == 18 || x == 68 || x == 74
	case 20:
		return x == 20 || x == 70 || x == 76
	case 21:
		return x == 21 || x == 71 || x == 77
	case 31:
		return x == 31 || x == 81 || x == 87
	case 32:
		return x == 32 || x == 82 || x == 88
	case 33:
		return x == 33 || x == 83 || x == 89
	case 35:
		return x == 35 || x == 85 || x == 91
	case 36:
		return x == 36 || x == 86 || x == 92
	}
	return false
}

// EquipSlot returns the equipment slot of an armor type (equipSlot).
func EquipSlot(t int) int {
	switch t {
	case 0x10, 0x1F, 0x42, 0x48, 0x51, 0x57:
		return 4
	case 0x11, 0x20, 0x43, 0x49, 0x52, 0x58:
		return 6
	case 0x12, 0x21, 0x44, 0x4A, 0x53, 0x59:
		return 5
	case 0x14, 0x23, 0x46, 0x4C, 0x55, 0x5B:
		return 3
	case 0x15, 0x24, 0x47, 0x4D, 0x56, 0x5C:
		return 1
	}
	return -1
}
