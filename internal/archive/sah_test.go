package archive

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRealArchive opens a real client archive; set SHSTUDIO_SAF to a
// data.saf path. It only reads.
func TestRealArchive(t *testing.T) {
	saf := os.Getenv("SHSTUDIO_SAF")
	if saf == "" {
		t.Skip("SHSTUDIO_SAF not set")
	}
	a := New(949)
	if err := a.Open(saf); err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.ReadIndex(nil); err != nil {
		t.Fatal(err)
	}
	n := a.Root.CountFiles()
	t.Logf("header count %d, tree count %d, folders %d, holes %d", a.FileCount, n, len(a.Root.Folders), len(a.Holes.Items))
	if n != a.FileCount {
		t.Errorf("count mismatch")
	}
	f := a.Root.FindFile("item.sdata")
	if f == nil {
		t.Fatal("item.sdata not found")
	}
	t.Logf("item.sdata at %s offset %d size %d", f.Path, f.Offset, f.Size)
	if g := a.Root.FindPath(`data\` + f.Path); g != f {
		t.Errorf("FindPath(%q) = %v", f.Path, g)
	}
	data, err := a.ReadFile(f)
	if err != nil || int64(len(data)) != f.Size {
		t.Fatal(err)
	}
	// Re-serializing the sorted tree must parse back to the same tree.
	img := a.IndexBytes()
	b := New(949)
	b.file = a.file
	b.index = img[HeaderSize : len(img)-8]
	if err := b.ReadIndex(nil); err != nil {
		t.Fatal(err)
	}
	if b.Root.CountFiles() != n {
		t.Errorf("re-read count %d", b.Root.CountFiles())
	}
}

func TestWriteInPlaceAndGrow(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	os.MkdirAll(filepath.Join(src, "item"), 0o755)
	os.WriteFile(filepath.Join(src, "item", "a.txt"), []byte("hello"), 0o644)
	os.WriteFile(filepath.Join(src, "b.txt"), []byte("world!"), 0o644)
	saf := filepath.Join(dir, "data.saf")
	root, err := ScanDir(src, 949, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := Build(saf, root, 949, nil); err != nil {
		t.Fatal(err)
	}
	a := New(949)
	if err := a.Open(saf); err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.ReadIndex(nil); err != nil {
		t.Fatal(err)
	}
	f := a.Root.FindPath(`item\a.txt`)
	if f == nil {
		t.Fatal("a.txt missing")
	}
	if err := a.WriteFile(f, []byte("hi")); err != nil {
		t.Fatal(err)
	}
	if err := a.WriteFile(f, []byte("a much longer content than before")); err != nil {
		t.Fatal(err)
	}
	a.Close()
	b := New(949)
	if err := b.Open(saf); err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err := b.ReadIndex(nil); err != nil {
		t.Fatal(err)
	}
	g := b.Root.FindPath(`item\a.txt`)
	data, _ := b.ReadFile(g)
	if string(data) != "a much longer content than before" {
		t.Fatalf("got %q", data)
	}
	if len(b.Holes.Items) == 0 {
		t.Fatal("expected a hole")
	}
	out := filepath.Join(dir, "out")
	if err := Extract(saf, out, nil); err != nil {
		t.Fatal(err)
	}
	if d, _ := os.ReadFile(filepath.Join(out, "b.txt")); string(d) != "world!" {
		t.Fatalf("extract b.txt = %q", d)
	}
}
