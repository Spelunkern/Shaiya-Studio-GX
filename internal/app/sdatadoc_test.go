package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
)

// TestMonsterDocSave opens an encrypted Monster.SData in a workspace,
// saves it and checks that the file is unchanged.
func TestMonsterDocSave(t *testing.T) {
	src := os.Getenv("SHSTUDIO_TESTDATA")
	if src == "" {
		t.Skip("SHSTUDIO_TESTDATA not set")
	}
	orig, err := os.ReadFile(filepath.Join(src, "Monster", "Monster.SData"))
	if err != nil {
		t.Skip(err)
	}
	ws := t.TempDir()
	os.MkdirAll(filepath.Join(ws, "Monster"), 0o755)
	path := filepath.Join(ws, "Monster", "Monster.SData")
	os.WriteFile(path, orig, 0o644)
	root, err := archive.ScanDir(ws, 1252, nil)
	if err != nil {
		t.Fatal(err)
	}
	content.Current = content.NewFolderProvider(root, ws, 1252)
	defer func() { content.Current = nil }()
	f := root.FindFile("Monster.SData")
	m := sdata.NewMonsters()
	v := &monsterView{monsters: m}
	d := &sdataDoc{file: f, data: m, view: v, hasTranslation: true, detectCP: true, decrypt: true}
	v.doc = d
	var loadErr error
	d.load(nil, func(err error) { loadErr = err })
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if len(m.Monsters) == 0 {
		t.Fatal("no monsters")
	}
	if err := d.store(nil); err != nil {
		t.Fatal(err)
	}
	saved, _ := os.ReadFile(path)
	if !bytes.Equal(saved, orig) {
		for i := range saved {
			if saved[i] != orig[i] {
				t.Logf("first diff at %d: % x vs % x", i, saved[i:min(i+16, len(saved))], orig[i:min(i+16, len(orig))])
				break
			}
		}
		t.Fatalf("saved file differs (len %d vs %d)", len(saved), len(orig))
	}
}
