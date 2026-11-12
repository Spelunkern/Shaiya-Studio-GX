package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/filedlg"
)

// serverNode is the "Server" node: a server files repository
// (PSM_Client\Bin\Data) shown as a folder tree (ServerNode).
type serverNode struct {
	name      string
	path      string // PSM_Client folder, with a trailing separator
	root      *archive.Root
	connected bool
}

func newServerNode() *serverNode { return &serverNode{name: "Server"} }

func (s *serverNode) title() string {
	if s.connected {
		return s.name + "    Connected"
	}
	return s.name
}

// isClientFolder accepts a PSM_Client folder (FolderOption::isClientFolder).
func isClientFolder(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "Bin", "Data"))
	return err == nil && st.IsDir()
}

func (s *serverNode) drawMenu(a *App) {
	// Menu 133 "shaiya server items".
	if imgui.MenuItemBool("Choose repository…") {
		location := prefs.Default.String("Server Data", "location", "")
		d := filedlg.New("Select the server files repository", filedlg.SelectFolder, location)
		d.Accept = isClientFolder
		a.pickFile(d, func(location string) {
			sep := string(os.PathSeparator)
			if !strings.HasSuffix(strings.ToLower(location), strings.ToLower("PSM_Client"+sep)) &&
				!strings.HasSuffix(strings.ToLower(location), "psm_client") {
				location = filepath.Join(location, "PSM_Client")
			}
			location = withSep(location)
			s.connect(a, location)
			prefs.Default.SetString("Server Data", "location", location)
			prefs.Default.SetInt("Server Data", "Connect", 1)
		})
	}
	if imgui.MenuItemBoolV("Unlink repository", "", false, s.path != "") {
		s.name = "Server"
		s.root = nil
		s.path = ""
		s.connected = false
		prefs.Default.SetInt("Server Data", "Connect", 0)
	}
}

// connect scans path\Bin\Data (ServerNode::connect / populate).
func (s *serverNode) connect(a *App, path string) {
	s.connected = false
	s.path = path
	root, err := archive.ScanDir(filepath.Join(path, "Bin", "Data"), 0, nil)
	if err != nil {
		a.status = "The selected folder is invalid."
		return
	}
	s.root = root
	s.name = "Server - " + path
	s.connected = true
}

func (a *App) drawServerChildren() {
	s := a.server
	if s.root == nil {
		return
	}
	flags := imgui.TreeNodeFlagsOpenOnArrow | imgui.TreeNodeFlagsOpenOnDoubleClick | imgui.TreeNodeFlagsSpanAvailWidth
	if imgui.TreeNodeExStrV(`PSM_Client\Bin\Data##srv`, flags) {
		a.drawFolder(&s.root.Folder)
		imgui.TreePop()
	}
}

// sqlNode is the "Databases" node: a SQL Server connection (SqlSvrNode).
// The SQL Server module is not ported yet.
type sqlNode struct {
	connecting, connected bool
}

func newSQLNode() *sqlNode { return &sqlNode{} }

func (s *sqlNode) title() string {
	state := "Offline"
	switch {
	case s.connecting:
		state = "Connecting..."
	case s.connected:
		state = "Connected"
	}
	return "Databases    " + state
}

func (s *sqlNode) drawMenu(a *App) {
	// Menu 134 "SQL Server items".
	if imgui.MenuItemBool("Connect …") {
		a.status = "SQL Server: not ported yet."
	}
	if imgui.MenuItemBool("Disconnect") {
		s.connected = false
	}
}
