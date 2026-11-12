package archive

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// HeaderSize is the size of the SAH header: "SAH", version, file count and
// 40 reserved bytes.
const HeaderSize = 51

// HolesFile is the name of the root file where shStudio keeps the list of
// free regions of data.saf.
const HolesFile = "holes.bin"

// DecodeIndex decodes an obfuscated data.sah (decodeSahIndex): each byte is
// XORed with a rolling key and nibble-swapped.
func DecodeIndex(in []byte) []byte {
	key := [16]byte{0xFE, 0xFF, 0xFC, 0xFD, 0xFA, 0xFB, 0xF8, 0xF9, 0xF6, 0xF7, 0xF4, 0xF5, 0xF2, 0xF3, 0xF0, 0xF1}
	out := make([]byte, 0, len(in))
	size := len(in)
	pos, done := 0, 0
	if size > 16 {
		limit := 16
		for {
			for i := 0; i < 16; i++ {
				k := key[i]
				b := k ^ in[pos+i]
				out = append(out, b<<4|b>>4)
				key[i] = k - 16
			}
			pos += 16
			done += 16
			limit += 16
			if limit >= size {
				break
			}
		}
	}
	for k := 0; done < size; k++ {
		b := key[k] ^ in[pos]
		out = append(out, b<<4|b>>4)
		pos++
		done++
	}
	return out
}

func isSAH(b []byte) bool { return len(b) >= 3 && b[0] == 'S' && b[1] == 'A' && b[2] == 'H' }

func isEncodedSAH(b []byte) bool { return len(b) >= 3 && b[0] == 0xCB && b[1] == 0xEB && b[2] == 0x78 }

// Root is the root folder of an archive or of a scanned disk folder
// (FolderRoot).
type Root struct{ Folder }

// stripPrefix removes the root path (or a leading "\\\\") from rel. Disk
// paths of a scanned workspace may use the OS separator; the result always
// uses Sep.
func (r *Root) stripPrefix(rel string) string {
	if r.Path != "" && strings.HasPrefix(rel, r.Path) {
		rel = rel[len(r.Path):]
		if os.PathSeparator != '\\' {
			rel = strings.ReplaceAll(rel, string(os.PathSeparator), Sep)
		}
		return rel
	}
	if strings.HasPrefix(rel, `\\`) {
		return strings.TrimLeft(rel, `\`)
	}
	return rel
}

// FindPath returns the file at an archive path such as
// "data\\item\\item.sdata" or "item\\item.sdata" (FolderRoot::findFile).
func (r *Root) FindPath(path string) *File {
	if strings.HasPrefix(path, `data\`) {
		path = path[5:]
	}
	rel := r.stripPrefix(path)
	d := &r.Folder
	for {
		slash := strings.Index(rel, Sep)
		if slash < 0 {
			return d.ChildFile(rel)
		}
		head := rel[:slash]
		rel = rel[slash+1:]
		if head != "" {
			d = d.ChildFolder(head)
		}
		if d == nil {
			return nil
		}
	}
}

// FindFolderPath returns the folder at path (FolderRoot::findFolderPath).
func (r *Root) FindFolderPath(path string) *Folder {
	rel := r.stripPrefix(path)
	d := &r.Folder
	for {
		if rel == "" {
			return d
		}
		slash := strings.Index(rel, Sep)
		if slash < 0 {
			return nil
		}
		head := rel[:slash]
		rel = rel[slash+1:]
		d = d.ChildFolder(head)
		if d == nil {
			return nil
		}
	}
}

// Hole is a free region of data.saf.
type Hole struct {
	Offset int64
	Size   int32
}

// Holes is the free-region list stored in holes.bin.
type Holes struct {
	Capacity int
	Items    []Hole // len(Items) is the count
}

func (h *Holes) add(offset int64, size int32) {
	if len(h.Items) == h.Capacity {
		h.Capacity += 128
	}
	h.Items = append(h.Items, Hole{offset, size})
}

func (h *Holes) consume(offset int64, size int32) {
	for i := range h.Items {
		if h.Items[i].Offset != offset {
			continue
		}
		it := &h.Items[i]
		if it.Size-size >= it.Size>>3 {
			it.Offset += int64(size)
			it.Size -= size
		} else {
			h.Items = append(h.Items[:i], h.Items[i+1:]...)
		}
		return
	}
}

func (h *Holes) findBest(size int32) int64 {
	best := int32(0x7FFFFFFF)
	var res int64
	for _, it := range h.Items {
		if it.Size >= size && it.Size-size < best {
			res, best = it.Offset, it.Size-size
		}
	}
	return res
}

func (h *Holes) read(data []byte) error {
	return binstream.Try(func() {
		r := binstream.NewReader(data, 0)
		h.Capacity = r.Int()
		count := r.Int()
		if h.Capacity < 0 || count < 0 || count > h.Capacity {
			binstream.Fail("invalid holes.bin")
		}
		h.Items = make([]Hole, 0, h.Capacity)
		for i := 0; i < h.Capacity; i++ {
			off := r.Int64()
			sz := r.Int32()
			if i < count {
				h.Items = append(h.Items, Hole{off, sz})
			}
		}
	})
}

func (h *Holes) bytes() []byte {
	w := binstream.NewWriter(0)
	w.Int(h.Capacity)
	w.Int(len(h.Items))
	for i := 0; i < h.Capacity; i++ {
		var it Hole
		if i < len(h.Items) {
			it = h.Items[i]
		}
		w.Int64(it.Offset)
		w.Int32(it.Size)
	}
	return w.Bytes()
}

// Archive is an open data.sah/data.saf pair (SahArchive).
type Archive struct {
	Root      *Root
	Holes     Holes
	CodePage  int
	FileCount int
	Writable  bool
	SahPath   string
	SafPath   string

	file     *os.File
	fileSize int64
	index    []byte // index body after the header, without the trailer
}

// Errors returned by Archive methods.
var (
	ErrNotFound    = errors.New("archive: file not found")
	ErrBadIndex    = errors.New("archive: the data.sah file doesn't start with 'SAH'")
	ErrNotOpen     = errors.New("archive: not open")
	ErrReadOnly    = errors.New("archive: data.saf is read-only")
	ErrExists      = errors.New("archive: a file with this name already exists")
	ErrTooLarge    = errors.New("archive: file too large")
	ErrShortRead   = errors.New("archive: short read")
	ErrNoIndex     = errors.New("archive: no index")
	ErrBufferSmall = errors.New("archive: buffer too small")
)

// New returns a closed archive whose strings use code page cp.
func New(cp int) *Archive { return &Archive{Root: &Root{}, CodePage: cp} }

// SahPathOf returns the data.sah path of a data.saf path (last letter
// replaced by 'h').
func SahPathOf(saf string) string {
	if saf == "" {
		return saf
	}
	return saf[:len(saf)-1] + "h"
}

// Close closes the archive and clears the tree.
func (a *Archive) Close() {
	if a.file != nil {
		a.file.Close()
		a.file = nil
	}
	a.Root.Reset()
	a.SahPath, a.SafPath = "", ""
	a.fileSize, a.FileCount = 0, 0
}

// Open opens safPath read-write (or read-only when that fails) and loads
// the index header of the matching .sah (SahArchive::open). Call ReadIndex
// to build the tree.
func (a *Archive) Open(safPath string) error {
	a.Close()
	if _, err := os.Stat(safPath); err != nil {
		return err
	}
	a.SafPath = safPath
	a.SahPath = SahPathOf(safPath)
	f, err := os.OpenFile(safPath, os.O_RDWR, 0)
	if err != nil {
		f, err = os.Open(safPath)
		if err != nil {
			return err
		}
		a.Writable = false
	} else {
		a.Writable = true
	}
	a.file = f
	st, err := f.Stat()
	if err != nil {
		a.Close()
		return err
	}
	a.fileSize = st.Size()
	data, err := os.ReadFile(a.SahPath)
	if err != nil {
		a.Close()
		return err
	}
	if len(data) < HeaderSize {
		a.Close()
		return ErrBadIndex
	}
	if !isSAH(data) {
		if !isEncodedSAH(data) {
			a.Close()
			return ErrBadIndex
		}
		data = DecodeIndex(data)
		if !isSAH(data) {
			a.Close()
			return ErrBadIndex
		}
	}
	a.FileCount = int(int32(binary.LittleEndian.Uint32(data[7:])))
	end := len(data) - 8
	if end < HeaderSize {
		end = HeaderSize
	}
	a.index = data[HeaderSize:end]
	return nil
}

// IsOpen reports whether data.saf is open.
func (a *Archive) IsOpen() bool { return a.file != nil }

// ReadIndex builds the tree from the index and loads holes.bin
// (SahArchive::readIndex).
func (a *Archive) ReadIndex(n Notifier) error {
	if a.file == nil || len(a.index) == 0 {
		return ErrNoIndex
	}
	r := binstream.NewReader(a.index, a.CodePage)
	err := binstream.Try(func() {
		a.Root.Folder = Folder{CodePage: a.CodePage}
		a.Root.ReadTree(r, n)
	})
	if err != nil {
		a.Root.Reset()
		return err
	}
	if h := a.Root.ChildFile(HolesFile); h != nil {
		if data, err := a.ReadFile(h); err == nil {
			a.Holes.read(data)
		}
	}
	return nil
}

// ReadFile reads the content of f from data.saf.
func (a *Archive) ReadFile(f *File) ([]byte, error) {
	if a.file == nil {
		return nil, ErrNotOpen
	}
	if f.Size > 0x7FFFFFFF {
		return nil, ErrTooLarge
	}
	buf := make([]byte, f.Size)
	n, err := a.file.ReadAt(buf, f.Offset)
	if n != len(buf) {
		if err == nil {
			err = ErrShortRead
		}
		return nil, fmt.Errorf("archive: reading %s: %w", f.Name, err)
	}
	return buf, nil
}

func (a *Archive) writeAt(off int64, data []byte) error {
	if a.file == nil {
		return ErrNotOpen
	}
	_, err := a.file.WriteAt(data, off)
	return err
}

func (a *Archive) size() (int64, error) {
	st, err := a.file.Stat()
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// WriteFile replaces the content of f (SahArchive::writeFile). Data that
// fits is written in place; otherwise it goes to the best hole or to the end
// of data.saf, and the old region becomes a hole.
func (a *Archive) WriteFile(f *File, data []byte) error {
	if a.file == nil {
		return ErrNotOpen
	}
	if !a.Writable {
		return ErrReadOnly
	}
	size := uint32(len(data))
	if size <= uint32(f.Size) {
		if err := a.writeAt(f.Offset, data); err != nil {
			return err
		}
		if int64(size) != f.Size {
			f.Size = int64(size)
			return a.SaveIndex()
		}
		return nil
	}
	if hole := a.Holes.findBest(int32(size)); hole != 0 {
		a.Holes.consume(hole, int32(size))
		if err := a.writeAt(hole, data); err != nil {
			return err
		}
		f.Offset = hole
	} else {
		end, err := a.size()
		if err != nil {
			return err
		}
		if err := a.writeAt(end, data); err != nil {
			return err
		}
		a.Holes.add(f.Offset, int32(uint32(f.Size)))
		f.Offset = end
	}
	f.Size = int64(size)
	if err := a.saveHoles(); err != nil {
		return err
	}
	return a.SaveIndex()
}

// AddFile adds the disk file at path to folder (SahArchive::addFile).
func (a *Archive) AddFile(folder *Folder, path string) (*File, error) {
	name := baseName(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if folder.FindFile(name) != nil {
		return nil, ErrExists
	}
	f := NewFile(folder, name, a.CodePage)
	if err := a.WriteFile(f, data); err != nil {
		return nil, err
	}
	folder.Files = append(folder.Files, f)
	f.Path = folder.Path + name
	return f, a.SaveIndex()
}

// IndexBytes returns the full data.sah image for the current tree.
func (a *Archive) IndexBytes() []byte {
	w := binstream.NewWriter(a.CodePage)
	w.Put([]byte("SAH"))
	w.Int(0)
	w.Int(a.Root.CountFiles())
	w.Put(make([]byte, 40))
	w.Put([]byte{1, 0, 0, 0, 0})
	body := binstream.NewWriter(a.CodePage)
	a.Root.WriteTree(body)
	a.index = body.Bytes()
	w.Put(a.index)
	w.Put(make([]byte, 8))
	return w.Bytes()
}

// SaveIndex writes data.sah (SahArchive::saveIndex).
func (a *Archive) SaveIndex() error {
	return os.WriteFile(a.SahPath, a.IndexBytes(), 0o644)
}

// saveHoles writes holes.bin (SahArchive::saveHoles). The previous region
// of holes.bin is filled with 0xBF and becomes a hole when the size changes.
func (a *Archive) saveHoles() error {
	data := a.Holes.bytes()
	h := a.Root.ChildFile(HolesFile)
	if h != nil {
		if h.Size == int64(len(data)) {
			return a.WriteFile(h, data)
		}
		old := int32(uint32(h.Size))
		a.Holes.add(h.Offset, old)
		data = a.Holes.bytes()
		if err := a.writeAt(h.Offset, bytes.Repeat([]byte{0xBF}, int(old))); err != nil {
			return err
		}
	}
	end, err := a.size()
	if err != nil {
		return err
	}
	if err := a.writeAt(end, data); err != nil {
		return err
	}
	if h == nil {
		h = a.Root.NewChildFile(HolesFile, int64(len(data)))
		a.Root.Files = append(a.Root.Files, h)
	}
	h.Offset = end
	h.Size = int64(len(data))
	return nil
}

func baseName(p string) string {
	i := strings.LastIndexAny(p, `\/`)
	return p[i+1:]
}
