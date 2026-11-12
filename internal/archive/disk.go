package archive

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// ErrCancelled is returned when a Notifier cancels an operation.
var ErrCancelled = errors.New("archive: cancelled")

// isHidden mirrors FILE_ATTRIBUTE_HIDDEN on Unix (dot files).
func isHidden(name string) bool { return strings.HasPrefix(name, ".") }

// ScanDir builds a tree of the disk folder dir (FolderRoot::rescan). File
// and folder paths are disk paths; folder paths end with the OS separator.
func ScanDir(dir string, cp int, n Notifier) (*Root, error) {
	root := &Root{Folder: Folder{Path: withSep(dir), CodePage: cp}}
	if err := root.scan(n); err != nil {
		return root, err
	}
	root.ComputeSize()
	return root, nil
}

func withSep(p string) string {
	if p != "" && !strings.HasSuffix(p, string(os.PathSeparator)) {
		p += string(os.PathSeparator)
	}
	return p
}

// scan reads the folder's directory (FolderTree::scan / addFindData).
func (d *Folder) scan(n Notifier) error {
	entries, err := os.ReadDir(d.Path)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() {
			if isHidden(name) {
				continue
			}
			info, err := e.Info()
			if err != nil {
				return err
			}
			if n != nil && !n.Progress(name, info.Size()) {
				return ErrCancelled
			}
			f := d.NewChildFile(name, info.Size())
			f.Path = d.Path + name
			f.Time = info.ModTime()
			d.Files = append(d.Files, f)
			d.TotalSize += info.Size()
			continue
		}
		if name[0] == '.' {
			continue
		}
		s := &Folder{Parent: d, Name: name, Path: d.Path + name + string(os.PathSeparator), CodePage: d.CodePage}
		var serr error
		if n == nil || n.SetStatus(s.Path) {
			serr = s.scan(n)
		} else {
			serr = ErrCancelled
		}
		d.Folders = append(d.Folders, s)
		d.TotalSize += s.TotalSize
		if serr != nil {
			return serr
		}
	}
	d.SortFiles()
	d.SortFolders()
	return nil
}

// Build writes the disk tree root into a new data.saf at safPath and its
// data.sah (ArchiveWriter + StudioFrame "Save the workspace as"). On error
// or cancel the partial data.saf is removed.
func Build(safPath string, root *Root, cp int, n Notifier) (err error) {
	sah := SahPathOf(safPath)
	os.Remove(sah)
	out, err := os.Create(safPath)
	if err != nil {
		return fmt.Errorf("the archive file can not be created: %w", err)
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(safPath)
		}
	}()
	w := &builder{out: out, idx: binstream.NewWriter(cp), n: n, rootLen: len(root.Path)}
	w.idx.Put([]byte("SAH"))
	w.idx.Put(make([]byte, 48))
	w.idx.String("", 4, cp)
	if err = w.folder(&root.Folder); err != nil {
		return err
	}
	w.idx.Int(0)
	w.idx.Int(0)
	img := w.idx.Bytes()
	binary.LittleEndian.PutUint32(img[7:], uint32(w.count))
	return os.WriteFile(sah, img, 0o644)
}

type builder struct {
	out     *os.File
	idx     *binstream.Writer
	n       Notifier
	rootLen int
	offset  int64
	count   int
}

func (b *builder) folder(d *Folder) error {
	if b.n != nil {
		full := d.Path + d.Name
		if len(full) >= b.rootLen {
			full = full[b.rootLen:]
		}
		b.n.SetStatus(full)
	}
	b.idx.Int(len(d.Files))
	for _, f := range d.Files {
		if err := b.file(f); err != nil {
			return err
		}
		if b.n != nil && !b.n.Progress(f.Name, f.Size>>16) {
			return ErrCancelled
		}
	}
	b.idx.Int(len(d.Folders))
	for _, s := range d.Folders {
		b.idx.String(s.Name, 4, b.idx.CodePage)
		if err := b.folder(s); err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) file(f *File) error {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return fmt.Errorf("%s: %w", f.Path, err)
	}
	f.Offset = b.offset
	f.Size = int64(uint32(len(data)))
	if _, err := b.out.Write(data); err != nil {
		return err
	}
	b.offset += int64(len(data))
	f.write(b.idx)
	b.count++
	return nil
}

// Extract expands the archive safPath into dest, reading the index in file
// order (ArchiveExtractor). Folders are created as needed; a file that
// fails is reported in the returned error, which joins all messages.
func Extract(safPath, dest string, n Notifier) error {
	return extract(safPath, dest, n, nil)
}

func extract(safPath, dest string, n Notifier, count *int) error {
	data, err := os.ReadFile(SahPathOf(safPath))
	if err != nil {
		return fmt.Errorf("the indexes (data.sah file) cannot be read: %w", err)
	}
	if len(data) < HeaderSize || !isSAH(data) {
		data = DecodeIndex(data)
		if len(data) < HeaderSize || !isSAH(data) {
			return ErrBadIndex
		}
	}
	saf, err := os.Open(safPath)
	if err != nil {
		return fmt.Errorf("the data.saf file cannot be opened for reading: %w", err)
	}
	defer saf.Close()
	r := binstream.NewReader(data[HeaderSize:], 0)
	x := &extractor{saf: saf, idx: r, n: n, buf: make([]byte, 0x800000)}
	if n != nil {
		if st, err := saf.Stat(); err == nil {
			if rs, ok := n.(interface{ SetRange(int64) }); ok {
				rs.SetRange(st.Size() >> 16)
			}
		}
	}
	root := &Folder{}
	if err := binstream.Try(func() { err = x.folder(root, dest) }); err != nil {
		return err
	}
	if count != nil {
		*count = x.Extracted
	}
	if err != nil {
		return err
	}
	return errors.Join(x.errs...)
}

// ExtractFolder extracts the subtree d of an open archive into dest
// (ArchiveExtractor with m_fromTree).
func (a *Archive) ExtractFolder(d *Folder, dest string, n Notifier) error {
	if a.file == nil {
		return ErrNotOpen
	}
	x := &extractor{saf: a.file, fromTree: true, n: n, buf: make([]byte, 0x800000)}
	if err := x.folder(d, dest); err != nil {
		return err
	}
	return errors.Join(x.errs...)
}

type extractor struct {
	saf       *os.File
	idx       *binstream.Reader
	fromTree  bool
	n         Notifier
	buf       []byte
	cancelled bool
	errs      []error
	Extracted int
}

func (x *extractor) folder(d *Folder, parent string) error {
	path := parent
	if !x.fromTree {
		d.Name = x.idx.String(4, x.idx.CodePage)
	}
	if d.Name != "" {
		path = filepath.Join(path, d.Name)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("the directory %q cannot be created: %w", path, err)
	}
	if x.n != nil && !x.n.SetStatus(path) {
		x.cancelled = true
	}
	var files []*File
	if x.fromTree {
		files = d.Files
	} else {
		count := x.idx.Int()
		for i := 0; i < count; i++ {
			f := &File{}
			f.read(x.idx)
			files = append(files, f)
		}
	}
	for _, f := range files {
		if x.n != nil && !x.n.Progress(f.Name, f.Size>>16) {
			x.cancelled = true
		}
		if x.cancelled {
			return ErrCancelled
		}
		if err := x.file(f, path); err != nil {
			x.errs = append(x.errs, err)
		} else {
			x.Extracted++
		}
	}
	if x.fromTree {
		for _, s := range d.Folders {
			if err := x.folder(s, path); err != nil {
				return err
			}
		}
		return nil
	}
	count := x.idx.Int()
	for i := 0; i < count; i++ {
		if err := x.folder(&Folder{}, path); err != nil {
			return err
		}
	}
	return nil
}

func (x *extractor) file(f *File, dir string) error {
	name := filepath.Join(dir, f.Name)
	out, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("the file %q cannot be created: %w", name, err)
	}
	src := io.NewSectionReader(x.saf, f.Offset, int64(uint32(f.Size)))
	_, err = io.CopyBuffer(out, src, x.buf)
	out.Close()
	if err != nil {
		os.Remove(name)
		return fmt.Errorf("the file %q cannot be read from archive: %w", f.Name, err)
	}
	return nil
}

// ExtractCount is Extract that also returns the number of extracted files.
func ExtractCount(safPath, dest string, n Notifier) (int, error) {
	count := 0
	err := extract(safPath, dest, n, &count)
	return count, err
}
