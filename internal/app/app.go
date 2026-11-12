// Package app is the main window of Shaiya Studio GX (StudioFrame,
// Application and StudioList in shStudio): the menu, the workspace tree,
// the status bar and the document windows.
package app

import (
	"os"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	_ "github.com/AllenDang/cimgui-go/impl/glfw"
	_ "github.com/AllenDang/cimgui-go/impl/opengl3"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
)

// AppName is the window title (g_appName is "shStudio 10" in the original).
const AppName = "Shaiya Studio GX"

// App is the running program.
type App struct {
	be      backend.Backend[glfwbackend.GLFWWindowFlags]
	client  *client
	server  *serverNode
	sql     *sqlNode
	tree    workspaceTree
	docs    []*openDoc
	modals  []func() bool
	status  string
	busy    bool
	canPack bool
	idSeq   int
	quit    bool
	dark    bool

	afterOpen func() // debug hook run once the client is open
}

var current *App

// Run opens the main window and runs until it is closed.
func Run() {
	a := &App{client: newClient(), server: newServerNode(), sql: newSQLNode()}
	current = a
	be, _ := backend.CreateBackend(glfwbackend.NewGLFWBackend())
	a.be = be
	p := prefs.Default
	w := p.Int(prefs.AppSection, "width", 480)
	h := p.Int(prefs.AppSection, "height", 300)
	be.SetAfterCreateContextHook(a.afterCreate)
	be.SetBgColor(imgui.NewVec4(0.94, 0.94, 0.94, 1))
	be.CreateWindow(AppName, w, h)
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		// Wayland does not let programs place their windows.
		be.SetWindowPos(p.Int(prefs.AppSection, "left", 50), p.Int(prefs.AppSection, "top", 50))
	}
	be.SetCloseCallback(func() { a.savePlacement() })
	be.SetDropCallback(func(paths []string) {
		for _, path := range paths {
			a.openFile(path)
		}
	})
	be.SetAfterRenderHook(a.afterRender)
	a.onStartup()
	a.debugOpen()
	be.Run(a.frame)
}

func (a *App) afterCreate() {
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	a.dark = darkModeEnabled()
	applyTheme(a.dark)
	if a.dark {
		a.be.SetBgColor(imgui.NewVec4(0.12, 0.12, 0.12, 1))
	}
}

// onStartup runs the startup work (StudioFrame::onStartup): reconnect the
// server repository when asked to.
func (a *App) onStartup() {
	p := prefs.Default
	if p.Int("Server Data", "Connect", 0) == 1 {
		if location := p.String("Server Data", "location", ""); location != "" {
			a.server.connect(a, location)
		}
	}
}

func (a *App) savePlacement() {
	p := prefs.Default
	x, y := a.be.GetWindowPos()
	w, h := a.be.DisplaySize()
	if x > 0 && y > 0 {
		p.SetInt(prefs.AppSection, "left", int(x))
		p.SetInt(prefs.AppSection, "top", int(y))
		p.SetInt(prefs.AppSection, "width", int(w))
		p.SetInt(prefs.AppSection, "height", int(h))
	}
}

func (a *App) wake() { a.be.Refresh() }

// frame draws one frame.
func (a *App) frame() {
	if a.quit {
		a.savePlacement()
		a.be.SetShouldClose(true)
	}
	a.shortcuts()
	vp := imgui.MainViewport()
	a.drawMainMenu()
	statusH := imgui.FrameHeight()
	pos := vp.WorkPos()
	size := vp.WorkSize()
	imgui.SetNextWindowPos(pos)
	imgui.SetNextWindowSize(imgui.NewVec2(size.X, size.Y-statusH))
	imgui.SetNextWindowViewport(vp.ID())
	flags := imgui.WindowFlagsNoDecoration | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoSavedSettings |
		imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoDocking
	if imgui.BeginV("##workspace", nil, flags) {
		a.drawTree()
	}
	imgui.End()
	// Status bar.
	imgui.SetNextWindowPos(imgui.NewVec2(pos.X, pos.Y+size.Y-statusH))
	imgui.SetNextWindowSize(imgui.NewVec2(size.X, statusH))
	imgui.SetNextWindowViewport(vp.ID())
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.NewVec2(6, 2))
	if imgui.BeginV("##status", nil, flags|imgui.WindowFlagsNoScrollbar) {
		imgui.TextUnformatted(a.status)
	}
	imgui.End()
	imgui.PopStyleVar()
	a.drawDocs()
	a.drawModals()
}

// shortcuts handles the accelerators of the main window (ACCELERATORS 128).
func (a *App) shortcuts() {
	if len(a.modals) > 0 {
		return
	}
	ctrl := imgui.CurrentIO().KeyCtrl()
	shift := imgui.CurrentIO().KeyShift()
	pressed := func(k imgui.Key) bool { return imgui.IsKeyPressedBool(k) }
	switch {
	case ctrl && pressed(imgui.KeyN):
		a.cmdNewWorkspace()
	case ctrl && shift && pressed(imgui.KeyO):
		a.cmdOpen(true)
	case ctrl && pressed(imgui.KeyO):
		a.cmdOpen(false)
	case ctrl && pressed(imgui.KeyQ):
		a.quit = true
	}
}

func (a *App) drawMainMenu() {
	if !imgui.BeginMainMenuBar() {
		return
	}
	// MENU 128.
	if imgui.BeginMenu("File") {
		if imgui.MenuItemBoolV("New Workspace…", "Ctrl+N", false, true) {
			a.cmdNewWorkspace()
		}
		if imgui.MenuItemBoolV("Open Workspace…", "Ctrl+O", false, true) {
			a.cmdOpen(false)
		}
		if imgui.MenuItemBoolV("Open Archive…", "Shift+Ctrl+O", false, true) {
			a.cmdOpen(true)
		}
		imgui.Separator()
		if imgui.MenuItemBoolV("Close Workspace", "Ctrl+W", false, !a.client.isClosed()) {
			a.closeAll()
		}
		if imgui.MenuItemBoolV("Save as Archive …", "Ctrl+S", false, a.canPack && a.canSaveAsArchive()) {
			a.cmdSaveAsArchive()
		}
		imgui.MenuItemBoolV("Generate Patch …", "Ctrl+P", false, false)
		imgui.Separator()
		if imgui.MenuItemBoolV("Exit", "Ctrl+Q", false, true) {
			a.quit = true
		}
		imgui.EndMenu()
	}
	if imgui.BeginMenu("Edit") {
		imgui.MenuItemBoolV("Undo", "Ctrl+Z", false, false)
		imgui.MenuItemBoolV("Redo", "Ctrl+Y", false, false)
		imgui.Separator()
		imgui.MenuItemBoolV("Cut", "Ctrl+X", false, false)
		if imgui.MenuItemBoolV("Copy", "Ctrl+C", false, true) {
			a.copySelection()
		}
		imgui.MenuItemBoolV("Paste", "Ctrl+V", false, false)
		imgui.EndMenu()
	}
	if imgui.BeginMenu("Views") {
		if imgui.MenuItemBool("Settings") {
			a.status = "Settings: not ported yet."
		}
		if imgui.MenuItemBoolV("Dark mode", "", a.dark, true) {
			a.dark = !a.dark
			setDarkMode(a.dark)
			applyTheme(a.dark)
		}
		imgui.Separator()
		if imgui.BeginMenu("Scripts") {
			imgui.MenuItemBoolV("no script", "", false, false)
			imgui.EndMenu()
		}
		imgui.Separator()
		if len(a.docs) == 0 {
			imgui.MenuItemBoolV("no data view", "", false, false)
		}
		for _, d := range a.docs {
			if imgui.MenuItemBool(d.doc.Title()) {
				d.focus = true
			}
		}
		imgui.EndMenu()
	}
	if imgui.BeginMenu("?") {
		if imgui.MenuItemBool("About") {
			a.about()
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
}

// copySelection copies the name of the selected tree item.
func (a *App) copySelection() {
	switch n := a.tree.selected.(type) {
	case interface{ CopyText() string }:
		imgui.SetClipboardText(n.CopyText())
	default:
		_ = n
	}
}

func (a *App) about() {
	a.alert(AppName+"\n\nA Go port of shStudio 10 by ilya (MIT license).\nhttps://github.com/Spelunkern/Shaiya-Studio-GX", iconInfo, nil)
}
