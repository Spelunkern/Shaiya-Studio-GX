package sdata

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// testDir is the folder of sample client files (Parsec's test data); set
// SHSTUDIO_TESTDATA to it.
func testFile(t *testing.T, rel string) []byte {
	t.Helper()
	dir := os.Getenv("SHSTUDIO_TESTDATA")
	if dir == "" {
		t.Skip("SHSTUDIO_TESTDATA not set")
	}
	data, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Skip(err)
	}
	return decrypt(t, data)
}

// table is any SData table.
type table interface {
	Read(data []byte, cp int, f Format) error
	Write(cp int, f Format) ([]byte, error)
}

// roundTrip loads data and checks that saving gives the same bytes.
func roundTrip(t *testing.T, name string, tab table, data []byte, cp int, f Format) {
	t.Helper()
	if err := tab.Read(data, cp, f); err != nil {
		t.Fatalf("%s: read: %v", name, err)
	}
	out, err := tab.Write(cp, f)
	if err != nil {
		t.Fatalf("%s: write: %v", name, err)
	}
	if !bytes.Equal(out, data) {
		n := 0
		for n < len(out) && n < len(data) && out[n] == data[n] {
			n++
		}
		t.Fatalf("%s: output differs at %d (len %d vs %d)", name, n, len(out), len(data))
	}
}

// stable loads data, saves it, loads the result again and checks that a
// second save gives the same bytes: used where shStudio itself does not
// write back identical files (renamed empty rows, DB text NULs, dropped
// extra columns).
func stable(t *testing.T, name string, newTab func() table, data []byte, cp int, f Format) table {
	t.Helper()
	a := newTab()
	if err := a.Read(data, cp, f); err != nil {
		t.Fatalf("%s: read: %v", name, err)
	}
	out1, err := a.Write(cp, f)
	if err != nil {
		t.Fatalf("%s: write: %v", name, err)
	}
	b := newTab()
	if err := b.Read(out1, cp, f); err != nil {
		t.Fatalf("%s: re-read: %v", name, err)
	}
	out2, _ := b.Write(cp, f)
	if !bytes.Equal(out1, out2) {
		t.Fatalf("%s: second save differs", name)
	}
	return a
}

func TestItemSetsRoundTrip(t *testing.T) {
	roundTrip(t, "SetItem", &ItemSets{}, testFile(t, "SetItem/SetItem.SData"), 1252, File)
	s := &ItemSets{}
	roundTrip(t, "DBSetItemData", s, testFile(t, "SetItem/DBSetItemData.SData"), 1252, Flnt)
	text := testFile(t, "SetItem/DBSetItemText_USA.SData")
	stable(t, "DBSetItemText", func() table {
		x := &ItemSets{}
		x.Read(testFile(t, "SetItem/DBSetItemData.SData"), 1252, Flnt)
		return x
	}, text, 1252, Ltxt)
}

func TestItemsSamples(t *testing.T) {
	for _, n := range []string{"ItemEp5.SData", "ItemEp6.SData", "ItemEp64_1251.SData", "Item_ES_ps0224.SData"} {
		s := NewItems()
		roundTrip(t, n, s, testFile(t, "Item/"+n), 1252, File)
		t.Logf("%s: %d items, version %#x", n, len(s.Items), s.Version)
	}
	// This file has rows with type 0, which shStudio renames "???<id>".
	stable(t, "Item_US_ps0216", func() table { return NewItems() }, testFile(t, "Item/Item_US_ps0216.SData"), 1252, File)
	db := testFile(t, "Item/DBItemData.SData")
	s := stable(t, "DBItemText", func() table {
		x := NewItems()
		x.Read(db, 1252, Flnt)
		return x
	}, testFile(t, "Item/DBItemText_USA.SData"), 1252, Ltxt).(*Items)
	if it := s.Find(1, 1); it == nil || it.Name != "Long Sword - Modified by Mati" {
		t.Fatalf("1-1 = %+v", it)
	}
}

func TestCashRoundTrip(t *testing.T) {
	s := NewCashProducts(0x500)
	roundTrip(t, "Cash", s, testFile(t, "Cash/Cash.SData"), 1252, File)
	t.Logf("%d products, bags %v", len(s.Products), s.Bags())
	// shStudio writes the product codes of DB files with a NUL, which the
	// original files do not have.
	db := stable(t, "DBItemSellData", func() table { return NewCashProducts(0) }, testFile(t, "Cash/DBItemSellData.SData"), 1252, Flnt).(*CashProducts)
	t.Logf("%d DB products, version %#x, first %q", len(db.Products), db.Version, db.Products[0].Code)
}

func TestMonstersRoundTrip(t *testing.T) {
	for _, n := range []string{"Monster.SData", "Monster_1252.SData"} {
		s := NewMonsters()
		roundTrip(t, n, s, testFile(t, "Monster/"+n), 1252, File)
		t.Logf("%s: %d monsters, format %#x", n, len(s.Monsters), s.Format)
	}
	db := NewMonsters()
	roundTrip(t, "DBMonsterData", db, testFile(t, "Monster/DBMonsterData.SData"), 1252, Flnt)
	stable(t, "DBMonsterText", func() table {
		x := NewMonsters()
		x.Read(testFile(t, "Monster/DBMonsterData.SData"), 1252, Flnt)
		return x
	}, testFile(t, "Monster/DBMonsterText_USA.SData"), 1252, Ltxt)
	t.Logf("DB: %d monsters, first %q", len(db.Monsters), db.Monsters[0].Name)
}

func TestNpcSkillsRoundTrip(t *testing.T) {
	for _, n := range []string{"NpcSkillEp5.SData", "NpcSkillEp6.SData"} {
		s := NewMobsSkills()
		roundTrip(t, n, s, testFile(t, "NpcSkill/"+n), 1252, File)
		t.Logf("%s: %d skills, format %#x", n, len(s.Skills), s.Format)
	}
	db := NewMobsSkills()
	roundTrip(t, "DBNpcSkillData", db, testFile(t, "NpcSkill/DBNpcSkillData.SData"), 1252, Flnt)
	if err := db.Read(testFile(t, "NpcSkill/DBNpcSkillText_USA.SData"), 1252, Ltxt); err != nil {
		t.Fatal(err)
	}
	t.Logf("DB: %d skills, first %q", len(db.Skills), db.Skills[0].Name)
}

func TestSkillsRoundTrip(t *testing.T) {
	for _, n := range []string{"SkillEp5.SData", "SkillEp6.SData"} {
		s := NewToonSkills()
		roundTrip(t, n, s, testFile(t, "Skill/"+n), 1252, File)
		t.Logf("%s: %d skills, format %#x, first %q", n, len(s.Skills), s.Format, s.Skills[0].Name)
	}
	// This sample has 101 columns; shStudio only reads the 100-column
	// layout and rejects it.
	db := NewToonSkills()
	if err := db.Read(testFile(t, "Skill/DBSkillData.SData"), 1252, Flnt); err == nil {
		t.Fatal("101-column DBSkillData accepted")
	}
	ep6 := NewToonSkills()
	ep6.Read(testFile(t, "Skill/SkillEp6.SData"), 1252, File)
	ep6.TextHead = nil
	if err := ep6.Read(testFile(t, "Skill/DBSkillText_USA.SData"), 1252, Ltxt); err != nil {
		t.Fatal(err)
	}
	t.Logf("text on EP6 table: first %q", ep6.Skills[0].Name)
}
