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

func TestNpcQuestRoundTrip(t *testing.T) {
	raw := func() table {
		d := NewNpcQuest()
		d.Options = NpcOptions{}
		return d
	}
	for _, n := range []string{"NpcQuestEp5.SData", "NpcQuestEp6.SData"} {
		d := raw().(*NpcQuest)
		// shStudio escapes the CR LF of texts as "\n" when saving; apart
		// from that the output must be identical.
		in := testFile(t, "NpcQuest/"+n)
		if err := d.Read(in, 1252, File); err != nil {
			t.Fatal(err)
		}
		out, _ := d.Write(1252, File)
		if len(out) != len(in) {
			t.Fatalf("%s: length %d vs %d", n, len(out), len(in))
		}
		for i := 0; i < len(in); i++ {
			if in[i] == out[i] {
				continue
			}
			if string(in[i:i+2]) != "\r\n" || string(out[i:i+2]) != `\n` {
				t.Fatalf("%s: differs at %d", n, i)
			}
			i++
		}
		t.Logf("%s: %d merchants, %d gate keepers, %d quests, format %#x, raw %d bytes", n, len(d.NPCs[1]),
			len(d.NPCs[2]), len(d.Quests.Quests), d.Format, len(d.Raw))
		// With "clean CR" on, shStudio trims CR LF, spaces and CR LF once per
		// load, so a text can lose more on the next save: check that it
		// converges.
		first := NewNpcQuest()
		first.Read(testFile(t, "NpcQuest/"+n), 1252, File)
		out1, _ := first.Write(1252, File)
		second := NewNpcQuest()
		second.Read(out1, 1252, File)
		out2, _ := second.Write(1252, File)
		stable(t, n+" (default options)", func() table { return NewNpcQuest() }, out2, 1252, File)
	}
	d := raw().(*NpcQuest)
	roundTrip(t, "NpcQuestEp8", d, testFile(t, "NpcQuest/NpcQuestEp8.SData"), 1252, Flnt)
	t.Logf("Ep8: %d merchants, %d quests, format %#x", len(d.NPCs[1]), len(d.Quests.Quests), d.Format)
	if err := d.Read(testFile(t, "NpcQuest/NpcQuestTrans_USA.SData"), 1252, Ltxt); err != nil {
		t.Fatal(err)
	}
	q := d.Quests.Quests[0]
	if q.Desc[:10] != "Hey there," || q.Replies[0][:13] != "Welcome back." {
		t.Fatalf("language file texts misplaced: desc %.20q reply %.20q", q.Desc, q.Replies[0])
	}
	t.Logf("merchant 0: %q / %q", d.NPCs[1][0].Name, d.NPCs[1][0].Welcome)
	t.Logf("quest %d: name %q summary %q desc %.60q reply1 %.60q", q.QuestID, q.Name, q.Summary, q.Desc, q.Replies[0])
}

func TestDBTextTables(t *testing.T) {
	for _, n := range []string{"Item/DBItemText_USA.SData", "Monster/DBMonsterText_USA.SData", "Skill/DBSkillText_USA.SData",
		"NpcSkill/DBNpcSkillText_USA.SData", "SetItem/DBSetItemText_USA.SData"} {
		tab := NewDBTextTable()
		roundTrip(t, n, tab, testFile(t, n), 1252, Ltxt)
		t.Logf("%s: %s; field 0 %q, row 0: %q", n, tab.StatusText(), tab.FieldName(0), tab.GetText(0, tab.FieldCount()-1))
	}
}
