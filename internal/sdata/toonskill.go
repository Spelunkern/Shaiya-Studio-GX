package sdata

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// MaxSkillLevels is the number of levels a player skill can have.
const MaxSkillLevels = 15

// ToonSkill is a player skill: Skill.SData / DBSkillData (ToonSkill). It
// has 3 (EP4), 9 (EP5) or 15 (EP6+) levels.
type ToonSkill struct {
	Format     int
	LevelCount int
	SkillID    int32
	Name       string
	LevelNames [MaxSkillLevels]string
	LevelTexts [MaxSkillLevels]string
	Levels     [MaxSkillLevels]SkillData
}

// NewToonSkill returns an empty skill for format.
func NewToonSkill(id int32, format int) *ToonSkill {
	s := &ToonSkill{Format: format & 0xFFF, SkillID: id, LevelCount: 9}
	switch {
	case format&0xFFF == 0x400:
		s.LevelCount = 3
	case format&0xFFF >= 0x600:
		s.LevelCount = 15
	}
	for i := range s.Levels {
		s.Levels[i].Version = format
	}
	return s
}

// IconID returns the icon of the first level (getId).
func (s *ToonSkill) IconID() int { return int(s.Levels[0].Icon) }

// Text returns "name", "desc" (level 1 text) or "tpdt".
func (s *ToonSkill) Text(tag Tag) string {
	switch tag {
	case "name":
		return s.Name
	case "desc":
		return s.LevelTexts[0]
	case "tpdt":
		t := s.Levels[0].TypeDetail
		return fmt.Sprintf("%d - %s", t, SkillTypeName(t))
	}
	return ""
}

// SetText sets the name.
func (s *ToonSkill) SetText(tag Tag, v string) bool {
	if tag != "name" {
		return false
	}
	s.Name = v
	return true
}

// Get returns "skid" or a column of level 1.
func (s *ToonSkill) Get(tag Tag) int {
	if tag == "skid" {
		return int(s.SkillID)
	}
	return s.Levels[0].Get(tag)
}

// Set sets "skid" or a column of level 1.
func (s *ToonSkill) Set(tag Tag, v int) bool {
	if tag == "skid" {
		s.SkillID = int32(v)
		return true
	}
	return s.Levels[0].Set(tag, v)
}

// Reset clears every level (keeping version and id of each).
func (s *ToonSkill) Reset() {
	s.Name = ""
	for i := range s.Levels {
		l := &s.Levels[i]
		*l = SkillData{Version: l.Version, SkillId: l.SkillId}
		s.LevelNames[i], s.LevelTexts[i] = "", ""
	}
}

// Compare compares on tag.
func (s *ToonSkill) Compare(o *ToonSkill, tag Tag) int {
	switch tag {
	case "name":
		return compareLogical(s.Name, o.Name)
	case "desc":
		return compareLogical(s.LevelTexts[0], o.LevelTexts[0])
	case "skid":
		return cmp(s.SkillID, o.SkillID)
	}
	return s.Levels[0].Compare(&o.Levels[0], tag)
}

// SetFormat changes the format (setFormat); unknown formats are ignored.
func (s *ToonSkill) SetFormat(format int) {
	switch format & 0xFFF {
	case 0x400:
		s.LevelCount = 3
	case 0x500:
		s.LevelCount = 9
	case 0x600:
		s.LevelCount = 15
	default:
		return
	}
	s.Format = format
	for i := range s.Levels {
		s.Levels[i].Version = format
	}
}

// Assign copies the levels, names and texts of other (assign).
func (s *ToonSkill) Assign(o *ToonSkill) bool {
	s.LevelCount = o.LevelCount
	s.Name = o.Name
	s.LevelNames = o.LevelNames
	s.LevelTexts = o.LevelTexts
	s.Levels = o.Levels
	for i := range s.Levels {
		s.Levels[i].flnt = nil
	}
	return true
}

func (s *ToonSkill) read(r *binstream.Reader) {
	for i := 0; i < s.LevelCount; i++ {
		s.LevelNames[i] = r.String(4, r.CodePage)
		s.LevelTexts[i] = r.String(4, r.CodePage)
		s.Levels[i].Read(r, File)
	}
	s.Name = SkillBaseName(s.LevelNames[0])
}

func (s *ToonSkill) write(w *binstream.Writer, f Format) {
	for i := 0; i < s.LevelCount; i++ {
		switch f {
		case Ltxt:
			w.String(s.LevelNames[i], 4, w.CodePage)
			w.String(s.LevelTexts[i], 4, w.CodePage)
		case Flnt:
			s.Levels[i].Write(w, Flnt)
		case File:
			w.String(s.LevelNames[i], 4, w.CodePage)
			w.String(s.LevelTexts[i], 4, w.CodePage)
			s.Levels[i].Write(w, File)
		}
	}
}

// Line exports the skill (toText): one line per level.
func (s *ToonSkill) Line(format string) string {
	sep := exportSeparator(format)
	var b strings.Builder
	name := func(i int) string {
		if s.LevelNames[i] != "" {
			return s.LevelNames[i]
		}
		return s.Name
	}
	switch format {
	case TextCSV:
		for i := 0; i < s.LevelCount; i++ {
			text := strings.ReplaceAll(escape(s.LevelTexts[i]), ";", ",")
			fmt.Fprintf(&b, "%d%s%d%s%s%s%s%s%s\r\n", s.SkillID, sep, s.Levels[i].Level, sep, name(i), sep, text, sep, s.Levels[i].Line(TextCSV))
		}
	case TextSQL:
		cols := "SkillID,SkillLevel,SkillName,Country,Attackfighter,Defensefighter,Patrolrogue,Shootrogue,Attackmage,Defensemage,PrevSkillID,ReqLevel,Grow,SkillPoint,TypeShow,TypeAttack,TypeEffect,TypeDetail,NeedWeapon1,NeedWeapon2,NeedWeapon3,NeedWeapon4,NeedWeapon5,NeedWeapon6,NeedWeapon7,NeedWeapon8,NeedWeapon9,NeedWeapon10,NeedWeapon11,NeedWeapon12,NeedWeapon13,NeedWeapon14,NeedWeapon15,Shield,SP,MP,ReadyTime,ResetTime,AttackRange,StateType,AttrType,[Disable],SuccessType,SuccessValue,TargetType,ApplyRange,MultiAttack,KeepTime,Weapon1,Weapon2,Weaponvalue,Bag,Arrow,DamageType,DamageHP,DamageSP,DamageMP,TimeDamageType,TimeDamageHP,TimeDamageSP,TimeDamageMP,AddDamageHP,AddDamageSP,AddDamageMP,AbilityType1,AbilityValue1,AbilityType2,AbilityValue2,AbilityType3,AbilityValue3,"
		if s.Format >= 0x600 {
			cols += "AbilityType4,AbilityValue4,AbilityType5,AbilityValue5,AbilityType6,AbilityValue6,AbilityType7,AbilityValue7,AbilityType8,AbilityValue8,AbilityType9,AbilityValue9,AbilityType10,AbilityValue10,"
		}
		cols += "HealHP,HealSP,HealMP,TimeHealHP,TimeHealSP,TimeHealMP,DefenceType,DefenceValue,LimitHP,FixRange,ChangeType,ChangeLevel,UpdateDate"
		b.WriteString("INSERT INTO [PS_GameDefs].[dbo].[Skills] (" + cols + ") VALUES \r\n")
		for i := 0; i < s.LevelCount; i++ {
			end := ","
			if i == s.LevelCount-1 {
				end = ";"
			}
			fmt.Fprintf(&b, "(%d%s%d%s%s%s%s)%s\r\n", s.SkillID, sep, s.Levels[i].Level, sep, sqlQuote(name(i)), sep, s.Levels[i].Line(TextSQL), end)
		}
	default:
		for i := 0; i < s.LevelCount; i++ {
			b.WriteString(s.Levels[i].Line(format))
		}
	}
	return b.String()
}

// SkillBaseName removes the level suffix of a level name ("Slash Lv 3" ->
// "Slash") (ToonSkill::baseName).
func SkillBaseName(src string) string {
	d := []rune(src)
	drop := func(n int) {
		if n > len(d) {
			n = len(d)
		}
		d = d[:len(d)-n]
	}
	hasSuffix := func(suf string) bool {
		return len(d) >= len(suf) && strings.EqualFold(string(d[len(d)-len(suf):]), suf)
	}
	if hasSuffix("...") {
		drop(3)
	}
	length := len(d)
	if length <= 0 {
		return string(d)
	}
	if c := d[length-1]; c >= '1' && c <= '9' {
		drop(1)
		if prev := length - 2; prev >= 0 && prev < len(d) && d[prev] == '1' {
			drop(1)
		}
	}
	if len(d) > 0 && d[len(d)-1] == ' ' {
		drop(1)
	}
	if hasSuffix("lvl") {
		drop(3)
	} else if hasSuffix("lv") {
		drop(2)
	}
	if len(d) > 0 && d[len(d)-1] == ' ' {
		drop(1)
	}
	return string(d)
}

// ToonSkills is Skill.SData / DBSkillData / DBSkillText (ToonSkills).
type ToonSkills struct {
	Skills   []*ToonSkill
	Format   int
	DataHead []byte
	TextHead []byte
}

// NewToonSkills returns an empty table (format 0x500).
func NewToonSkills() *ToonSkills { return &ToonSkills{Format: 0x500} }

// FilePrefix is the prefix of the DB text files.
func (t *ToonSkills) FilePrefix() string { return `BinarySData\DBSkillText_` }

// SetVersion sets the format of the table and its skills (setCodePage).
func (t *ToonSkills) SetVersion(v int) {
	t.Format = v
	for _, s := range t.Skills {
		s.SetFormat(v)
	}
}

// Find returns the skill with an id (findById).
func (t *ToonSkills) Find(id int32) *ToonSkill {
	for _, s := range t.Skills {
		if s.SkillID == id {
			return s
		}
	}
	return nil
}

// Filter lists skills for the skill views (ToonSkillSet::fill): mode 0 all,
// 1 shown skills usable by classIndex (0 = country), 2 hidden skills; modes
// 1 and 2 skip skills without a type.
func (t *ToonSkills) Filter(mode, classIndex int) []*ToonSkill {
	var out []*ToonSkill
	for _, s := range t.Skills {
		if mode != 0 {
			l := &s.Levels[0]
			if l.TypeDetail == 0 {
				continue
			}
			if mode == 1 {
				cls := l.Country
				if classIndex > 0 {
					cls = l.Classes[classIndex-1]
				}
				if l.TypeShow == 0 || classIndex != 0 && cls == 0 {
					continue
				}
			} else if mode != 2 || l.TypeShow != 0 {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

// Read loads the table (ToonSkills::read).
func (t *ToonSkills) Read(data []byte, cp int, f Format) error {
	switch f {
	case Ltxt:
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			t.TextHead = r.Copy(128)
			if r.Uint32() != 4 {
				failDB()
			}
			for _, c := range []string{"id", "skilllevel", "name", "text"} {
				checkSignature(r, c)
			}
			for n := int(r.Uint32()); n > 0; n-- {
				id := int32(r.Int64())
				level := int(r.Int64())
				s := t.Find(id)
				if s != nil && level >= 1 && level <= 15 {
					s.LevelNames[level-1] = r.String(4, r.CodePage)
					s.LevelTexts[level-1] = r.String(4, r.CodePage)
					if level == 1 {
						s.Name = SkillBaseName(s.LevelNames[0])
					}
				} else {
					r.String(4, -1)
					r.String(4, -1)
				}
			}
		})
	case Flnt:
		t.Skills = nil
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			t.DataHead = r.Copy(128)
			if r.Uint32() != 100 {
				failDB()
			}
			t.Format = 0x1600
			for _, c := range skillColumns() {
				checkSignature(r, c)
			}
			rows := int(r.Uint32())
			for i := 0; i < rows; i++ {
				id := int32(r.Int64())
				level := int(r.Int64())
				if level < 1 || level > 15 {
					binstream.Fail("Invalid skill level value (out of range).")
				}
				r.Pos -= 16
				s := t.Find(id)
				if s == nil {
					s = NewToonSkill(id, t.Format)
					t.Skills = append(t.Skills, s)
				}
				s.Levels[level-1].Read(r, Flnt)
			}
		})
	case File:
		for _, v := range []int{0x600, 0x500, 0x400} {
			t.Skills = nil
			t.Format = v
			err := binstream.Try(func() {
				r := binstream.NewReader(data, cp)
				n := int(r.Uint32())
				for id := 1; id <= n; id++ {
					s := NewToonSkill(int32(id), v)
					t.Skills = append(t.Skills, s)
					s.read(r)
				}
			})
			if err == nil {
				return nil
			}
		}
		t.Skills = nil
		t.Format = 0
		return errors.New("Unsupported skills format")
	}
	return ErrFormat
}

// Write serializes the table (ToonSkills::write); an empty table writes
// nothing.
func (t *ToonSkills) Write(cp int, f Format) ([]byte, error) {
	w := binstream.NewWriter(cp)
	if len(t.Skills) < 1 {
		return w.Bytes(), nil
	}
	switch f {
	case Ltxt:
		if t.TextHead == nil {
			return nil, errors.New("error")
		}
		w.Put(t.TextHead)
		w.Uint32(4)
		for _, c := range []string{"id", "skilllevel", "name", "text"} {
			writeColumn(w, c)
		}
		w.Uint32(uint32(15 * len(t.Skills)))
		for _, s := range t.Skills {
			for level := 1; level <= 15; level++ {
				w.Int64(int64(s.SkillID))
				w.Int64(int64(level))
				w.String(s.LevelNames[level-1], 4, -1)
				w.String(s.LevelTexts[level-1], 4, -1)
			}
		}
	case Flnt:
		if t.DataHead == nil {
			return nil, errors.New("error")
		}
		w.Put(t.DataHead)
		w.Uint32(100)
		for _, c := range skillColumns() {
			writeColumn(w, c)
		}
		w.Uint32(uint32(15 * len(t.Skills)))
		for _, s := range t.Skills {
			s.write(w, Flnt)
		}
	case File:
		w.Uint32(uint32(len(t.Skills)))
		for _, s := range t.Skills {
			s.write(w, File)
		}
	}
	return w.Bytes(), nil
}

// Text exports the table (toText).
func (t *ToonSkills) Text(format string) string {
	var b strings.Builder
	switch format {
	case "----", TextCSV:
		b.WriteString("SkillID;SkillLevel;SkillName;Description;Icon;Ani;Country;Fighter;Defenser;Ranger;Archer;Mage;Healer;PrevSkill;ReqLevel;Grow;Point;TypeShow;TypeAttack;TypeEffect;TypeDetail;NeedWeapon1;NeedWeapon2;NeedWeapon3;NeedWeapon4;NeedWeapon5;NeedWeapon6;NeedWeapon7;NeedWeapon8;NeedWeapon9;NeedWeapon10;NeedWeapon11;NeedWeapon12;NeedWeapon13;NeedWeapon14;NeedWeapon15;Shield;SP;MP;ReadyTime;ResetTime;AttackRange;StateType;AttribType;Disable;SuccessType;SuccessValue;TargetType;ApplyRange;MultiAttack;KeepTime;Weapon1;Weapon2;WeaponValue;Bag;Arrow;DamageType;DamageHP;DamageSP;DamageMP;TimeDamageType;TimeDamageHP;TimeDamageSP;TimeDamageMP;AddDamageHP;AddDamageSP;AddDamageMP;AbilityType1;AbilityValue1;AbilityType2;AbilityValue2;AbilityType3;AbilityValue3;")
		if t.Format >= 0x600 {
			b.WriteString("AbilityType4;AbilityValue4;AbilityType5;AbilityValue5;AbilityType6;AbilityValue6;AbilityType7;AbilityValue7;AbilityType8;AbilityValue8;AbilityType9;AbilityValue9;AbilityType10;AbilityValue10;")
		}
		b.WriteString("HealHP;HealSP;HealMP;TimeHealHP;TimeHealSP;TimeHealMP;DefenceType;DefenceValue;LimitHP;FixRange;ChangeType;ChangeLevel\r\n")
		for _, s := range t.Skills {
			b.WriteString(s.Line(format))
		}
	case TextSQL:
		b.WriteString("DELETE FROM [PS_GameDefs].[dbo].[Skills] WHERE SkillLevel<100;\r\nGO\r\n")
		for _, s := range t.Skills {
			b.WriteString(s.Line(TextSQL))
		}
	}
	return b.String()
}
