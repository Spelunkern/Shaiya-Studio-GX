package seed

import (
	"bytes"
	"os"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	plain := []byte("Shaiya Studio GX round trip test data, not aligned!")
	hdr, body, err := EncryptFile(plain, Header{})
	if err != nil {
		t.Fatal(err)
	}
	img := append(hdr[:], body...)
	got, h2, enc, err := DecryptFile(img)
	if err != nil || !enc {
		t.Fatalf("enc=%v err=%v", enc, err)
	}
	// Like shStudio, a plain size inside the last block yields the whole
	// padded block.
	if len(got) != 64 || !bytes.Equal(got[:len(plain)], plain) {
		t.Fatalf("got %q", got)
	}
	if h2 != hdr {
		t.Fatal("header mismatch")
	}
}

// TestRealFile decrypts and re-encrypts a real client data file. Set
// SHSTUDIO_SDATA to a DB*Data.SData path to run it.
func TestRealFile(t *testing.T) {
	path := os.Getenv("SHSTUDIO_SDATA")
	if path == "" {
		t.Skip("SHSTUDIO_SDATA not set")
	}
	img, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plain, hdr, enc, err := DecryptFile(img)
	if err != nil {
		t.Fatal(err)
	}
	if !enc {
		t.Skip("file is not encrypted")
	}
	// The stored size is exact; DecryptFile may return up to 15 padding bytes.
	h2, body, err := EncryptFile(plain[:hdr.PlainSize()], hdr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(h2[:], body...), img) {
		t.Fatal("re-encrypted image differs")
	}
	t.Logf("plain %d bytes, first: % x", len(plain), plain[:16])
}
