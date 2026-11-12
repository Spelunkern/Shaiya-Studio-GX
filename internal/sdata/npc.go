package sdata

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// NpcOptions are the NPCQuest settings that change how NpcQuest.SData is
// read ("NPCQuest" section of the settings).
type NpcOptions struct {
	CleanCR       bool // "clean CR": trim trailing CR LF and spaces of texts
	CleanMerchant bool // "clean merchant": drop empty merchant slots
}

// DefaultNpcOptions are the defaults of shStudio (both on).
var DefaultNpcOptions = NpcOptions{CleanCR: true, CleanMerchant: true}

// NpcTypeName names an NPC type (StandardNPC::typeName).
func NpcTypeName(t int) string {
	switch t {
	case 1:
		return "Merchant"
	case 2:
		return "Gate-Keeper"
	case 3:
		return "Blacksmith"
	case 4:
		return "PvP-Manager"
	case 5:
		return "Gambling House"
	case 6:
		return "Warehouse"
	case 7:
		return "Normal NPC"
	case 8:
		return "Guard"
	case 9:
		return "Animal"
	case 10:
		return "Apprentice"
	case 11:
		return "Guild-Master"
	case 12:
		return "Dead NPC"
	case 13:
		return "Combat Commander"
	}
	return "Unknown"
}

// MerchantTypeNames are the merchant professions (g_merchantTypeNames).
var MerchantTypeNames = []string{
	"Generic Merchant", "Weapon Merchant", "Protection Merchant", "Accessory Merchant",
	"Liquid Medecine Merchant", "Professional Trader", "Riding Merchant", "Mantle Merchant",
	"Badge Merchant", "Bank Teller", "Gold Bar Merchant", "Episo 6.x Merchant",
	"Professional Blacksmith 1", "Professional Blacksmith 2", "Etin Manager", "Etin Seller",
	"Action Board", "PvP Quests",
}

// MerchantProfIndex maps a profession to MerchantTypeNames (profIndex).
func MerchantProfIndex(prof uint8) int {
	switch {
	case prof <= 11:
		return int(prof)
	case prof == 15:
		return 12
	case prof == 16:
		return 13
	case prof == 100:
		return 14
	case prof == 101:
		return 15
	case prof == 117:
		return 16
	}
	return 17
}

// GateTarget is a destination of a gate keeper.
type GateTarget struct {
	Name    string
	MapID   int16
	X, Y, Z float32
	Cost    int32
}

// NPC is an NPC of NpcQuest.SData (StandardNPC, Merchant, GateKeeper).
type NPC struct {
	Type      int
	ID        int16
	Shape     int32
	Unk1      int32
	Unk2      int32
	Country   int32
	Name      string
	Welcome   string
	QuestsIn  []int16
	QuestsOut []int16

	// Merchant (type 1).
	Prof      uint8
	Items     []int32 // type*1000 + id
	LastEmpty bool
	// Gate keeper (type 2).
	Targets [3]GateTarget
}

// NewNPC returns an NPC of a type.
func NewNPC(t int) *NPC { return &NPC{Type: t} }

// Reset clears the NPC but keeps type and id.
func (n *NPC) Reset() { *n = NPC{Type: n.Type, ID: n.ID} }

// QuestsInText lists the incoming quests ("1,2,3"; "0" when empty).
func (n *NPC) QuestsInText() string { return joinShorts(n.QuestsIn) }

// QuestsOutText lists the outgoing quests.
func (n *NPC) QuestsOutText() string { return joinShorts(n.QuestsOut) }

func joinShorts(v []int16) string {
	if len(v) == 0 {
		return "0"
	}
	p := make([]string, len(v))
	for i, x := range v {
		p[i] = fmt.Sprint(x)
	}
	return strings.Join(p, ",")
}

// ItemsText lists the merchant items ("0" when empty).
func (n *NPC) ItemsText() string {
	if len(n.Items) == 0 {
		return "0"
	}
	p := make([]string, len(n.Items))
	for i, x := range n.Items {
		p[i] = fmt.Sprint(x)
	}
	return strings.Join(p, ",")
}

// Text returns a text column.
func (n *NPC) Text(tag Tag) string {
	switch tag {
	case "name":
		return n.Name
	case "welc":
		return n.Welcome
	case "fami":
		return NpcTypeName(n.Type)
	case "qsti":
		return n.QuestsInText()
	case "qsto":
		return n.QuestsOutText()
	case "sitm":
		if n.Type == 1 {
			return n.ItemsText()
		}
	case "prof":
		if n.Type == 1 {
			return MerchantTypeNames[MerchantProfIndex(n.Prof)]
		}
	case "loc1", "loc2", "loc3":
		if n.Type == 2 {
			return n.Targets[tag[3]-'1'].Name
		}
	}
	return ""
}

// SetText sets the name, welcome text or a gate name.
func (n *NPC) SetText(tag Tag, v string) bool {
	switch tag {
	case "name":
		n.Name = v
	case "welc":
		n.Welcome = v
	case "loc1", "loc2", "loc3":
		if n.Type != 2 {
			return false
		}
		n.Targets[tag[3]-'1'].Name = v
	default:
		return false
	}
	return true
}

// Get returns a numeric column.
func (n *NPC) Get(tag Tag) int {
	switch tag {
	case "shpe":
		return int(n.Shape)
	case "kqou":
		return len(n.QuestsOut)
	case "kqin":
		return len(n.QuestsIn)
	case "cntr":
		return int(n.Country)
	case "fami", "type":
		return n.Type
	case "unk1":
		return int(n.Unk1)
	case "uid ":
		return int(n.ID)
	case "unk2":
		return int(n.Unk2)
	case "prof":
		if n.Type == 1 {
			return int(n.Prof)
		}
	case "kitm":
		if n.Type == 1 {
			return len(n.Items)
		}
	}
	return 0
}

// Set sets a numeric column.
func (n *NPC) Set(tag Tag, v int) bool {
	switch tag {
	case "type":
		n.Type = v
	case "uid ":
		n.ID = int16(v)
	case "shpe":
		n.Shape = int32(v)
	case "cntr":
		n.Country = int32(v)
	case "prof":
		if n.Type != 1 {
			return false
		}
		n.Prof = uint8(v)
	default:
		return false
	}
	return true
}

// Compare compares two NPCs on tag.
func (n *NPC) Compare(o *NPC, tag Tag) int {
	switch tag {
	case "type", "fami":
		return cmp(n.Type, o.Type)
	case "uid ":
		return cmp(n.ID, o.ID)
	case "name":
		return compareLogical(n.Name, o.Name)
	case "shpe":
		return cmp(n.Shape, o.Shape)
	case "cntr":
		return cmp(n.Country, o.Country)
	case "kqin":
		return cmp(len(n.QuestsIn), len(o.QuestsIn))
	case "kqou":
		return cmp(len(n.QuestsOut), len(o.QuestsOut))
	case "prof":
		return cmp(n.Prof, o.Prof)
	case "kitm":
		return cmp(len(n.Items), len(o.Items))
	case "loc1", "loc2", "loc3":
		i := tag[3] - '1'
		return compareLogical(n.Targets[i].Name, o.Targets[i].Name)
	}
	return 0
}

// Assign copies src into n except the type and id.
func (n *NPC) Assign(src *NPC) {
	t, id := n.Type, n.ID
	*n = *src
	n.Type, n.ID = t, id
	n.QuestsIn = append([]int16(nil), src.QuestsIn...)
	n.QuestsOut = append([]int16(nil), src.QuestsOut...)
	n.Items = append([]int32(nil), src.Items...)
	if src.Type != 1 || t != 1 {
		n.Items, n.Prof = nil, 0
	}
}

func (n *NPC) trimStrings() {
	n.Name = cleanText(n.Name)
	n.Welcome = cleanText(n.Welcome)
}

func (n *NPC) readMain(r *binstream.Reader, f Format) {
	switch f {
	case File, Flnt:
		t := r.Byte()
		n.ID = r.Int16()
		if n.Type == 1 {
			n.Prof = r.Byte()
		}
		n.Shape = r.Int32()
		n.Unk1 = r.Int32()
		n.Unk2 = r.Int32()
		n.Country = r.Int32()
		n.Type = int(t)
		if f == File {
			n.Name = r.String(4, r.CodePage)
			n.Welcome = unescape(r.String(4, r.CodePage))
		}
	case Ltxt:
		n.Name = r.String(4, -1)
		n.Welcome = unescape(r.String(4, -1))
	}
}

func (n *NPC) writeMain(w *binstream.Writer, f Format) {
	switch f {
	case File, Flnt:
		w.Byte(byte(n.Type))
		w.Int16(n.ID)
		if n.Type == 1 {
			w.Byte(n.Prof)
		}
		w.Int32(n.Shape)
		w.Int32(n.Unk1)
		w.Int32(n.Unk2)
		w.Int32(n.Country)
		if f == File {
			name := n.Name
			if n.Type != 1 {
				name = escape(name)
			}
			w.String(name, 4, w.CodePage)
			w.String(escape(n.Welcome), 4, w.CodePage)
		}
	case Ltxt:
		name := n.Name
		if n.Type != 1 {
			name = escape(name)
		}
		w.String(name, 4, -1)
		w.String(escape(n.Welcome), 4, -1)
	}
}

func (n *NPC) readLists(r *binstream.Reader, f Format) {
	if f == Ltxt {
		return
	}
	c := r.Uint32()
	if c > 0x3E7 {
		binstream.Fail("too many quests")
	}
	for ; c > 0; c-- {
		n.QuestsIn = append(n.QuestsIn, r.Int16())
	}
	m := r.Int()
	if m < 0 {
		binstream.Fail("negative quest count")
	}
	for ; m > 0; m-- {
		n.QuestsOut = append(n.QuestsOut, r.Int16())
	}
}

func (n *NPC) writeLists(w *binstream.Writer, f Format) {
	if f == Ltxt {
		return
	}
	w.Uint32(uint32(len(n.QuestsIn)))
	for _, q := range n.QuestsIn {
		w.Int16(q)
	}
	w.Uint32(uint32(len(n.QuestsOut)))
	for _, q := range n.QuestsOut {
		w.Int16(q)
	}
}

func (g *GateTarget) read(r *binstream.Reader, f Format) {
	switch f {
	case File, Flnt:
		g.MapID = r.Int16()
		g.X, g.Y, g.Z = r.Float32(), r.Float32(), r.Float32()
		if f == File {
			g.Name = r.String(4, r.CodePage)
		}
		g.Cost = r.Int32()
	case Ltxt:
		g.Name = r.String(4, -1)
	}
}

func (g *GateTarget) write(w *binstream.Writer, f Format) {
	switch f {
	case File, Flnt:
		w.Int16(g.MapID)
		w.Float32(g.X)
		w.Float32(g.Y)
		w.Float32(g.Z)
		if f == File {
			w.String(g.Name, 4, w.CodePage)
		}
		w.Int32(g.Cost)
	case Ltxt:
		w.String(g.Name, 4, -1)
	}
}

// Read reads the NPC (StandardNPC/Merchant/GateKeeper::read).
func (n *NPC) Read(r *binstream.Reader, f Format, opt NpcOptions) {
	switch n.Type {
	case 1:
		if f == Ltxt {
			n.Name = r.String(4, r.CodePage)
			n.Welcome = unescape(r.String(4, r.CodePage))
			return
		}
		n.readMain(r, f)
		count := r.Int()
		for i := 0; i < count; i++ {
			v := int32(r.Byte())*1000 + int32(r.Byte())
			switch {
			case v != 0:
				n.Items = append(n.Items, v)
			case opt.CleanMerchant:
				if i == count-1 {
					n.LastEmpty = true
				}
			default:
				n.Items = append(n.Items, 0)
			}
		}
		n.readLists(r, f)
	case 2:
		n.readMain(r, f)
		for i := range n.Targets {
			n.Targets[i].read(r, f)
		}
		n.readLists(r, f)
	default:
		n.readMain(r, f)
		n.readLists(r, f)
	}
}

// Write writes the NPC.
func (n *NPC) Write(w *binstream.Writer, f Format) {
	n.writeMain(w, f)
	switch n.Type {
	case 1:
		if f != Ltxt {
			count := len(n.Items)
			if n.LastEmpty {
				count++
			}
			w.Uint32(uint32(count))
			for _, it := range n.Items {
				hi := it / 1000
				w.Byte(byte(hi))
				w.Byte(byte(it + 24*hi))
			}
			if n.LastEmpty {
				w.Uint16(0)
			}
		}
	case 2:
		for i := range n.Targets {
			n.Targets[i].write(w, f)
		}
	}
	n.writeLists(w, f)
}

// Line exports the NPC as text (toText).
func (n *NPC) Line(format string) string {
	sep := exportSeparator(format)
	return fmt.Sprintf("%s\r\n%s\r\n%d%s%d%s%d%s%d%s%d%s%d", n.Name, n.Welcome, int32(n.Type), sep, n.ID, sep,
		n.Shape, sep, n.Unk1, sep, n.Unk2, sep, n.Country)
}

// NpcQuest is NpcQuest.SData and NpcQuestTrans_<LANG>.SData (NpcData):
// the NPCs of the 13 types, an opaque block and the quests.
type NpcQuest struct {
	NPCs    [14][]*NPC // index = type, 1..13
	Raw     []byte     // bytes between the NPC lists and the quests
	Quests  Quests
	Format  int
	IsNew   bool
	Options NpcOptions
}

// NewNpcQuest returns an empty table with the default options.
func NewNpcQuest() *NpcQuest { return &NpcQuest{IsNew: true, Options: DefaultNpcOptions} }

// FilePrefix is the prefix of the language files.
func (d *NpcQuest) FilePrefix() string { return `Npc\NpcQuestTrans_` }

// TextCodePage returns the code page of a language file name
// (NpcData::getCodePage), or -1.
func TextCodePage(fileName string) int {
	base := strings.ToUpper(strings.TrimSuffix(fileName, fileExt(fileName)))
	switch {
	case hasAnySuffix(base, "FRANCE", "GERMANY", "ITALY", "USA"):
		return 1252
	case hasAnySuffix(base, "POLAND", "TURKEY"):
		return 1250
	case strings.HasSuffix(base, "CHINA"):
		return 936
	case strings.HasSuffix(base, "KOREA"):
		return 949
	}
	return -1
}

func fileExt(s string) string {
	if i := strings.LastIndexByte(s, '.'); i >= 0 && !strings.ContainsAny(s[i:], `\/`) {
		return s[i:]
	}
	return ""
}

func hasAnySuffix(s string, suffixes ...string) bool {
	for _, x := range suffixes {
		if strings.HasSuffix(s, x) {
			return true
		}
	}
	return false
}

// Find returns the NPC of a type and id (NpcPerso::find).
func (d *NpcQuest) Find(t uint8, id int16) *NPC {
	if t < 1 || t > 13 {
		return nil
	}
	for _, n := range d.NPCs[t] {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// SetVersion sets the quest format (setCodePage).
func (d *NpcQuest) SetVersion(v int) {
	d.Format = v
	d.Quests.Version = v
	for _, q := range d.Quests.Quests {
		q.Version = v
	}
}

func (d *NpcQuest) readList(r *binstream.Reader, t int, f Format) {
	if f == Ltxt {
		n := r.Int()
		for i := 0; i < n; i++ {
			if i < len(d.NPCs[t]) {
				d.NPCs[t][i].Read(r, Ltxt, d.Options)
			} else {
				NewNPC(t).Read(r, Ltxt, d.Options)
			}
		}
		return
	}
	d.NPCs[t] = nil
	n := r.Int()
	for i := 0; i < n; i++ {
		npc := NewNPC(t)
		npc.Read(r, f, d.Options)
		d.NPCs[t] = append(d.NPCs[t], npc)
	}
}

// Read loads the table (NpcData::read). Ltxt merges a language file into
// the loaded NPCs and quests.
func (d *NpcQuest) Read(data []byte, cp int, f Format) error {
	if f != Ltxt {
		d.NPCs = [14][]*NPC{}
		d.Raw = nil
		d.Quests = Quests{}
		d.Format = 0
	}
	err := binstream.Try(func() {
		r := binstream.NewReader(data, cp)
		for t := 1; t < 14; t++ {
			d.readList(r, t, f)
			if f != Flnt && d.Options.CleanCR {
				for _, n := range d.NPCs[t] {
					n.trimStrings()
				}
			}
		}
		d.IsNew = false
		if f != Ltxt {
			start := r.Pos
			for {
				for uint16(r.Int16()) <= 1000 {
				}
				a, b := r.Int16(), r.Int16()
				if a == 0 && b == 1 {
					break
				}
			}
			if r.Pos >= 6 {
				r.Pos -= 6
			}
			if r.Len() < 4 {
				binstream.Fail("quests not found")
			}
			size := r.Pos - start
			r.Pos = start
			d.Raw = r.Copy(size)
		}
		d.Quests.read(r, f, d.Options)
		if f != Ltxt {
			d.Format = d.Quests.Version
		}
	})
	return err
}

// Write serializes the table (NpcData::write). Strings are written without
// NUL; NPCs and quests are sorted by id first (except for Ltxt).
func (d *NpcQuest) Write(cp int, f Format) ([]byte, error) {
	if d.IsNew && f != Ltxt {
		return nil, errors.New("npcquest: empty")
	}
	w := binstream.NewWriter(cp)
	w.WriteNul = false
	if f != Ltxt {
		SortList(d.NPCs[1], "uid ", "", true)
		SortList(d.NPCs[2], "uid ", "", true)
		SortList(d.Quests.Quests, "quid", "", true)
	}
	for t := 1; t < 14; t++ {
		w.Uint32(uint32(len(d.NPCs[t])))
		for _, n := range d.NPCs[t] {
			n.Write(w, f)
		}
	}
	if f != Ltxt {
		w.Put(d.Raw)
	}
	d.Quests.write(w, f)
	return w.Bytes(), nil
}
