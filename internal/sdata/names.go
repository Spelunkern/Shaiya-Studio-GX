package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/codepage"
)

// CountryName returns the language name of a country code used for
// NpcQuestTrans files (countryName), or "".
func CountryName(code int) string {
	switch code {
	case 156:
		return "CHINA"
	case 250:
		return "FRANCE"
	case 276:
		return "GERMANY"
	case 380:
		return "ITALY"
	case 410:
		return "KOREA"
	case 616:
		return "POLAND"
	case 792:
		return "TURKEY"
	case 840:
		return "USA"
	}
	return ""
}

// CountryCode returns the language code of a country used for DB*Text
// files (countryCode), or "".
func CountryCode(code int) string {
	switch code {
	case 156:
		return "CHN"
	case 250:
		return "FRC"
	case 276:
		return "GER"
	case 380:
		return "ITA"
	case 410:
		return "KOR"
	case 616:
		return "POL"
	case 792:
		return "TUR"
	case 840:
		return "USA"
	}
	return ""
}

// WeaponTypeName names a weapon item type (weaponTypeName), or "".
func WeaponTypeName(t int) string {
	switch t {
	case 1:
		return "1-hand Sword"
	case 2:
		return "2-hands Sword"
	case 3:
		return "1-hand Axe"
	case 4:
		return "2-hands Axe"
	case 5:
		return "Dual Sword or Axe"
	case 6:
		return "Spear"
	case 7:
		return "1-hand Blunt"
	case 8:
		return "2-hands Blunt"
	case 9:
		return "Reverse Sword"
	case 0xA:
		return "Dagger"
	case 0xB:
		return "Javelin"
	case 0xC:
		return "Staff"
	case 0xD:
		return "Bow"
	case 0xE:
		return "Cross-bow"
	case 0xF:
		return "Knuckle"
	case 0x13, 0x22:
		return "Shield"
	case 0x2D:
		return "1-hand Sword, 2nd set"
	case 0x2E:
		return "2-hands Sword, 2nd set"
	case 0x2F:
		return "1-hand Axe, 2nd set"
	case 0x30:
		return "2-hands Axe, 2nd set"
	case 0x31:
		return "Dual Swords, 2nd set"
	case 0x32:
		return "Dual Axes, 2nd set"
	case 0x33, 0x34:
		return "Spears, 2nd set"
	case 0x35, 0x36:
		return "1-hand Blunt, 2nd set"
	case 0x37, 0x38:
		return "2-hands Blunt, 2nd set"
	case 0x39:
		return "Reverse Sword, 2nd set"
	case 0x3A:
		return "Dagger, 2nd set"
	case 0x3B:
		return "Javelin, 2nd set"
	case 0x3C, 0x3D:
		return "Staff, 2nd set"
	case 0x3E, 0x3F:
		return "Bow, 2nd set"
	case 0x40:
		return "Cross-bow, 2nd set"
	case 0x41:
		return "Knuckle, 2nd set"
	case 0x45, 0x54:
		return "Shield, 2nd set"
	}
	return ""
}

// Name is an entry of a fixed-width name list (Name): 256 bytes per name.
type Name struct {
	ID      int
	File    string
	Display string // the name, with " (missing file)" when flagged
}

// Names is a list of 256-byte names (Names).
type Names struct{ Names []*Name }

// Read loads the list.
func (n *Names) Read(data []byte, cp int, f Format) error {
	n.Names = nil
	return binstream.Try(func() {
		r := binstream.NewReader(data, cp)
		if r.Len() < 4 {
			binstream.Fail("too short")
		}
		count := r.Int()
		if count < 0 || count > r.Len()/256 {
			binstream.Fail("invalid count")
		}
		for i := 0; i < count; i++ {
			s := binstream.CString(r.Bytes(256), cp)
			n.Names = append(n.Names, &Name{ID: i, File: s, Display: s})
		}
	})
}

// Write serializes the list: each name in a NUL-padded 256-byte field
// (see docs/DIFFERENCES.md for writeFixedString256).
func (n *Names) Write(cp int, f Format) ([]byte, error) {
	w := binstream.NewWriter(cp)
	w.Uint32(uint32(len(n.Names)))
	for _, x := range n.Names {
		b := codepage.Encode(x.File, cp)
		if len(b) > 255 {
			b = b[:255]
		}
		field := make([]byte, 256)
		copy(field, b)
		w.Put(field)
	}
	return w.Bytes(), nil
}

// NameOf returns the file name of id.
func (n *Names) NameOf(id int) string {
	for _, x := range n.Names {
		if x.ID == id {
			return x.File
		}
	}
	return ""
}

// SetFileFound updates the display name ("file" flag of setInt).
func (x *Name) SetFileFound(found bool) {
	x.Display = x.File
	if !found {
		x.Display += " (missing file)"
	}
}

// Text exports the list (toText).
func (n *Names) Text(format string) string {
	var b strings.Builder
	for _, x := range n.Names {
		if format == "dump" {
			fmt.Fprintf(&b, "%d:\t%s\r\n", x.ID, x.File)
		} else {
			fmt.Fprintf(&b, "%d%s%s\r\n", x.ID, exportSeparator(format), x.File)
		}
	}
	return b.String()
}
