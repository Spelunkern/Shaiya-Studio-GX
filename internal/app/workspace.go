package app

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/filedlg"
)

// client is the "Client" node of the workspace tree (ClientNode): the open
// workspace folder or data.saf archive.
type client struct {
	name     string
	provider content.Provider
	root     *archive.Root
	format   int
	data     clientData
}

func newClient() *client { return &client{name: "Client", root: &archive.Root{}} }

func (c *client) isClosed() bool { return c.provider == nil }

// asn1Now formats the current time like Time::toAsn1 (local time, no zone).
func asn1Now() string { return time.Now().Format("20060102150405") + "Z" }

var safFilter = filedlg.Filter{Name: "Archive Shaiya (*.saf)", Patterns: []string{"*.saf"}}

// isWorkspaceFolder reports whether dir holds an extracted client: an Item
// folder and cash.sdata.
func isWorkspaceFolder(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "Item"))
	if err != nil || !st.IsDir() {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, "cash.sdata"))
	return err == nil || fileExistsFold(dir, "cash.sdata")
}

// fileExistsFold looks for name in dir ignoring case (Windows file names are
// case-insensitive; Linux ones are not).
func fileExistsFold(dir, name string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.EqualFold(e.Name(), name) {
			return true
		}
	}
	return false
}

// cmdOpen handles File > Open Workspace (0xE101) and Open Archive (0xE110).
func (a *App) cmdOpen(archiveFile bool) {
	if !a.closeAll() {
		return
	}
	a.status = ""
	key := "last"
	if archiveFile {
		key = "file"
	}
	location := prefs.Default.String("Workspace", key, "")
	var d *filedlg.Dialog
	if archiveFile {
		d = filedlg.New("Choose an archive to open", filedlg.OpenFile, location, safFilter)
	} else {
		d = filedlg.New("Select a workspace to open", filedlg.SelectFolder, location)
		d.Accept = isWorkspaceFolder
	}
	a.pickFile(d, func(location string) {
		prefs.Default.SetString("Workspace", key, location)
		options := prefs.Default.String("Workspace", location, "")
		cp, have := 0, false
		if i := strings.Index(options, "&cp="); i > 0 {
			if v, err := strconv.Atoi(options[i+4:]); err == nil && v != 0 {
				cp, have = v, true
			}
		}
		open := func(cp int) {
			a.openClient(location, cp, !archiveFile)
		}
		if have {
			open(cp)
			return
		}
		a.chooseCodepage(location, func(cp int) {
			if cp < 0 {
				return
			}
			prefs.Default.SetString("Workspace", location, asn1Now()+"&cp="+strconv.Itoa(cp))
			open(cp)
		})
	})
}

// openClient opens a workspace folder or an archive (ClientNode::open).
func (a *App) openClient(path string, cp int, workspace bool) {
	c := a.client
	c.format = cp
	if strings.HasSuffix(strings.ToLower(path), ".saf") {
		p, err := content.NewArchiveProvider(path, cp)
		if err != nil {
			a.status = "The archive can not be opened: " + err.Error()
			return
		}
		a.startTask("Loading Archive", int64(p.Archive.FileCount), func(t *task) error {
			return p.Load(t)
		}, func(err error) {
			if err != nil {
				p.Archive.Close()
				if !errors.Is(err, archive.ErrCancelled) {
					a.status = "The archive can not be read: " + err.Error()
				}
				return
			}
			c.provider = p
			c.root = p.Root()
			content.Current = p
			a.refreshClient()
		})
		return
	}
	var root *archive.Root
	a.startTask("Opening Workspace", 0, func(t *task) error {
		var err error
		root, err = archive.ScanDir(path, cp, t)
		return err
	}, func(err error) {
		if err != nil {
			if !errors.Is(err, archive.ErrCancelled) {
				a.status = "The workspace can not be read: " + err.Error()
			}
			return
		}
		p := content.NewFolderProvider(root, path, cp)
		c.provider = p
		c.root = root
		content.Current = p
		a.refreshClient()
		a.canPack = workspace
	})
}

// refreshClient rebuilds the tree under the Client node
// (ClientNode::refresh).
func (a *App) refreshClient() {
	c := a.client
	c.name = "Client - " + content.CurrentDirectory()
	a.tree.build(c)
	if a.afterOpen != nil {
		f := a.afterOpen
		a.afterOpen = nil
		f()
	}
}

// cmdNewWorkspace handles File > New Workspace (0xE100): expand an archive
// into a folder (ClientNode::createWorkspace).
func (a *App) cmdNewWorkspace() {
	if !a.closeAll() {
		return
	}
	file := prefs.Default.String("Workspace", "file", "")
	a.pickFile(filedlg.New("Select a Shaiya data file", filedlg.OpenFile, file, safFilter), func(file string) {
		prefs.Default.SetString("Workspace", "file", file)
		ws := prefs.Default.String("Workspace", "last", "")
		a.pickFile(filedlg.New("Select or create a workspace", filedlg.SelectFolder, ws), func(ws string) {
			prefs.Default.SetString("Workspace", "last", ws)
			st, err := os.Stat(file)
			if err != nil {
				a.alert("An error has occured.\nThe archive can not be opened.", iconError, nil)
				return
			}
			var extracted int
			a.startTask("Creating Workspace", st.Size()>>16, func(t *task) error {
				var err error
				extracted, err = archive.ExtractCount(file, ws, t)
				return err
			}, func(err error) {
				if errors.Is(err, archive.ErrCancelled) {
					return
				}
				if errors.Is(err, archive.ErrBadIndex) {
					a.alert("The .sah indexes file is invalid.", iconError, nil)
					return
				}
				if err != nil {
					a.alert("One or more errors has occurred.\n"+err.Error(), iconError, nil)
					return
				}
				a.alert("Archive successfully expanded.\n"+strconv.Itoa(extracted)+" files extracted.", iconInfo, func() {
					prefs.Default.SetString("Workspace", ws, asn1Now())
					a.chooseCodepage(ws+string(os.PathSeparator), func(cp int) {
						a.openClient(ws, cp, true)
					})
				})
			})
		})
	})
}

// cmdSaveAsArchive handles File > Save as Archive (0xE105): pack the
// workspace into a new data.saf.
func (a *App) cmdSaveAsArchive() {
	p := content.Current
	if p == nil || p.IsArchive() {
		return
	}
	ws := p.Directory()
	target := filepath.Join(filepath.Dir(filepath.Clean(ws)), "data.saf")
	d := filedlg.New("Save the workspace as", filedlg.SaveFile, target, safFilter)
	a.pickFile(d, func(target string) {
		if !strings.HasSuffix(strings.ToLower(target), ".saf") {
			target += ".saf"
		}
		if strings.HasPrefix(target, withSep(ws)) {
			a.alert("The archive file can not be saved under its own hierarchy.\nChoose a parent folder instead.", iconError, nil)
			return
		}
		root := p.Root()
		a.startTask("Creating Archive", root.TotalSize>>16, func(t *task) error {
			return archive.Build(target, root, p.CodePage(), t)
		}, func(err error) {
			if err != nil && !errors.Is(err, archive.ErrCancelled) {
				a.alert("The archive file can not be created.\nCheck path & access right.\n"+err.Error(), iconError, nil)
			}
		})
	})
}

func withSep(p string) string {
	if !strings.HasSuffix(p, string(os.PathSeparator)) {
		p += string(os.PathSeparator)
	}
	return p
}

// closeClient unloads the workspace (ClientNode::unload).
func (a *App) closeClient() {
	c := a.client
	c.name = "Client"
	if ap, ok := c.provider.(*content.ArchiveProvider); ok {
		ap.Archive.Close()
	}
	c.provider = nil
	c.root = &archive.Root{}
	c.data = clientData{}
	content.Current = nil
	a.releaseIcons()
	a.tree.clear()
	a.canPack = false
}
