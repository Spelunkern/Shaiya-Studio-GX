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
