// Package archive implements the Shaiya data.sah/data.saf archive: the
// folder/file tree (FolderTree/FileTree/FolderRoot), reading and in-place
// writing of the archive (SahArchive), building a new archive from a folder
// (ArchiveWriter) and extracting it (ArchiveExtractor).
package archive

import (
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/natural"
)

// Sep is the separator of archive paths, which keep shStudio's Windows form
// ("folder\\sub\\file.ext") on every OS.
const Sep = "\\"

// File state flags (FileTree::State).
const (
	StateLoaded = 0x01
	StateDirty  = 0x02
	StateNewer  = 0x10
)

// File is a file of the archive or of a scanned disk folder (FileTree).
type File struct {
	Parent   *Folder
	Path     string // archive path, or disk path for scanned folders
	Name     string
	Attr     uint32
	Offset   int64
	Size     int64
	Time     time.Time
	Icon     int
	State    int
	Key      [64]byte
	CodePage int
}

// NewFile returns a file named name (FileTree(parent, name, codePage)). A name
// containing a backslash is taken as a path.
func NewFile(parent *Folder, name string, cp int) *File {
	f := &File{Parent: parent, Icon: -1, CodePage: cp}
	if strings.Contains(name, Sep) {
		f.Path = name
		name = name[strings.LastIndex(name, Sep)+1:]
	}
	f.Name = name
	return f
}

// StateText returns the text of the "stat" column.
func (f *File) StateText() string {
	switch {
	case f.State&StateDirty != 0:
		return "dirty"
	case f.State&StateLoaded != 0:
		return "loaded"
	case f.State&StateNewer != 0:
		return "newer"
	}
	return ""
}

func (f *File) read(r *binstream.Reader) {
	*f = File{Icon: 0}
	f.Name = r.String(4, r.CodePage)
	f.Offset = r.Int64()
	f.Size = int64(r.Uint32())
	f.Attr = r.Uint32()
}

func (f *File) write(w *binstream.Writer) {
	w.String(f.Name, 4, w.CodePage)
	w.Int64(f.Offset)
	w.Uint32(uint32(f.Size))
	w.Uint32(f.Attr)
}

// Folder is a folder of the tree (FolderTree). The root of an archive has
// an empty name.
type Folder struct {
	Parent    *Folder
	Path      string // ends with Sep unless empty
	Name      string
	Folders   []*Folder
	Files     []*File
	TotalSize int64
	CodePage  int
}

// NewFolder returns an empty folder.
func NewFolder(name, path string, cp int) *Folder {
	return &Folder{Name: name, Path: path, CodePage: cp}
}

// Reset clears the folder.
func (d *Folder) Reset() {
	d.Path, d.Name = "", ""
	d.Folders, d.Files = nil, nil
	d.TotalSize = 0
}

// HasChildren reports whether the folder has files or folders ("?chl").
func (d *Folder) HasChildren() bool { return len(d.Files)+len(d.Folders) > 0 }

// CountFiles returns the number of files in the whole subtree.
func (d *Folder) CountFiles() int {
	n := len(d.Files)
	for _, s := range d.Folders {
		n += s.CountFiles()
	}
	return n
}

// ComputeSize recomputes TotalSize. Like FolderTree::computeSize it skips
// the first file of each folder.
func (d *Folder) ComputeSize() int64 {
	d.TotalSize = 0
	for i := 1; i < len(d.Files); i++ {
		d.TotalSize += d.Files[i].Size
	}
	for _, s := range d.Folders {
		d.TotalSize += s.ComputeSize()
	}
	return d.TotalSize
}

func equalFold(a, b string) bool { return strings.EqualFold(a, b) }

// ChildFolder returns the direct subfolder named name (case-insensitive).
func (d *Folder) ChildFolder(name string) *Folder {
	for _, s := range d.Folders {
		if equalFold(s.Name, name) {
			return s
		}
	}
	return nil
}

// ChildFile returns the direct file named name (case-insensitive).
func (d *Folder) ChildFile(name string) *File {
	for _, f := range d.Files {
		if equalFold(f.Name, name) {
			return f
		}
	}
	return nil
}

// FindFile searches the subtree depth-first for a file named name
// (FolderTree::findFile).
func (d *Folder) FindFile(name string) *File {
	if f := d.ChildFile(name); f != nil {
		return f
	}
	for _, s := range d.Folders {
		if f := s.FindFile(name); f != nil {
			return f
		}
	}
	return nil
}

// NewChildFile returns a file for this folder (FolderTree::addFile); the
// caller appends it to Files.
func (d *Folder) NewChildFile(name string, size int64) *File {
	return &File{Parent: d, Name: name, Size: size, CodePage: d.CodePage}
}

// NewChildFolder returns a folder for this folder (FolderTree::addFolder);
// the caller appends it to Folders.
func (d *Folder) NewChildFolder(name, parentPath string) *Folder {
	return &Folder{Parent: d, Name: name, Path: parentPath + name + Sep, CodePage: d.CodePage}
}

// ReadTree reads the folder body of the index: files, then subfolders, each
// prefixed by its name (FolderTree::readTree). The name of this folder is
// read first. Progress may return false to cancel.
func (d *Folder) ReadTree(r *binstream.Reader, n Notifier) {
	d.TotalSize = 0
	d.Folders, d.Files = nil, nil
	d.Name = r.String(4, r.CodePage)
	if d.Name != "" {
		d.Path += d.Name + Sep
	}
	if n != nil {
		n.SetStatus(d.Path)
	}
	files := r.Int()
	for i := 0; i < files; i++ {
		f := &File{}
		f.read(r)
		if n != nil && !n.Progress(f.Name, 1) {
			binstream.Fail("cancelled")
		}
		f.Path = d.Path + f.Name
		f.Parent = d
		f.CodePage = d.CodePage
		d.TotalSize += f.Size
		d.Files = append(d.Files, f)
	}
	if len(d.Files) >= 2 {
		quickSort(d.Files, 0, len(d.Files)-1, func(a, b *File) int { return natural.Compare(a.Name, b.Name) })
	}
	folders := r.Uint32()
	if folders > 0x400 {
		binstream.Fail("too many folders: %d", folders)
	}
	for i := 0; i < int(folders); i++ {
		s := &Folder{Parent: d, Path: d.Path, CodePage: d.CodePage}
		s.ReadTree(r, n)
		d.TotalSize += s.TotalSize
		d.Folders = append(d.Folders, s)
	}
}

// WriteTree writes the folder body (FolderTree::writeTree).
func (d *Folder) WriteTree(w *binstream.Writer) {
	w.Int(len(d.Files))
	for _, f := range d.Files {
		f.write(w)
	}
	w.Int(len(d.Folders))
	for _, s := range d.Folders {
		w.String(s.Name, 4, w.CodePage)
		s.WriteTree(w)
	}
}

// SortFiles sorts the files by name in logical order.
func (d *Folder) SortFiles() {
	if len(d.Files) >= 2 {
		quickSort(d.Files, 0, len(d.Files)-1, func(a, b *File) int { return natural.Compare(a.Name, b.Name) })
	}
}

// SortFolders sorts the subfolders by name in logical order.
func (d *Folder) SortFolders() {
	if len(d.Folders) >= 2 {
		quickSort(d.Folders, 0, len(d.Folders)-1, func(a, b *Folder) int { return natural.Compare(a.Name, b.Name) })
	}
}

// Dump appends the text form of the tree (toText). With dump=true it uses
// the tab-separated "dump" format, otherwise the quoted CSV-like one.
func (d *Folder) Dump(sb *strings.Builder, dump bool, indent int) {
	tabs := strings.Repeat("\t", indent)
	if dump {
		sb.WriteString(d.Name + "\\")
	} else {
		sb.WriteString("\"" + d.Name + "\"")
	}
	sb.WriteString("\r\n")
	in := tabs + "\t"
	for _, s := range d.Folders {
		sb.WriteString(in)
		s.Dump(sb, dump, indent+1)
	}
	for _, f := range d.Files {
		sb.WriteString(in)
		if dump {
			sb.WriteString(f.Name + "\t" + itoa(f.Offset) + "\t" + itoa(f.Size))
		} else {
			sb.WriteString("\"" + f.Name + "\";" + itoa(f.Offset) + ";" + itoa(f.Size) + ";" + itoa(int64(int32(f.Attr))))
		}
		sb.WriteString("\r\n")
	}
}

// quickSort is the exact quicksort of ptrArray sort in FolderTree.cpp, kept
// so that files with equal names keep shStudio's order.
func quickSort[T any](items []T, lo, hi int, cmp func(a, b T) int) {
	for lo < hi {
		i, j := lo, hi
		pivot := items[(lo+hi)/2]
		for {
			for ; i < hi; i++ {
				if cmp(items[i], pivot) >= 0 {
					break
				}
			}
			for ; j > 0; j-- {
				if cmp(items[j], pivot) <= 0 {
					break
				}
			}
			if j < i {
				break
			}
			if j != i {
				items[i], items[j] = items[j], items[i]
			}
			j--
			i++
			if i > j {
				break
			}
		}
		if lo < j {
			quickSort(items, lo, j, cmp)
		}
		if i >= hi {
			return
		}
		lo = i
	}
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [24]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// Notifier receives progress while reading, scanning, writing or
// extracting (StreamNotifier). Returning false cancels.
type Notifier interface {
	SetStatus(text string) bool
	Progress(name string, units int64) bool
}
