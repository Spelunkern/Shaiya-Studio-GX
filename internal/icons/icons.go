// Package icons loads the icon sheets of the client (interface\icon\*.dds)
// and numbers their cells like shStudio's ClientData (ImageStrip and the
// 16x16 ImageList): cells are counted row by row from a first id.
package icons

import (
	"image"
	"strconv"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/imageio"
)

// Sheet is an icon sheet cut in square cells.
type Sheet struct {
	Image   *image.NRGBA
	Cell    int // cell size in pixels (32 for the client sheets)
	FirstID int
	PerRow  int
	Rows    int
	// skip marks cells dropped as blank (ImageStrip); nil for ImageList
	// sheets, which keep every cell.
	skip map[int]bool
	// Tex is the texture of the sheet once uploaded by the UI.
	Tex any
}

// Count returns the number of cells.
func (s *Sheet) Count() int { return s.PerRow * s.Rows }

// Cell returns the rectangle of id, or false when the sheet has no such
// icon.
func (s *Sheet) Rect(id int) (image.Rectangle, bool) {
	i := id - s.FirstID
	if s == nil || i < 0 || i >= s.Count() || s.skip[i] {
		return image.Rectangle{}, false
	}
	x, y := (i%s.PerRow)*s.Cell, (i/s.PerRow)*s.Cell
	return image.Rect(x, y, x+s.Cell, y+s.Cell), true
}

// NewSheet cuts img in cells of size cell, numbered from firstID.
// premultiplied tells whether shStudio would see premultiplied colors
// (32-bit TGA), which matters for the blank cell test.
func NewSheet(img *image.NRGBA, cell, firstID int, dropBlank, premultiplied bool) *Sheet {
	b := img.Bounds()
	s := &Sheet{Image: img, Cell: cell, FirstID: firstID, PerRow: b.Dx() / cell, Rows: b.Dy() / cell}
	if dropBlank {
		s.skip = map[int]bool{}
		for i := 0; i < s.Count(); i++ {
			x, y := (i%s.PerRow)*cell, (i/s.PerRow)*cell
			if isBlankIcon(img, x, y, cell, premultiplied) {
				s.skip[i] = true
			}
		}
	}
	return s
}

func pixel(img *image.NRGBA, x, y int, premul bool) (r, g, b int) {
	o := img.PixOffset(x, y)
	p := img.Pix[o : o+4]
	r, g, b = int(p[0]), int(p[1]), int(p[2])
	if premul {
		a := float64(p[3]) / 255
		r, g, b = int(float64(r)*a), int(float64(g)*a), int(float64(b)*a)
	}
	return
}

func bright(r, g, b int) bool { return r > 0x60 || g > 0x60 || b > 0x60 }

// isBlankIcon detects an empty cell: no bright pixel on the diagonals or
// near the center, and a dark center (isBlankIcon).
func isBlankIcon(img *image.NRGBA, ox, oy, cx int, premul bool) bool {
	cy := cx
	for i := 3; i < cx-3; i++ {
		if bright(pixel(img, ox+i, oy+i, premul)) {
			return false
		}
	}
	for i := 3; i < cx-3; i++ {
		if bright(pixel(img, ox+i, oy+cx-i-1, premul)) {
			return false
		}
	}
	var red, green, blue, n int
	midY := cy / 2
	for x := cx/2 - 4; x <= cx/2+4; x++ {
		for y := midY - 4; y <= midY+4; y++ {
			r, g, b := pixel(img, ox+x, oy+y, premul)
			if bright(r, g, b) {
				return false
			}
			n++
			red += r
			green += g
			blue += b
		}
	}
	if n > 0 {
		red, green, blue = red/n, green/n, blue/n
	}
	return red < 64 && green < 64 && blue < 64
}

// Set is a list of sheets searched in order (an ImageStrip made of several
// files, or the 16x16 ImageList).
type Set struct {
	Sheets []*Sheet
}

// Find returns the sheet and cell of id.
func (s *Set) Find(id int) (*Sheet, image.Rectangle, bool) {
	if s == nil {
		return nil, image.Rectangle{}, false
	}
	for _, sh := range s.Sheets {
		if r, ok := sh.Rect(id); ok {
			return sh, r, true
		}
	}
	return nil, image.Rectangle{}, false
}

// Empty reports whether the set has no sheet.
func (s *Set) Empty() bool { return s == nil || len(s.Sheets) == 0 }

// Reader reads a client file by archive path; it returns nil when absent.
type Reader func(path string) []byte

// Loader loads and caches the icon sets of a workspace (the icon part of
// ClientData).
type Loader struct {
	Read   Reader
	cache  map[string]*Set
	list16 *Set
}

// NewLoader returns a loader reading files with read.
func NewLoader(read Reader) *Loader { return &Loader{Read: read, cache: map[string]*Set{}} }

func ext(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return strings.ToLower(name[i:])
	}
	return ""
}

// loadSheet decodes interface\icon\<name>, trying the other of .dds/.tga
// when the file is missing (loadIconStrip / addIcon).
func (l *Loader) loadSheet(name string, cell, firstID int, dropBlank bool) *Sheet {
	path := `interface\icon\` + name
	data := l.Read(path)
	isDDS := ext(name) == ".dds"
	if data == nil {
		alt := strings.TrimSuffix(path, name[strings.LastIndexByte(name, '.'):])
		if isDDS {
			alt += ".tga"
		} else {
			alt += ".dds"
		}
		data = l.Read(alt)
		isDDS = !isDDS
		if data == nil {
			return nil
		}
	}
	if len(data) < 18 {
		return nil
	}
	premul := false
	var img *image.NRGBA
	var err error
	if isDDS && imageio.HasDDSMagic(data) {
		img, err = imageio.DecodeDDS(data)
	} else {
		img, err = imageio.DecodeTGA(data)
		premul = len(data) > 16 && data[16] == 32
	}
	if err != nil {
		return nil
	}
	return NewSheet(img, cell, firstID, dropBlank, premul)
}

// Strip returns a named 32x32 set (ClientData::getIcons): "icon_skills",
// "icon_rapis", "icon_somo1", "icon_somo2", "icon_somo3", "icon_pet".
func (l *Loader) Strip(name string) *Set {
	if s, ok := l.cache[name]; ok {
		return s
	}
	s := &Set{}
	switch name {
	case "icon_skills":
		for i, n := range []string{"icon_skill", "icon_skill2", "icon_skill3"} {
			if sh := l.skillSheet(n, 1+1000*i); sh != nil {
				s.Sheets = append(s.Sheets, sh)
			}
		}
	case "icon_rapis", "icon_somo1", "icon_somo2", "icon_somo3", "icon_pet":
		if sh := l.loadSheet(name+".dds", 32, 1, true); sh != nil {
			s.Sheets = append(s.Sheets, sh)
		}
	}
	l.cache[name] = s
	return s
}

// skillSheet loads icon_skillN, preferring the file (.dds or .tga) with the
// most non-blank icons (loadSkillIcons).
func (l *Loader) skillSheet(name string, firstID int) *Sheet {
	var best *Sheet
	bestN := -1
	for _, e := range []string{".dds", ".tga"} {
		data := l.Read(`interface\icon\` + name + e)
		if len(data) < 18 {
			continue
		}
		var img *image.NRGBA
		var err error
		premul := false
		if e == ".dds" && imageio.HasDDSMagic(data) {
			img, err = imageio.DecodeDDS(data)
		} else {
			img, err = imageio.DecodeTGA(data)
			premul = data[16] == 32
		}
		if err != nil {
			continue
		}
		sh := NewSheet(img, 32, firstID, true, premul)
		n := sh.Count() - len(sh.skip)
		if n > bestN || (n == bestN && e == ".tga") {
			best, bestN = sh, n
		}
	}
	return best
}

// ItemStrip returns the 32x32 set of an item type (getItemIcons).
func (l *Loader) ItemStrip(t uint8) *Set {
	key := "item" + strconv.Itoa(int(t))
	if s, ok := l.cache[key]; ok {
		return s
	}
	s := &Set{}
	one := func(file string) {
		if sh := l.loadSheet(file, 32, 1, true); sh != nil {
			s.Sheets = append(s.Sheets, sh)
		}
	}
	number := 0
	switch t {
	case 30, 98, 95:
		one("icon_rapis.dds")
	case 25, 41, 38, 42, 43, 44, 94:
		one("icon_somo.dds")
	case 78, 79, 80:
		one("icon_somo1.dds")
	case 27, 28, 29:
		one("icon_quest.dds")
	case 99:
		one("icon_quest2.dds")
	case 100, 101:
		one("icon_somo2.dds")
	case 102, 103:
		one("icon_somo3.dds")
	case 120:
		one("icon_pet.dds")
	case 121:
		one("icon_Wing.dds")
	case 150:
		one("icon_DualLayer.dds")
	default:
		number = itemSheetNumber(t)
	}
	if number > 0 {
		if sh := l.loadSheet(pad2(number)+".dds", 32, 1, true); sh != nil {
			s.Sheets = append(s.Sheets, sh)
		} else if sh := l.loadSheet(strconv.Itoa(number+100)+".dds", 32, 101, true); sh != nil {
			s.Sheets = append(s.Sheets, sh)
		}
	}
	l.cache[key] = s
	return s
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// itemSheetNumber maps an item type to its numbered icon sheet.
func itemSheetNumber(t uint8) int {
	switch t {
	case 1, 45:
		return 1
	case 2, 46:
		return 2
	case 3, 47:
		return 3
	case 4, 48:
		return 4
	case 5, 49, 50:
		return 5
	case 6, 51, 52:
		return 6
	case 7, 53, 54:
		return 7
	case 8, 55, 56:
		return 8
	case 9, 57:
		return 9
	case 10, 58:
		return 10
	case 11, 59:
		return 11
	case 12, 60, 61:
		return 12
	case 13, 62, 63:
		return 13
	case 14, 64:
		return 14
	case 15, 65:
		return 15
	case 16, 66, 72:
		return 16
	case 17, 67, 73:
		return 17
	case 18, 68, 74:
		return 18
	case 19, 69:
		return 19
	case 20, 70, 76:
		return 20
	case 21, 71, 77:
		return 21
	case 22, 37:
		return 22
	case 23, 96:
		return 23
	case 24:
		return 24
	case 31, 81, 87:
		return 31
	case 32, 82, 88:
		return 32
	case 33, 83, 89:
		return 33
	case 34, 84:
		return 34
	case 35, 85, 91:
		return 35
	case 36, 86, 92:
		return 36
	case 39:
		return 39
	case 40, 97:
		return 40
	}
	return 0
}

// List16 returns the 16x16 item icon list (getImageList): every item
// sheet, numbered so that an item's icon is ItemIconBase(type) + icon.
func (l *Loader) List16() *Set {
	if l.list16 != nil {
		return l.list16
	}
	s := &Set{}
	add := func(file string, id int) {
		if sh := l.loadSheet(file, 32, id, false); sh != nil {
			s.Sheets = append(s.Sheets, sh)
		}
	}
	number := 1
	for id := 1101; id <= 15101; id += 1000 {
		add(pad2(number)+".dds", id-100)
		add(pad2(number+100)+".dds", id)
		number++
	}
	number = 16
	for id := 16001; id <= 24001; id += 1000 {
		add(pad2(number)+".dds", id)
		number++
	}
	add("icon_somo.dds", 25001)
	add("icon_quest.dds", 27001)
	add("icon_rapis.dds", 30001)
	number = 31
	for id := 31001; id <= 40001; id += 1000 {
		add(pad2(number)+".dds", id)
		number++
	}
	add("icon_somo1.dds", 78001)
	add("icon_quest2.dds", 99001)
	add("icon_somo2.dds", 100001)
	add("icon_somo3.dds", 102001)
	add("icon_pet.dds", 120001)
	add("icon_Wing.dds", 121001)
	add("icon_DualLayer.dds", 150001)
	l.list16 = s
	return s
}
