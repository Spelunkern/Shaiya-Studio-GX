package sdata

import (
	"bytes"
	"os"
	"testing"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/crypto/seed"
)

// archiveFile reads path from the archive in SHSTUDIO_SAF.
func archiveFile(t *testing.T, path string) []byte {
	t.Helper()
	saf := os.Getenv("SHSTUDIO_SAF")
	if saf == "" {
		t.Skip("SHSTUDIO_SAF not set")
	}
	a := archive.New(949)
	if err := a.Open(saf); err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.ReadIndex(nil); err != nil {
		t.Fatal(err)
	}
	f := a.Root.FindPath(path)
	if f == nil {
		t.Skipf("%s not in archive", path)
	}
	data, err := a.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestItemsFileRoundTrip(t *testing.T) {
	data := decrypt(t, archiveFile(t, `Item\Item.SData`))
	s := NewItems()
	if err := s.Read(data, 949, File); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d items, version %#x, max type %d", len(s.Items), s.Version, s.MaxType)
	out, err := s.Write(949, File)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, data) {
		n := 0
		for n < len(out) && n < len(data) && out[n] == data[n] {
			n++
		}
		t.Fatalf("output differs at %d (len %d vs %d)", n, len(out), len(data))
	}
	if it := s.Find(1, 1); it != nil {
		t.Logf("1-1 %q level %d", it.Name, it.Level)
	}
}

func TestItemsDBRoundTrip(t *testing.T) {
	path := os.Getenv("SHSTUDIO_SDATA")
	if path == "" {
		t.Skip("SHSTUDIO_SDATA not set")
	}
	img, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plain, hdr, _, err := seed.DecryptFile(img)
	if err != nil {
		t.Fatal(err)
	}
	plain = plain[:hdr.PlainSize()]
	s := NewItems()
	if err := s.Read(plain, 949, Flnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d DB items, %d columns", len(s.Items), plain[128])
	out, err := s.Write(949, Flnt)
	if err != nil {
		t.Fatal(err)
	}
	if plain[128] == 69 && !bytes.Equal(out, plain) {
		t.Fatal("69-column file: output differs")
	}
	// Like shStudio, extra columns are dropped: the rows must read back the
	// same.
	s2 := NewItems()
	if err := s2.Read(out, 949, Flnt); err != nil {
		t.Fatal(err)
	}
	if len(s2.Items) != len(s.Items) {
		t.Fatalf("re-read %d rows", len(s2.Items))
	}
	for i := range s.Items {
		a, b := *s.Items[i], *s2.Items[i]
		a.flnt, b.flnt = nil, nil
		if a != b {
			t.Fatalf("row %d differs:\n%+v\n%+v", i, a, b)
		}
	}
	out2, _ := s2.Write(949, Flnt)
	if !bytes.Equal(out, out2) {
		t.Fatal("second save differs")
	}
}

// decrypt returns the plain content of a data file (or data itself).
func decrypt(t *testing.T, data []byte) []byte {
	plain, hdr, enc, err := seed.DecryptFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if enc {
		return plain[:hdr.PlainSize()]
	}
	return data
}
