package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// SkillAbility is one ability bonus of a skill level.
type SkillAbility struct {
	Type  uint8
	Value int16
}

// SkillData is the data of one skill level, shared by player and NPC skills
// (SkillData).
type SkillData struct {
	Version        int
	SkillId        int32
	Level          uint8
	Icon           int16
	Anim           uint8
	Effect         int16
	ToggleType     int16
	Sound          uint8
	ReqLevel       int16
	Country        uint8
	Classes        [6]uint8
	Grow           uint8
	Points         uint8
	TypeShow       uint8
	TypeAttack     uint8
	TypeEffect     uint8
	TypeDetail     int16
	NeedWeapon     [15]uint8
	Shield         uint8
	SP, MP         int16
	ReadyTime      uint8
	ResetTime      int16
	AttackRange    uint8
	StateType      uint8
	AttribType     uint8
	Disable        int16
	PrevSkill      int16
	SuccessType    uint8
	SuccessValue   uint8
	TargetType     uint8
	ApplyRange     uint8
	MultiAttack    uint8
	KeepTime       uint16
	Weapon         [2]uint8
	WeaponValue    uint8
	Bag            uint8
	Arrow          int16
	DamageType     uint8
	Damage         [3]int16
	TimeDamageType uint8
	TimeDamage     [3]int16
	AddDamage      [3]int16
	Ability        [10]SkillAbility
	Heal           [3]uint16
	TimeHeal       [3]uint16
	DefenceType    uint8
	DefenceValue   uint8
	LimitHP        uint8
	FixRange       uint8
	ChangeType     int16
	ChangeLevel    int16

	flnt *flntRow
}

// NewSkillData returns a cleared level (version 0x500).
func NewSkillData(id int32) SkillData { return SkillData{Version: 0x500, SkillId: id} }

func (d *SkillData) ptr(tag Tag) any {
	switch tag {
	case "skid":
		return &d.SkillId
	case "lvel":
		return &d.Level
	case "icon":
		return &d.Icon
	case "anim":
		return &d.Anim
	case "efct":
		return &d.Effect
	case "totp":
		return &d.ToggleType
	case "sond":
		return &d.Sound
	case "rlvl":
		return &d.ReqLevel
	case "cntr":
		return &d.Country
	case "figh":
		return &d.Classes[0]
	case "defs":
		return &d.Classes[1]
	case "rngr":
		return &d.Classes[2]
	case "arch":
		return &d.Classes[3]
	case "mage":
		return &d.Classes[4]
	case "prst":
		return &d.Classes[5]
	case "grow":
		return &d.Grow
	case "pnts":
		return &d.Points
	case "tpsh":
		return &d.TypeShow
	case "tpat":
		return &d.TypeAttack
	case "tpef":
		return &d.TypeEffect
	case "tpdt":
		return &d.TypeDetail
	case "nwp1":
		return &d.NeedWeapon[0]
	case "nwp2":
		return &d.NeedWeapon[1]
	case "nwp3":
		return &d.NeedWeapon[2]
	case "nwp4":
		return &d.NeedWeapon[3]
	case "nwp5":
		return &d.NeedWeapon[4]
	case "nwp6":
		return &d.NeedWeapon[5]
	case "nwp7":
		return &d.NeedWeapon[6]
	case "nwp8":
		return &d.NeedWeapon[7]
	case "nwp9":
		return &d.NeedWeapon[8]
	case "nwpA":
		return &d.NeedWeapon[9]
	case "nwpB":
		return &d.NeedWeapon[10]
	case "nwpC":
		return &d.NeedWeapon[11]
	case "nwpD":
		return &d.NeedWeapon[12]
	case "nwpE":
		return &d.NeedWeapon[13]
	case "nwpF":
		return &d.NeedWeapon[14]
	case "shld":
		return &d.Shield
	case "cssp":
		return &d.SP
	case "csmp":
		return &d.MP
	case "rdtm":
		return &d.ReadyTime
	case "rstm":
		return &d.ResetTime
	case "atrg":
		return &d.AttackRange
	case "sttp":
		return &d.StateType
	case "attp":
		return &d.AttribType
	case "disa":
		return &d.Disable
	case "prvs":
		return &d.PrevSkill
	case "sctp":
		return &d.SuccessType
	case "scvl":
		return &d.SuccessValue
	case "trtp":
		return &d.TargetType
	case "aprg":
		return &d.ApplyRange
	case "mult":
		return &d.MultiAttack
	case "kptm":
		return &d.KeepTime
	case "wpn1":
		return &d.Weapon[0]
	case "wpn2":
		return &d.Weapon[1]
	case "wval":
		return &d.WeaponValue
	case "bags":
		return &d.Bag
	case "arrw":
		return &d.Arrow
	case "dmtp":
		return &d.DamageType
	case "dmhp":
		return &d.Damage[0]
	case "dmsp":
		return &d.Damage[1]
	case "dmmp":
		return &d.Damage[2]
	case "tdtp":
		return &d.TimeDamageType
	case "tdhp":
		return &d.TimeDamage[0]
	case "tdsp":
		return &d.TimeDamage[1]
	case "tdmp":
		return &d.TimeDamage[2]
	case "adhp":
		return &d.AddDamage[0]
	case "adsp":
		return &d.AddDamage[1]
	case "admp":
		return &d.AddDamage[2]
	case "abt1":
		return &d.Ability[0].Type
	case "abt2":
		return &d.Ability[1].Type
	case "abt3":
		return &d.Ability[2].Type
	case "abt4":
		return &d.Ability[3].Type
	case "abt5":
		return &d.Ability[4].Type
	case "abt6":
		return &d.Ability[5].Type
	case "abt7":
		return &d.Ability[6].Type
	case "abt8":
		return &d.Ability[7].Type
	case "abt9":
		return &d.Ability[8].Type
	case "abt:":
		return &d.Ability[9].Type
	case "abv1":
		return &d.Ability[0].Value
	case "abv2":
		return &d.Ability[1].Value
	case "abv3":
		return &d.Ability[2].Value
	case "abv4":
		return &d.Ability[3].Value
	case "abv5":
		return &d.Ability[4].Value
	case "abv6":
		return &d.Ability[5].Value
	case "abv7":
		return &d.Ability[6].Value
	case "abv8":
		return &d.Ability[7].Value
	case "abv9":
		return &d.Ability[8].Value
	case "abv:":
		return &d.Ability[9].Value
	case "hlhp":
		return &d.Heal[0]
	case "hlsp":
		return &d.Heal[1]
	case "hlmp":
		return &d.Heal[2]
	case "thhp":
		return &d.TimeHeal[0]
	case "thsp":
		return &d.TimeHeal[1]
	case "thmp":
		return &d.TimeHeal[2]
	case "dftp":
		return &d.DefenceType
	case "dfvl":
		return &d.DefenceValue
	case "lmhp":
		return &d.LimitHP
	case "fxrg":
		return &d.FixRange
	case "chtp":
		return &d.ChangeType
	case "chlv":
		return &d.ChangeLevel
	}
	return nil
}

// skillNoCompare lists the tags that SkillData::compare does not handle.
var skillNoCompare = map[Tag]bool{"efct": true, "totp": true, "sond": true}

// Get returns a column value (getValue).
func (d *SkillData) Get(tag Tag) int { return getInt(d.ptr(tag)) }

// Set sets a column value (setValue).
func (d *SkillData) Set(tag Tag, v int) bool { return setInt(d.ptr(tag), v) }

// Compare compares a column (compare); the effect, toggle, sound and
// needed-weapon columns compare equal.
func (d *SkillData) Compare(o *SkillData, tag Tag) int {
	if skillNoCompare[tag] || strings.HasPrefix(string(tag), "nwp") {
		return 0
	}
	a := d.ptr(tag)
	if a == nil {
		return 0
	}
	return cmp(getInt(a), getInt(o.ptr(tag)))
}

func (d *SkillData) flntCells() []any {
	c := []any{&d.Level, &d.Icon, &d.Anim, &d.Effect, &d.ToggleType, &d.Sound, &d.ReqLevel, &d.Country}
	for i := range d.Classes {
		c = append(c, &d.Classes[i])
	}
	c = append(c, &d.Grow, &d.Points, &d.TypeShow, &d.TypeAttack, &d.TypeEffect, &d.TypeDetail)
	for i := range d.NeedWeapon {
		c = append(c, &d.NeedWeapon[i])
	}
	c = append(c, &d.Shield, &d.SP, &d.MP, &d.ReadyTime, &d.ResetTime, &d.AttackRange, &d.StateType,
		&d.AttribType, &d.Disable, &d.PrevSkill, &d.SuccessType, &d.SuccessValue, &d.TargetType,
		&d.ApplyRange, &d.MultiAttack, &d.KeepTime, &d.Weapon[0], &d.Weapon[1], &d.WeaponValue, &d.Bag,
		&d.Arrow, &d.DamageType, &d.Damage[0], &d.Damage[1], &d.Damage[2], &d.TimeDamageType,
		&d.TimeDamage[0], &d.TimeDamage[1], &d.TimeDamage[2], &d.AddDamage[0], &d.AddDamage[1], &d.AddDamage[2])
	for i := range d.Ability {
		c = append(c, &d.Ability[i].Type, &d.Ability[i].Value)
	}
	return append(c, &d.Heal[0], &d.Heal[1], &d.Heal[2], &d.TimeHeal[0], &d.TimeHeal[1], &d.TimeHeal[2],
		&d.DefenceType, &d.DefenceValue, &d.LimitHP, &d.FixRange, &d.ChangeType, &d.ChangeLevel)
}

// Read reads a level (SkillData::read).
func (d *SkillData) Read(r *binstream.Reader, f Format) {
	start := r.Pos
	switch f {
	case File:
		d.Level = r.Byte()
		d.Icon = r.Int16()
		d.Anim = r.Byte()
		d.Effect = r.Int16()
		d.ToggleType = r.Int16()
		if d.Version&0xFFF < 0x600 {
			d.Sound = 0
		} else {
			d.Sound = r.Byte()
		}
		d.ReqLevel = r.Int16()
		d.Country = r.Byte()
		r.Read(d.Classes[:])
		d.Grow = r.Byte()
		d.Points = r.Byte()
		d.TypeShow = r.Byte()
		d.TypeAttack = r.Byte()
		d.TypeEffect = r.Byte()
		d.TypeDetail = r.Int16()
		r.Read(d.NeedWeapon[:])
		d.Shield = r.Byte()
		d.SP = r.Int16()
		d.MP = r.Int16()
		d.ReadyTime = r.Byte()
		d.ResetTime = r.Int16()
		d.AttackRange = r.Byte()
		d.StateType = r.Byte()
		d.AttribType = r.Byte()
		d.Disable = r.Int16()
		d.PrevSkill = r.Int16()
		d.SuccessType = r.Byte()
		d.SuccessValue = r.Byte()
		d.TargetType = r.Byte()
		d.ApplyRange = r.Byte()
		d.MultiAttack = r.Byte()
		d.KeepTime = r.Uint16()
		d.Weapon[0] = r.Byte()
		d.Weapon[1] = r.Byte()
		d.WeaponValue = r.Byte()
		d.Bag = r.Byte()
		d.Arrow = r.Int16()
		d.DamageType = r.Byte()
		for i := range d.Damage {
			d.Damage[i] = r.Int16()
		}
		d.TimeDamageType = r.Byte()
		for i := range d.TimeDamage {
			d.TimeDamage[i] = r.Int16()
		}
		for i := range d.AddDamage {
			d.AddDamage[i] = r.Int16()
		}
		n := 3
		if d.Version&0xFFF >= 0x600 {
			n = 10
		}
		for i := 0; i < n; i++ {
			d.Ability[i].Type = r.Byte()
			d.Ability[i].Value = r.Int16()
		}
		for i := range d.Heal {
			d.Heal[i] = r.Uint16()
		}
		for i := range d.TimeHeal {
			d.TimeHeal[i] = r.Uint16()
		}
		d.DefenceType = r.Byte()
		d.DefenceValue = r.Byte()
		d.LimitHP = r.Byte()
		d.FixRange = r.Byte()
		d.ChangeType = r.Int16()
		d.ChangeLevel = r.Int16()
	case Flnt:
		d.SkillId = int32(int16(r.Int64()))
		for _, p := range d.flntCells() {
			setInt(p, int(r.Int64()))
		}
		w := binstream.NewWriter(0)
		d.writeFlnt(w)
		d.flnt = rememberRow(r.Data[start:r.Pos], w.Bytes())
	}
}

func (d *SkillData) writeFlnt(w *binstream.Writer) {
	w.Int64(int64(d.SkillId))
	for _, p := range d.flntCells() {
		w.Int64(int64(getInt(p)))
	}
}

// Write writes a level (SkillData::write).
func (d *SkillData) Write(w *binstream.Writer, f Format) {
	switch f {
	case File:
		w.Byte(d.Level)
		w.Int16(d.Icon)
		w.Byte(d.Anim)
		w.Int16(d.Effect)
		w.Int16(d.ToggleType)
		if d.Version&0xFFF >= 0x600 {
			w.Byte(d.Sound)
		}
		w.Int16(d.ReqLevel)
		w.Byte(d.Country)
		w.Put(d.Classes[:])
		w.Put([]byte{d.Grow, d.Points, d.TypeShow, d.TypeAttack, d.TypeEffect})
		w.Int16(d.TypeDetail)
		w.Put(d.NeedWeapon[:])
		w.Byte(d.Shield)
		w.Int16(d.SP)
		w.Int16(d.MP)
		w.Byte(d.ReadyTime)
		w.Int16(d.ResetTime)
		w.Put([]byte{d.AttackRange, d.StateType, d.AttribType})
		w.Int16(d.Disable)
		w.Int16(d.PrevSkill)
		w.Put([]byte{d.SuccessType, d.SuccessValue, d.TargetType, d.ApplyRange, d.MultiAttack})
		w.Uint16(d.KeepTime)
		w.Put([]byte{d.Weapon[0], d.Weapon[1], d.WeaponValue, d.Bag})
		w.Int16(d.Arrow)
		w.Byte(d.DamageType)
		for _, v := range d.Damage {
			w.Int16(v)
		}
		w.Byte(d.TimeDamageType)
		for _, v := range d.TimeDamage {
			w.Int16(v)
		}
		for _, v := range d.AddDamage {
			w.Int16(v)
		}
		n := 3
		if d.Version&0xFFF >= 0x600 {
			n = 10
		}
		for i := 0; i < n; i++ {
			w.Byte(d.Ability[i].Type)
			w.Int16(d.Ability[i].Value)
		}
		for _, v := range d.Heal {
			w.Uint16(v)
		}
		for _, v := range d.TimeHeal {
			w.Uint16(v)
		}
		w.Put([]byte{d.DefenceType, d.DefenceValue, d.LimitHP, d.FixRange})
		w.Int16(d.ChangeType)
		w.Int16(d.ChangeLevel)
	case Flnt:
		start := w.Len()
		d.writeFlnt(w)
		d.flnt.restore(w.Buf[start:])
	}
}

// exportSeparator returns the field separator of an export format
// (skillExportSeparator).
func exportSeparator(format string) string {
	switch format {
	case "dump", "----", "csv ":
		return ";"
	case "tabs":
		return "\t"
	}
	return ","
}

// Line exports the level (exportText).
func (d *SkillData) Line(format string) string {
	s := exportSeparator(format)
	var v []string
	add := func(x ...any) {
		for _, e := range x {
			v = append(v, fmt.Sprint(e))
		}
	}
	if format == TextCSV {
		add(d.Icon, d.Anim)
	}
	add(d.Country)
	for _, c := range d.Classes {
		add(c)
	}
	add(d.PrevSkill, d.ReqLevel, d.Grow, d.Points, d.TypeShow, d.TypeAttack, d.TypeEffect, d.TypeDetail)
	for _, c := range d.NeedWeapon {
		add(c)
	}
	add(d.Shield, d.SP, d.MP, d.ReadyTime, d.ResetTime, d.AttackRange, d.StateType, d.AttribType, d.Disable,
		d.SuccessType, d.SuccessValue, d.TargetType, d.ApplyRange, d.MultiAttack, int16(d.KeepTime),
		d.Weapon[0], d.Weapon[1], d.WeaponValue, d.Bag, d.Arrow, d.DamageType, d.Damage[0], d.Damage[1],
		d.Damage[2], d.TimeDamageType, d.TimeDamage[0], d.TimeDamage[1], d.TimeDamage[2], d.AddDamage[0],
		d.AddDamage[1], d.AddDamage[2])
	n := 3
	if d.Version&0xFFF >= 0x600 {
		n = 10
	}
	for i := 0; i < n; i++ {
		add(d.Ability[i].Type, d.Ability[i].Value)
	}
	add(d.Heal[0], d.Heal[1], d.Heal[2], d.TimeHeal[0], d.TimeHeal[1], d.TimeHeal[2], d.DefenceType,
		d.DefenceValue, d.LimitHP, d.FixRange, d.ChangeType, d.ChangeLevel)
	if format == TextSQL {
		v = append(v, "GetDate()")
	}
	return strings.Join(v, s)
}
