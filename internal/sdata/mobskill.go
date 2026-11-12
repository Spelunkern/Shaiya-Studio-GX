package sdata

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// MobsSkill is an NPC skill: NpcSkill.SData / DBNpcSkillData (MobsSkill).
type MobsSkill struct {
	Name  string
	Desc  string
	Skill SkillData
}

// NewMobsSkill returns a cleared skill.
func NewMobsSkill(version int, id int32) *MobsSkill {
	s := &MobsSkill{Skill: SkillData{Version: version, SkillId: id}}
	return s
}

// Reset clears the skill but keeps its id and version.
func (s *MobsSkill) Reset() {
	s.Skill = SkillData{Version: s.Skill.Version, SkillId: s.Skill.SkillId}
	s.Name, s.Desc = "", ""
}

// IconID returns the icon (getId).
func (s *MobsSkill) IconID() int { return int(s.Skill.Icon) }

// Assign copies src (operator=).
func (s *MobsSkill) Assign(src *MobsSkill) { *s = *src }

// Text returns "name", "desc" or "tpdt" ("<type> - <type name>").
func (s *MobsSkill) Text(tag Tag) string {
	switch tag {
	case "name":
		return s.Name
	case "desc":
		return s.Desc
	case "tpdt":
		return fmt.Sprintf("%d - %s", s.Skill.TypeDetail, SkillTypeName(s.Skill.TypeDetail))
	}
	return ""
}

// SetText sets the name or the description.
func (s *MobsSkill) SetText(tag Tag, v string) bool {
	switch tag {
	case "name":
		s.Name = v
	case "desc":
		s.Desc = v
	default:
		return false
	}
	return true
}

// Get / Set access the skill columns.
func (s *MobsSkill) Get(tag Tag) int         { return s.Skill.Get(tag) }
func (s *MobsSkill) Set(tag Tag, v int) bool { return s.Skill.Set(tag, v) }

// Compare compares on tag.
func (s *MobsSkill) Compare(o *MobsSkill, tag Tag) int {
	switch tag {
	case "name":
		return compareLogical(s.Name, o.Name)
	case "desc":
		return compareLogical(s.Desc, o.Desc)
	}
	return s.Skill.Compare(&o.Skill, tag)
}

func (s *MobsSkill) read(r *binstream.Reader, f Format) {
	if f == Ltxt {
		return
	}
	if f == File {
		s.Name = r.String(4, r.CodePage)
		s.Desc = r.String(4, r.CodePage)
	}
	s.Skill.Read(r, f)
}

func (s *MobsSkill) write(w *binstream.Writer, f Format) {
	if f == Ltxt {
		return
	}
	if f == File {
		w.String(s.Name, 4, w.CodePage)
		w.String(s.Desc, 4, w.CodePage)
	}
	s.Skill.Write(w, f)
}

// skillColumns are the 100 columns of DBSkillData and DBNpcSkillData.
func skillColumns() []string {
	c := []string{"id", "skilllevel", "image", "ani", "effect", "toggletype", "sound", "level", "country",
		"attackfighter", "defensefighter", "patrolrogue", "shootrogue", "attackmage", "defensemage", "grow",
		"point", "typeshow", "typeattack", "typeeffect", "type"}
	for i := 1; i <= 15; i++ {
		c = append(c, fmt.Sprintf("needweapon%d", i))
	}
	c = append(c, "shield", "sp", "mp", "readytime", "resettime", "attackrange", "statetype", "attribtype",
		"disable", "prevskill", "successtype", "successvalue", "targettype", "applyrange", "multiattack",
		"keeptime", "weapon1", "weapon2", "weaponvalue", "bag", "arrow", "damagetype", "damage1", "damage2",
		"damage3", "timedamagetype", "timedamage1", "timedamage2", "timedamage3", "adddamage1", "adddamage2",
		"adddamage3")
	for i := 1; i <= 10; i++ {
		c = append(c, fmt.Sprintf("abilitytype%d", i), fmt.Sprintf("abilityvalue%d", i))
	}
	return append(c, "heal1", "heal2", "heal3", "timeheal1", "timeheal2", "timeheal3", "defencetype",
		"defencevalue", "limithp", "fixrange", "changetype", "changelevel")
}

// Line exports the skill (toText): "csv ", "sql " or a quoted dump.
func (s *MobsSkill) Line(format string) string {
	sep := exportSeparator(format)
	switch format {
	case TextCSV:
		desc := strings.ReplaceAll(escape(s.Desc), ";", ",")
		return fmt.Sprintf("%d%s100%s%s%s%s%s%s\r\n", s.Skill.SkillId, sep, sep, s.Name, sep, desc, sep, s.Skill.Line(TextCSV))
	case TextSQL:
		cols := "SkillID,SkillLevel,SkillName,Country,Attackfighter,Defensefighter,Patrolrogue,Shootrogue,Attackmage,Defensemage,PrevSkillID,ReqLevel,Grow,SkillPoint,TypeShow,TypeAttack,TypeEffect,TypeDetail,NeedWeapon1,NeedWeapon2,NeedWeapon3,NeedWeapon4,NeedWeapon5,NeedWeapon6,NeedWeapon7,NeedWeapon8,NeedWeapon9,NeedWeapon10,NeedWeapon11,NeedWeapon12,NeedWeapon13,NeedWeapon14,NeedWeapon15,Shield,SP,MP,ReadyTime,ResetTime,AttackRange,StateType,AttrType,[Disable],SuccessType,SuccessValue,TargetType,ApplyRange,MultiAttack,KeepTime,Weapon1,Weapon2,Weaponvalue,Bag,Arrow,DamageType,DamageHP,DamageSP,DamageMP,TimeDamageType,TimeDamageHP,TimeDamageSP,TimeDamageMP,AddDamageHP,AddDamageSP,AddDamageMP,AbilityType1,AbilityValue1,AbilityType2,AbilityValue2,AbilityType3,AbilityValue3,"
		if s.Skill.Version >= 0x600 {
			cols += "AbilityType4,AbilityValue4,AbilityType5,AbilityValue5,AbilityType6,AbilityValue6,AbilityType7,AbilityValue7,AbilityType8,AbilityValue8,AbilityType9,AbilityValue9,AbilityType10,AbilityValue10,"
		}
		cols += "HealHP,HealSP,HealMP,TimeHealHP,TimeHealSP,TimeHealMP,DefenceType,DefenceValue,LimitHP,FixRange,ChangeType,ChangeLevel,UpdateDate"
		return "INSERT INTO [PS_GameDefs].[dbo].[Skills] (" + cols + ") VALUES \r\n(" +
			fmt.Sprintf("%d%s100%s%s%s", s.Skill.SkillId, sep, sep, sqlQuote(s.Name), sep) + s.Skill.Line(TextSQL) + ");\r\n"
	}
	return `"` + s.Name + `"` + sep + `"` + s.Desc + `"` + sep + s.Skill.Line(format) + ",\r\n"
}

// MobsSkills is NpcSkill.SData / DBNpcSkillData / DBNpcSkillText
// (MobsSkills).
type MobsSkills struct {
	Skills   []*MobsSkill
	Format   int // 0x400, 0x500, 0x600 or 0x1600
	DataHead []byte
	TextHead []byte
}

// NewMobsSkills returns an empty table (format 0x500).
func NewMobsSkills() *MobsSkills { return &MobsSkills{Format: 0x500} }

// FilePrefix is the prefix of the DB text files.
func (t *MobsSkills) FilePrefix() string { return `BinarySData\DBNpcSkillText_` }

// SetVersion sets the format of the table and its rows.
func (t *MobsSkills) SetVersion(v int) {
	t.Format = v
	for _, s := range t.Skills {
		s.Skill.Version = v
	}
}

// Visible lists the skills with a type (MobsSkillSet::sub_44D4A0).
func (t *MobsSkills) Visible() []*MobsSkill {
	var out []*MobsSkill
	for _, s := range t.Skills {
		if s.Skill.TypeDetail != 0 {
			out = append(out, s)
		}
	}
	return out
}

func (t *MobsSkills) levels() int {
	switch {
	case t.Format == 0x400:
		return 3
	case t.Format >= 0x600:
		return 15
	}
	return 9
}

// Read loads the table (MobsSkills::read). The .SData format holds 3, 9 or
// 15 levels per skill; like shStudio, only the first one is kept.
func (t *MobsSkills) Read(data []byte, cp int, f Format) error {
	switch f {
	case Ltxt:
		return binstream.Try(func() {
			r := binstream.NewReader(data, cp)
			t.TextHead = r.Copy(128)
			if r.Int() != 4 {
				failDB()
			}
			for _, c := range []string{"id", "skilllevel", "name", "text"} {
				checkSignature(r, c)
			}
			for n := r.Int(); n > 0; n-- {
				id := int32(r.Int64())
				r.Int64()
				var found *MobsSkill
				for _, s := range t.Skills {
					if s.Skill.SkillId == id {
						found = s
						break
					}
				}
				name, desc := r.String(4, -1), r.String(4, -1)
				if found != nil {
					found.Name, found.Desc = name, desc
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
			for n := int(r.Uint32()); n > 0; n-- {
				id := int32(r.Int64())
				r.Int64()
				r.Pos -= 16
				s := NewMobsSkill(0x1620, id)
				t.Skills = append(t.Skills, s)
				s.read(r, Flnt)
			}
		})
	case File:
		for _, v := range []int{0x600, 0x500, 0x400} {
			t.Skills = nil
			t.Format = v
			err := binstream.Try(func() {
				r := binstream.NewReader(data, cp)
				extra := NewMobsSkill(v, 0)
				levels := t.levels()
				n := r.Int()
				for i := 0; i < n; i++ {
					s := NewMobsSkill(v, int32(int16(i+1)))
					s.read(r, File)
					t.Skills = append(t.Skills, s)
					for j := 1; j < levels; j++ {
						extra.read(r, File)
					}
				}
				if r.Len() >= 32 {
					binstream.Fail("trailing data")
				}
			})
			if err == nil {
				return nil
			}
		}
		t.Skills = nil
		t.Format = 0
		return ErrFormat
	}
	return ErrFormat
}

// Write serializes the table (MobsSkills::write); an empty table writes
// nothing.
func (t *MobsSkills) Write(cp int, f Format) ([]byte, error) {
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
		w.Uint32(uint32(len(t.Skills)))
		for _, s := range t.Skills {
			w.Int64(int64(s.Skill.SkillId))
			w.Int64(1)
			w.String(s.Name, 4, -1)
			w.String(s.Desc, 4, -1)
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
		w.Uint32(uint32(len(t.Skills)))
		for _, s := range t.Skills {
			s.Skill.Write(w, Flnt)
		}
	case File:
		extra := NewMobsSkill(t.Format, 0)
		w.Uint32(uint32(len(t.Skills)))
		for _, s := range t.Skills {
			s.write(w, File)
			for k := t.levels() - 1; k > 0; k-- {
				extra.write(w, File)
			}
		}
	}
	return w.Bytes(), nil
}

// Text exports the table (toText).
func (t *MobsSkills) Text(format string) string {
	var b strings.Builder
	switch format {
	case "----", TextCSV:
		b.WriteString("SkillID;SkillLevel;SkillName;Description;Icon;Ani;Country;Fighter;Defenser;Ranger;Archer;Mage;Healer;PrevSkill;ReqLevel;Grow;Point;TypeShow;TypeAttack;TypeEffect;TypeDetail;NeedWeapon1;NeedWeapon2;NeedWeapon3;NeedWeapon4;NeedWeapon5;NeedWeapon6;NeedWeapon7;NeedWeapon8;NeedWeapon9;NeedWeapon10;NeedWeapon11;NeedWeapon12;NeedWeapon13;NeedWeapon14;NeedWeapon15;Shield;SP;MP;ReadyTime;ResetTime;AttackRange;StateType;AttribType;Disable;SuccessType;SuccessValue;TargetType;ApplyRange;MultiAttack;KeepTime;Weapon1;Weapon2;WeaponValue;Bag;Arrow;DamageType;DamageHP;DamageSP;DamageMP;TimeDamageType;TimeDamageHP;TimeDamageSP;TimeDamageMP;AddDamageHP;AddDamageSP;AddDamageMP;AbilityType1;AbilityValue1;AbilityType2;AbilityValue2;AbilityType3;AbilityValue3;")
		if t.Format >= 0x600 {
			b.WriteString("AbilityType4;AbilityValue4;AbilityType5;AbilityValue5;AbilityType6;AbilityValue6;AbilityType7;AbilityValue7;AbilityType8;AbilityValue8;AbilityType9;AbilityValue9;AbilityType10;AbilityValue10;")
		}
		b.WriteString("HealHP;HealSP;HealMP;TimeHealHP;TimeHealSP;TimeHealMP;DefenceType;DefenceValue;LimitHP;FixRange;ChangeType;ChangeLevel\r\n")
	case TextSQL:
		b.WriteString("DELETE FROM [PS_GameDefs].[dbo].[Skills] WHERE SkillLevel=100;\r\nGO\r\n")
	}
	for _, s := range t.Skills {
		b.WriteString(s.Line(format))
	}
	return b.String()
}
