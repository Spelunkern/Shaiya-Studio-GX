package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// MonsterAttack is one of the three attacks of a monster (Monster::Attack).
type MonsterAttack struct {
	Ani     uint8
	Type    int16
	Time    int32
	Range   uint8
	Damage  uint16
	Add     uint16
	Attrib  uint8
	Special uint8
	Mode    uint8
}

func (a *MonsterAttack) read(r *binstream.Reader, f Format) {
	switch f {
	case File:
		a.Type = r.Int16()
		a.Time = r.Int32()
		a.Range = r.Byte()
		a.Damage = r.Uint16()
		a.Add = r.Uint16()
		a.Attrib = r.Byte()
		a.Special = r.Byte()
		a.Mode = r.Byte()
	case Flnt:
		a.Ani = uint8(r.Int64())
		a.Type = int16(r.Int64())
		a.Time = int32(r.Int64())
		a.Range = uint8(r.Int64())
		a.Damage = uint16(r.Int64())
		a.Add = uint16(r.Int64())
		a.Attrib = uint8(r.Int64())
		a.Special = uint8(r.Int64())
		a.Mode = uint8(r.Int64())
	}
}

func (a *MonsterAttack) write(w *binstream.Writer, f Format) {
	switch f {
	case File:
		w.Int16(a.Type)
		w.Int32(a.Time)
		w.Byte(a.Range)
		w.Uint16(a.Damage)
		w.Uint16(a.Add)
		w.Put([]byte{a.Attrib, a.Special, a.Mode})
	case Flnt:
		for _, v := range []int64{int64(a.Ani), int64(a.Type), int64(a.Time), int64(a.Range), int64(a.Damage),
			int64(a.Add), int64(a.Attrib), int64(a.Special), int64(a.Mode)} {
			w.Int64(v)
		}
	}
}

// text formats an attack for the SQL and CSV exports (sub_43D120).
func (a *MonsterAttack) text() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d", a.Type, a.Time, a.Range, int16(a.Damage), int16(a.Add), a.Attrib, a.Special, a.Mode)
}

// Monster is a row of Monster.SData / MonsterEx.SData / DBMonsterData
// (Monster).
type Monster struct {
	Version    int
	Extended   bool
	ID         int32
	Name       string
	Model      int16
	Unk3A      int16
	Level      int16
	Exp        uint16
	AI         uint8
	Gold1      uint16
	Gold2      uint16
	DropGroup  [9]uint16
	DropRate   [9]int32
	Quest      int16
	HP         int32
	SP, MP     uint16
	Dex        uint16
	Wis, Luc   uint16
	Day, Size  uint8
	Attrib     uint8
	Defense    uint16
	Magic      uint16
	Resist     [15]uint8
	RK         [6]uint8
	NormalTime int32
	NormalStep uint8
	ChaseTime  int32
	ChaseStep  uint8
	ChaseRange uint8
	Attacks    [3]MonsterAttack
	ColorType  int16
	ColorHue   uint8
	ColorSat   uint8
	ColorLight uint8

	flnt *flntRow
}

// NewMonster returns a reset monster.
func NewMonster(id int32, version int) *Monster {
	m := &Monster{ID: id, Version: version}
	m.Reset()
	return m
}

// Reset clears the monster; its name becomes "Error Monster<id>".
func (m *Monster) Reset() {
	*m = Monster{ID: m.ID, Version: m.Version, Extended: m.Extended, Unk3A: m.Unk3A, flnt: m.flnt}
	m.Name = fmt.Sprintf("Error Monster%d", m.ID)
	m.ColorSat = 0x32
}

// AttribName names the element (getString 'attr').
func (m *Monster) AttribName() string {
	switch m.Attrib {
	case 1:
		return "Fire"
	case 2:
		return "Water"
	case 3:
		return "Earth"
	case 4:
		return "Wind"
	}
	return "None"
}

// Text returns "name" or "attr".
func (m *Monster) Text(tag Tag) string {
	switch tag {
	case "name":
		return m.Name
	case "attr":
		return m.AttribName()
	}
	return ""
}

// SetText sets the name.
func (m *Monster) SetText(tag Tag, v string) bool {
	if tag != "name" {
		return false
	}
	m.Name = v
	return true
}

func (m *Monster) ptr(tag Tag) any {
	switch tag {
	case "mbid":
		return &m.ID
	case "modl":
		return &m.Model
	case "lvl ":
		return &m.Level
	case "exp_":
		return &m.Exp
	case "ai  ":
		return &m.AI
	case "gld1":
		return &m.Gold1
	case "gld2":
		return &m.Gold2
	case "qust":
		return &m.Quest
	case "hp  ":
		return &m.HP
	case "sp  ":
		return &m.SP
	case "mp  ":
		return &m.MP
	case "dex ":
		return &m.Dex
	case "wis ":
		return &m.Wis
	case "luc ":
		return &m.Luc
	case "day ":
		return &m.Day
	case "size":
		return &m.Size
	case "attr":
		return &m.Attrib
	case "defs":
		return &m.Defense
	case "magc":
		return &m.Magic
	case "ntme":
		return &m.NormalTime
	case "nstp":
		return &m.NormalStep
	case "ctme":
		return &m.ChaseTime
	case "cstp":
		return &m.ChaseStep
	case "crng":
		return &m.ChaseRange
	case "ctyp":
		return &m.ColorType
	case "chue":
		return &m.ColorHue
	case "csat":
		return &m.ColorSat
	case "clgt":
		return &m.ColorLight
	}
	if len(tag) == 4 {
		var n int
		switch {
		case tag[:2] == "rs" && sscan(tag[2:], &n) && n >= 1 && n <= 15:
			return &m.Resist[n-1]
		case tag[:2] == "rk" && sscan(tag[2:], &n) && n >= 1 && n <= 6:
			return &m.RK[n-1]
		case tag[:3] == "drp" && tag[3] >= '1' && tag[3] <= '9':
			return &m.DropGroup[tag[3]-'1']
		case tag[:3] == "drt" && tag[3] >= '1' && tag[3] <= '9':
			return &m.DropRate[tag[3]-'1']
		case tag[0] >= '1' && tag[0] <= '3':
			a := &m.Attacks[tag[0]-'1']
			switch tag[1:] {
			case "ani":
				return &a.Ani
			case "typ":
				return &a.Type
			case "tme":
				return &a.Time
			case "rng":
				return &a.Range
			case "dmg":
				return &a.Damage
			case "add":
				return &a.Add
			case "atr":
				return &a.Attrib
			case "spc":
				return &a.Special
			case "mod":
				return &a.Mode
			}
		}
	}
	return nil
}

func sscan(s Tag, n *int) bool {
	_, err := fmt.Sscanf(string(s), "%02d", n)
	return err == nil
}

// Get returns a numeric column (getInt; the attack "ani" columns read 0).
func (m *Monster) Get(tag Tag) int {
	if len(tag) == 4 && tag[1:] == "ani" {
		return 0
	}
	return getInt(m.ptr(tag))
}

// Set sets a numeric column (setInt; "rwid" is accepted and ignored).
func (m *Monster) Set(tag Tag, v int) bool {
	if tag == "rwid" {
		return true
	}
	return setInt(m.ptr(tag), v)
}

// Compare compares on tag ("rwid" and "mbid" compare ids).
func (m *Monster) Compare(o *Monster, tag Tag) int {
	switch tag {
	case "rwid", "mbid":
		return cmp(m.ID, o.ID)
	case "name":
		return compareLogical(m.Name, o.Name)
	}
	a := m.ptr(tag)
	if a == nil {
		return 0
	}
	return cmp(getInt(a), getInt(o.ptr(tag)))
}

// Assign copies src into m (assignFrom) except the id, version, Unk3A and
// the attack animations.
func (m *Monster) Assign(src *Monster) {
	keep := *m
	*m = *src
	m.ID, m.Version, m.Extended, m.Unk3A, m.flnt = keep.ID, keep.Version, keep.Extended, keep.Unk3A, keep.flnt
	for k := range m.Attacks {
		m.Attacks[k].Ani = keep.Attacks[k].Ani
	}
}

// Read reads a monster (Monster::read). For Flnt, lead is the position of
// the id cell that the table read before the row.
func (m *Monster) Read(r *binstream.Reader, f Format) {
	start := r.Pos - 8
	switch f {
	case File:
		if m.Extended {
			m.ID = r.Int32()
			m.Name = r.String(4, -1)
			m.Model = r.Int16()
			m.Unk3A = r.Int16()
			m.Level = r.Int16()
			m.Exp = r.Uint16()
			m.AI = r.Byte()
			m.Gold1 = r.Uint16()
			m.Gold2 = r.Uint16()
			m.Quest = r.Int16()
			m.HP = r.Int32()
			m.SP = r.Uint16()
			m.MP = r.Uint16()
			m.Dex = r.Uint16()
			m.Wis = r.Uint16()
			m.Luc = r.Uint16()
			m.Day = r.Byte()
			m.Size = r.Byte()
			m.Attrib = r.Byte()
			m.Defense = r.Uint16()
			m.Magic = r.Uint16()
			for i := range m.Resist {
				m.Resist[i] = r.Byte()
			}
			for i := range m.RK {
				m.RK[i] = r.Byte()
			}
			m.NormalTime = r.Int32()
			m.NormalStep = r.Byte()
			m.ChaseTime = r.Int32()
			m.ChaseStep = r.Byte()
			m.ChaseRange = r.Byte()
			for k := range m.Attacks {
				m.Attacks[k].read(r, File)
			}
			count := r.Int()
			for i := 0; i < count; i++ {
				g, rate := r.Int16(), r.Int32()
				if i < len(m.DropGroup) {
					m.DropGroup[i], m.DropRate[i] = uint16(g), rate
				}
			}
		} else {
			m.Name = r.String(4, -1)
			m.Model = r.Int16()
			m.Level = r.Int16()
			m.AI = r.Byte()
			m.HP = r.Int32()
			m.Day = r.Byte()
			m.Size = r.Byte()
			m.Attrib = r.Byte()
			m.NormalTime = r.Int32()
			m.NormalStep = r.Byte()
			m.ChaseTime = r.Int32()
			m.ChaseStep = r.Byte()
			b := r.Bytes(7)
			m.Quest = r.Int16()
			m.Attacks[0].Mode = b[1]
			m.Attacks[0].Type = int16(b[0])
			m.Attacks[1].Range = b[3]
			m.Attacks[1].Type = int16(b[2])
			// shStudio reads b[2] here and ignores b[4], which it then
			// overwrites with this value on save; b[4] is what it writes
			// back, so it is read from there (see docs/DIFFERENCES.md).
			m.Attacks[2].Type = int16(b[4])
			m.Attacks[2].Mode = b[5]
			m.Attacks[2].Add = uint16(b[6])
		}
		if m.Version&0xFFF >= 0x620 {
			m.ColorType = r.Int16()
			m.ColorHue = r.Byte()
			m.ColorSat = r.Byte()
			m.ColorLight = r.Byte()
		}
	case Flnt:
		m.Model = int16(r.Int64())
		m.Level = int16(r.Int64())
		m.Exp = uint16(r.Int64())
		m.AI = uint8(r.Int64())
		m.Gold1 = uint16(r.Int64())
		m.Gold2 = uint16(r.Int64())
		for i := range m.DropGroup {
			m.DropGroup[i] = uint16(r.Int64())
			m.DropRate[i] = int32(r.Int64())
		}
		m.Quest = int16(r.Int64())
		m.HP = int32(r.Int64())
		m.SP = uint16(r.Int64())
		m.MP = uint16(r.Int64())
		m.Dex = uint16(r.Int64())
		m.Wis = uint16(r.Int64())
		m.Luc = uint16(r.Int64())
		m.Day = uint8(r.Int64())
		m.Size = uint8(r.Int64())
		m.Attrib = uint8(r.Int64())
		m.Defense = uint16(r.Int64())
		m.Magic = uint16(r.Int64())
		for i := range m.Resist {
			m.Resist[i] = uint8(r.Int64())
		}
		for i := range m.RK {
			m.RK[i] = uint8(r.Int64())
		}
		m.NormalTime = int32(r.Int64())
		m.NormalStep = uint8(r.Int64())
		m.ChaseTime = int32(r.Int64())
		m.ChaseStep = uint8(r.Int64())
		m.ChaseRange = uint8(r.Int64())
		for k := range m.Attacks {
			m.Attacks[k].read(r, Flnt)
		}
		m.ColorType = int16(r.Int64())
		m.ColorHue = uint8(r.Int64())
		m.ColorSat = uint8(r.Int64())
		m.ColorLight = uint8(r.Int64())
		if start >= 0 {
			w := binstream.NewWriter(0)
			m.writeFlnt(w)
			m.flnt = rememberRow(r.Data[start:r.Pos], w.Bytes())
		}
	case Ltxt:
		m.Name = r.String(4, -1)
	}
}

func (m *Monster) writeFlnt(w *binstream.Writer) {
	w.Int64(int64(m.ID))
	for _, v := range []int64{int64(m.Model), int64(m.Level), int64(m.Exp), int64(m.AI), int64(m.Gold1), int64(m.Gold2)} {
		w.Int64(v)
	}
	for i := range m.DropGroup {
		w.Int64(int64(m.DropGroup[i]))
		w.Int64(int64(m.DropRate[i]))
	}
	for _, v := range []int64{int64(m.Quest), int64(m.HP), int64(m.SP), int64(m.MP), int64(m.Dex), int64(m.Wis),
		int64(m.Luc), int64(m.Day), int64(m.Size), int64(m.Attrib), int64(m.Defense), int64(m.Magic)} {
		w.Int64(v)
	}
	for _, v := range m.Resist {
		w.Int64(int64(v))
	}
	for _, v := range m.RK {
		w.Int64(int64(v))
	}
	for _, v := range []int64{int64(m.NormalTime), int64(m.NormalStep), int64(m.ChaseTime), int64(m.ChaseStep), int64(m.ChaseRange)} {
		w.Int64(v)
	}
	for k := range m.Attacks {
		m.Attacks[k].write(w, Flnt)
	}
	for _, v := range []int64{int64(m.ColorType), int64(m.ColorHue), int64(m.ColorSat), int64(m.ColorLight)} {
		w.Int64(v)
	}
}

// Write writes a monster (Monster::write).
func (m *Monster) Write(w *binstream.Writer, f Format) {
	switch f {
	case Ltxt:
		w.Int64(int64(m.ID))
		w.String(m.Name, 4, -1)
	case File:
		if m.Extended {
			w.Int32(m.ID)
			w.String(m.Name, 4, -1)
			w.Int16(m.Model)
			w.Int16(m.Unk3A)
			w.Int16(m.Level)
			w.Uint16(m.Exp)
			w.Byte(m.AI)
			w.Uint16(m.Gold1)
			w.Uint16(m.Gold2)
			w.Int16(m.Quest)
			w.Int32(m.HP)
			for _, v := range []uint16{m.SP, m.MP, m.Dex, m.Wis, m.Luc} {
				w.Uint16(v)
			}
			w.Put([]byte{m.Day, m.Size, m.Attrib})
			w.Uint16(m.Defense)
			w.Uint16(m.Magic)
			w.Put(m.Resist[:])
			w.Put(m.RK[:])
			w.Int32(m.NormalTime)
			w.Byte(m.NormalStep)
			w.Int32(m.ChaseTime)
			w.Byte(m.ChaseStep)
			w.Byte(m.ChaseRange)
			for k := range m.Attacks {
				m.Attacks[k].write(w, File)
			}
			w.Uint32(9)
			for i := range m.DropGroup {
				w.Uint16(m.DropGroup[i])
				w.Int32(m.DropRate[i])
			}
		} else {
			w.String(m.Name, 4, -1)
			w.Int16(m.Model)
			w.Int16(m.Level)
			w.Byte(m.AI)
			w.Int32(m.HP)
			w.Put([]byte{m.Day, m.Size, m.Attrib})
			w.Int32(m.NormalTime)
			w.Byte(m.NormalStep)
			w.Int32(m.ChaseTime)
			w.Byte(m.ChaseStep)
			w.Put([]byte{byte(m.Attacks[0].Type), m.Attacks[0].Mode, byte(m.Attacks[1].Type), m.Attacks[1].Range,
				byte(m.Attacks[2].Type), m.Attacks[2].Mode, byte(m.Attacks[2].Add)})
			w.Int16(m.Quest)
		}
		if m.Version&0xFFF >= 0x620 {
			w.Int16(m.ColorType)
			w.Put([]byte{m.ColorHue, m.ColorSat, m.ColorLight})
		}
	case Flnt:
		start := w.Len()
		m.writeFlnt(w)
		m.flnt.restore(w.Buf[start:])
	}
}

// Line exports the monster: "csv " (';'), "mobs" (SQL [Mobs]) or "drop"
// (SQL [MobItems]) (toText).
func (m *Monster) Line(format string) (string, bool) {
	var b strings.Builder
	switch format {
	case "drop":
		for i := range m.DropGroup {
			fmt.Fprintf(&b, "INSERT INTO [MobItems] values (%d, %d, %d, %d)\r\n", m.ID, i+1, m.DropGroup[i], m.DropRate[i])
		}
	case "mobs":
		fmt.Fprintf(&b, "INSERT INTO [Mobs] values (%d,%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,", m.ID,
			sqlQuote(m.Name), m.Level, int16(m.Exp), m.AI, int16(m.Gold1), int16(m.Gold2), m.Quest, m.HP,
			int16(m.SP), int16(m.MP), int16(m.Dex), int16(m.Wis), int16(m.Luc), m.Day, m.Size, m.Attrib,
			int16(m.Defense), int16(m.Magic))
		for _, v := range m.Resist {
			fmt.Fprintf(&b, "%d,", v)
		}
		for _, v := range m.RK {
			fmt.Fprintf(&b, "%d,", v)
		}
		fmt.Fprintf(&b, "%d,%d,%d,%d,%d,%s,%s,%s)\r\n", m.NormalTime, m.NormalStep, m.ChaseTime, m.ChaseStep,
			m.ChaseRange, m.Attacks[0].text(), m.Attacks[1].text(), m.Attacks[2].text())
	case TextCSV:
		const s = ";"
		fmt.Fprintf(&b, "%d;%s;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d;%d", m.ID, sqlQuote(m.Name),
			m.Model, m.Unk3A, m.Level, m.Exp, m.AI, m.Gold1, m.Gold2, m.Quest, m.HP, m.SP, m.MP, m.Dex, m.Wis,
			m.Luc, m.Day, m.Size, m.Attrib, m.Defense, m.Magic)
		for _, v := range m.Resist {
			fmt.Fprintf(&b, "%s%d", s, v)
		}
		for _, v := range m.RK {
			fmt.Fprintf(&b, "%s%d", s, v)
		}
		fmt.Fprintf(&b, ";%d;%d;%d;%d;%d;%s;%s;%s", m.NormalTime, m.NormalStep, m.ChaseTime, m.ChaseStep,
			m.ChaseRange, m.Attacks[0].text(), m.Attacks[1].text(), m.Attacks[2].text())
		for i := range m.DropGroup {
			fmt.Fprintf(&b, ";%d;%d", m.DropGroup[i], m.DropRate[i])
		}
		if m.Version&0xFFF >= 0x620 {
			fmt.Fprintf(&b, ";%d;%d;%d;%d", m.ColorType, m.ColorHue, m.ColorSat, m.ColorLight)
		}
	default:
		return "", false
	}
	return b.String(), true
}

// Monsters is Monster.SData, MonsterEx.SData, DBMonsterData and
// DBMonsterText (Monsters).
type Monsters struct {
	Monsters []*Monster
	Format   int // 0x500, 0x620, 0x1600
	Extended bool
	DataHead []byte
	TextHead []byte
}

// NewMonsters returns an empty table (format 0x500).
func NewMonsters() *Monsters { return &Monsters{Format: 0x500} }

// FilePrefix is the prefix of the DB text files.
func (s *Monsters) FilePrefix() string { return `BinarySData\DBMonsterText_` }

// SetVersion sets the format of the table and its rows (setCodePage).
func (s *Monsters) SetVersion(v int) {
	s.Format = v
	for _, m := range s.Monsters {
		m.Version = v
	}
}

// SetExtended switches between Monster.SData and MonsterEx.SData layouts.
func (s *Monsters) SetExtended(on bool) {
	s.Extended = on
	for _, m := range s.Monsters {
		m.Extended = on
	}
}

// Find returns the monster with an id (findById).
func (s *Monsters) Find(id int32) *Monster {
	for _, m := range s.Monsters {
		if m.ID == id {
			return m
		}
	}
	return nil
}

func monsterColumns() []string {
	c := []string{"id", "image", "level", "exp", "ai", "money1", "money2"}
	for j := 1; j <= 9; j++ {
		c = append(c, fmt.Sprintf("item%d", j), fmt.Sprintf("itemdroprate%d", j))
	}
	c = append(c, "questitem", "hp", "sp", "mp", "dex", "wis", "luc", "day", "size", "attrib", "defense", "magic")
	for k := 1; k <= 15; k++ {
		c = append(c, fmt.Sprintf("state%d", k))
	}
	for k := 1; k <= 6; k++ {
		c = append(c, fmt.Sprintf("skill%d", k))
	}
	c = append(c, "normaltime", "normalstep", "chasetime", "chasestep", "chaserange")
	for n := 1; n <= 3; n++ {
		for _, a := range []string{"attackani", "attacktype", "attacktime", "attackrange", "attack",
			"attackplus", "attackattrib", "attackspecial", "attackok"} {
			c = append(c, fmt.Sprintf("%s%d", a, n))
		}
	}
	return append(c, "colortype", "colorhue", "colorsaturation", "colorlight")
}

// Read loads the table (Monsters::read). The .SData format is tried as
// EP6.2 then EP5; a file starting with a zero count is read past a 64-byte
// header.
func (s *Monsters) Read(data []byte, cp int, f Format) error {
	switch f {
	case Ltxt:
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			s.TextHead = r.Copy(128)
			if r.Uint32() != 2 {
				failDB()
			}
			checkSignature(r, "id")
			checkSignature(r, "name")
			for n := r.Int(); n > 0; n-- {
				id := int32(r.Int64())
				if m := s.Find(id); m != nil {
					m.Read(r, Ltxt)
				} else {
					r.String(4, -1)
				}
			}
		})
	case Flnt:
		s.Monsters = nil
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			s.DataHead = r.Copy(128)
			if r.Uint32() != 94 {
				failDB()
			}
			s.Format = 0x1600
			for _, c := range monsterColumns() {
				checkSignature(r, c)
			}
			for n := int(r.Uint32()); n > 0; n-- {
				id := int32(r.Int64())
				m := NewMonster(id, 0x1620)
				s.Monsters = append(s.Monsters, m)
				m.Read(r, Flnt)
			}
		})
	case File:
		for _, v := range []int{0x620, 0x500} {
			s.Monsters = nil
			s.Format = v
			err := binstream.Try(func() {
				r := binstream.NewReader(data, cp)
				n := r.Int()
				if n == 0 && r.Len() > 0 {
					r.Pos = 0
					r.Skip(64)
					n = r.Int()
				}
				for i := 0; i < n; i++ {
					m := NewMonster(int32(i+1), v)
					s.Monsters = append(s.Monsters, m)
					m.Extended = s.Extended
					m.Read(r, File)
				}
			})
			if err == nil {
				return nil
			}
		}
		s.Monsters = nil
		s.Format = 0
		return ErrFormat
	}
	return ErrFormat
}

// Write serializes the table, sorted by id (Monsters::write).
func (s *Monsters) Write(cp int, f Format) ([]byte, error) {
	SortList(s.Monsters, "mbid", "", true)
	w := binstream.NewWriter(cp)
	switch f {
	case Ltxt:
		if s.TextHead == nil {
			return nil, fmt.Errorf("Error: textHeader does not exist")
		}
		w.Put(s.TextHead)
		w.Uint32(2)
		writeColumn(w, "id")
		writeColumn(w, "name")
	case Flnt:
		if s.DataHead == nil {
			return nil, fmt.Errorf("Error: dataHeader does not exist")
		}
		w.Put(s.DataHead)
		w.Uint32(94)
		for _, c := range monsterColumns() {
			writeColumn(w, c)
		}
	}
	w.Uint32(uint32(len(s.Monsters)))
	for _, m := range s.Monsters {
		m.Write(w, f)
	}
	return w.Bytes(), nil
}

func sqlTruncateHeader(b *strings.Builder, t1, t2 string) {
	b.WriteString("USE [PS_GameDefs]\r\nGO\r\nTRUNCATE TABLE [" + t1 + "]\r\n")
	if t2 != "" {
		b.WriteString("TRUNCATE TABLE [" + t2 + "]\r\n")
	}
	b.WriteString("GO\r\n")
}

// Text exports the table: "mobs", "drop", "mbdp" (both) or "csv ".
func (s *Monsters) Text(format string) string {
	var b strings.Builder
	each := func(f string) {
		for _, m := range s.Monsters {
			line, _ := m.Line(f)
			b.WriteString(line)
		}
	}
	switch format {
	case "mobs":
		sqlTruncateHeader(&b, "Mobs", "")
		each("mobs")
	case "drop":
		sqlTruncateHeader(&b, "MobItems", "")
		each("drop")
	case "mbdp":
		sqlTruncateHeader(&b, "Mobs", "MobItems")
		each("mobs")
		b.WriteString("\r\n")
		each("drop")
	default:
		each(format)
	}
	return b.String()
}
