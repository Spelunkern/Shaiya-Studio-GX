// Package content gives the rest of the program one way to read and write
// client files, whether the workspace is an extracted folder
// (DataFileProvider) or a data.saf archive (ArchiveProvider). Ported from
// shStudio src/archive/ContentProvider, DataFileProvider, ArchiveProvider
// and ContentFiles.
package content

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/crypto/seed"
)

// Provider is a source of client files (ContentProvider).
type Provider interface {
	IsArchive() bool
	// Folder is the workspace folder; it is empty for an archive.
	Folder() string
	// Directory is the workspace folder or the data.saf path.
	Directory() string
	Root() *archive.Root
	CodePage() int
	// ReadFile returns the content of f, decrypting BinarySData files when
	// decrypt is set (the header is kept in f.Key).
	ReadFile(f *archive.File, decrypt bool) ([]byte, error)
	// ReadPath reads a file by disk path or archive path.
	ReadPath(path string) ([]byte, error)
	// WriteFile stores data as the content of f, encrypting it with f.Key
	// when encrypt is set.
	WriteFile(f *archive.File, data []byte, encrypt bool) error
}

// Current is the active provider (g_contentProvider), or nil.
var Current Provider

// CurrentFolder returns Current.Folder() or "" (getContentFolder).
func CurrentFolder() string {
	if Current == nil {
		return ""
	}
	return Current.Folder()
}

// CurrentDirectory returns Current.Directory() or "" (getContentDirectory).
func CurrentDirectory() string {
	if Current == nil {
		return ""
	}
	return Current.Directory()
}

// decryptInto mirrors sub_418C00 with a key buffer: on an encrypted file it
// stores the header in f.Key and returns the plain data.
func decryptInto(f *archive.File, data []byte) ([]byte, error) {
	plain, hdr, enc, err := seed.DecryptFile(data)
	if err != nil {
		return nil, err
	}
	if enc {
		copy(f.Key[:], hdr[:])
	}
	return plain, nil
}

// EncryptDataFile returns key + the encrypted data (encryptDataFile); the
// updated header is stored back in key.
func EncryptDataFile(data []byte, key *[64]byte) ([]byte, error) {
	hdr, body, err := seed.EncryptFile(data, seed.Header(*key))
	if err != nil {
		return nil, err
	}
	*key = hdr
	return append(hdr[:], body...), nil
}

// LoadDataFile reads f from disk (loadDataFile).
func LoadDataFile(f *archive.File, decrypt bool) ([]byte, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}
	if !decrypt {
		return data, nil
	}
	return decryptInto(f, data)
}

// Folder provider -----------------------------------------------------------

// FolderProvider serves an extracted workspace folder (DataFileProvider).
type FolderProvider struct {
	root *archive.Root
	cp   int
	dir  string
}

// NewFolderProvider returns a provider for the workspace dir, whose tree is
// root.
func NewFolderProvider(root *archive.Root, dir string, cp int) *FolderProvider {
	return &FolderProvider{root: root, cp: cp, dir: dir}
}

func (p *FolderProvider) IsArchive() bool     { return false }
func (p *FolderProvider) Folder() string      { return p.dir }
func (p *FolderProvider) Directory() string   { return p.dir }
func (p *FolderProvider) Root() *archive.Root { return p.root }
func (p *FolderProvider) CodePage() int       { return p.cp }

func (p *FolderProvider) ReadFile(f *archive.File, decrypt bool) ([]byte, error) {
	return LoadDataFile(f, decrypt)
}

// ReadPath reads an absolute path, or a path relative to the workspace.
func (p *FolderProvider) ReadPath(path string) ([]byte, error) {
	if path == "" {
		return nil, os.ErrNotExist
	}
	if filepath.IsAbs(path) || isWindowsAbs(path) {
		return os.ReadFile(path)
	}
	return os.ReadFile(filepath.Join(p.dir, ToOSPath(path)))
}

func (p *FolderProvider) WriteFile(f *archive.File, data []byte, encrypt bool) error {
	if encrypt {
		var err error
		if data, err = EncryptDataFile(data, &f.Key); err != nil {
			return err
		}
	}
	return os.WriteFile(f.Path, data, 0o644)
}

// Archive provider ----------------------------------------------------------

// ArchiveProvider serves a data.saf archive (ArchiveProvider).
type ArchiveProvider struct {
	Archive *archive.Archive
	dir     string
}

// NewArchiveProvider opens safPath. Call Load to read the index.
func NewArchiveProvider(safPath string, cp int) (*ArchiveProvider, error) {
	a := archive.New(cp)
	if err := a.Open(safPath); err != nil {
		return nil, err
	}
	return &ArchiveProvider{Archive: a, dir: safPath}, nil
}

// characterLists are the .mlx files the provider adds to the tree when the
// archive does not have them (ArchiveProvider::open).
var characterLists = []struct{ dir, names string }{
	{`Character\DeathEater\`, "demf.mlx dewf.mlx demr.mlx dewr.mlx"},
	{`Character\Elf\`, "elmm.mlx elwm.mlx elmr.mlx elwr.mlx"},
	{`Character\Human\`, "humf.mlx huwf.mlx humm.mlx huwm.mlx"},
	{`Character\Vile\`, "vimm.mlx viwm.mlx vimr.mlx viwr.mlx"},
}

// Load reads the index (the ReadArchiveDialog step) and adds the character
// .mlx entries.
func (p *ArchiveProvider) Load(n archive.Notifier) error {
	if err := p.Archive.ReadIndex(n); err != nil {
		p.Archive.Root.Reset()
		return err
	}
	for _, c := range characterLists {
		for _, name := range strings.Fields(c.names) {
			AddArchiveFile(p.Archive.Root, c.dir, name, p.Archive.CodePage)
		}
	}
	return nil
}

func (p *ArchiveProvider) IsArchive() bool     { return true }
func (p *ArchiveProvider) Folder() string      { return "" }
func (p *ArchiveProvider) Directory() string   { return p.dir }
func (p *ArchiveProvider) Root() *archive.Root { return p.Archive.Root }
func (p *ArchiveProvider) CodePage() int       { return p.Archive.CodePage }

// ReadFile reads f from the archive. A file that is not stored in the
// archive (offset and size 0, except .mlx) is read from disk.
func (p *ArchiveProvider) ReadFile(f *archive.File, decrypt bool) ([]byte, error) {
	if f.Offset == 0 && f.Size == 0 {
		name := f.Name
		if f.Path != "" {
			name = f.Path
		}
		if !strings.HasSuffix(strings.ToLower(name), "mlx") {
			return LoadDataFile(f, decrypt)
		}
	}
	data, err := p.Archive.ReadFile(f)
	if err != nil || !decrypt {
		return data, err
	}
	return decryptInto(f, data)
}

// ReadPath reads a disk file when it exists, otherwise the archive file at
// that path.
func (p *ArchiveProvider) ReadPath(path string) ([]byte, error) {
	if path == "" {
		return nil, os.ErrNotExist
	}
	if path[0] != '\\' {
		if _, err := os.Stat(path); err == nil {
			return os.ReadFile(path)
		}
	}
	if f := p.Archive.Root.FindPath(path); f != nil {
		return p.Archive.ReadFile(f)
	}
	return nil, errors.New("content: " + path + " not found")
}

func (p *ArchiveProvider) WriteFile(f *archive.File, data []byte, encrypt bool) error {
	if encrypt {
		var err error
		if data, err = EncryptDataFile(data, &f.Key); err != nil {
			return err
		}
	}
	return p.Archive.WriteFile(f, append([]byte(nil), data...))
}

// AddFile adds the disk file path to folder (ArchiveProvider::addFile).
func (p *ArchiveProvider) AddFile(folder *archive.Folder, path string) (*archive.File, error) {
	return p.Archive.AddFile(folder, path)
}

// AddArchiveFile adds an empty entry dir+name to the tree when it is not
// there yet (addArchiveFile).
func AddArchiveFile(root *archive.Root, dir, name string, cp int) {
	if root.FindPath(dir+name) != nil {
		return
	}
	folder := root.FindFolderPath(dir)
	if folder == nil {
		return
	}
	f := archive.NewFile(folder, name, cp)
	f.Path = dir + name
	folder.Files = append(folder.Files, f)
}

// ToOSPath converts an archive path to the OS separator.
func ToOSPath(p string) string {
	if os.PathSeparator == '\\' {
		return p
	}
	return strings.ReplaceAll(p, `\`, "/")
}

func isWindowsAbs(p string) bool {
	return len(p) >= 2 && (p[1] == ':' || strings.HasPrefix(p, `\\`))
}

// File types of the tree (getContentFileType).
const (
	TypeNone    = -1
	TypeText    = 2
	TypeSound   = 3
	TypeList    = 4
	TypeConfig  = 5
	TypeImage   = 6
	TypeModel   = 7
	TypeObject  = 8
	TypeWorld   = 9
	TypeDungeon = 10
	TypeSData   = 11
	TypeMisc    = 12
	TypeEffect  = 13
	TypeCursor  = 14
	TypeOther   = 15
	TypeSVMap   = 20
)

// FileType returns the tree icon type of fileName (getContentFileType). For
// .wld it reads the magic of the disk file when it exists.
func FileType(fileName string) int {
	ext := strings.ToLower(filepath.Ext(strings.ReplaceAll(fileName, `\`, "/")))
	if ext == "" {
		return TypeNone
	}
	switch ext[1:] {
	case "bin":
		return TypeNone
	case "txt", "ctl", "dat", "wtr", "lua":
		return TypeText
	case "wav", "wave":
		return TypeSound
	case "alt", "mlt", "mlx", "mon", "itm":
		return TypeList
	case "cfg":
		return TypeConfig
	case "bmp", "dds", "jpg", "tga":
		return TypeImage
	case "3de", "3dc", "3do":
		return TypeModel
	case "dg", "smod", "vani":
		return TypeObject
	case "wld":
		f, err := os.Open(fileName)
		if err != nil {
			return TypeWorld
		}
		var magic [4]byte
		f.Read(magic[:])
		f.Close()
		if string(magic[:]) == "DUN\x00" {
			return TypeDungeon
		}
		return TypeWorld
	case "sdata":
		return TypeSData
	case "ani", "ini", "seff", "zon":
		return TypeMisc
	case "eft":
		return TypeEffect
	case "cur":
		return TypeCursor
	case "svmap":
		return TypeSVMap
	}
	return TypeOther
}
