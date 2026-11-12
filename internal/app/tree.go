package app

import (
	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
)

// shortcut groups shown under the Client node (ClientNode::refresh).
var (
	shortcuts = []string{
		`Cash.sdata`, `Character\Skill.sdata`, `Item\Item.sdata`, `Item\DualLayerClothes.sdata`,
		`Item\SetItem.sdata`, `Monster\Monster.mon`, `Monster\Monster.sdata`, `Monster\NpcSkill.sdata`,
		`Npc\npc.mon`, `Npc\NpcQuest.sdata`,
	}
	shortcuts6 = []string{
		`BinarySData\DBItemSellData.sdata`, `BinarySData\DBSkillData.sdata`, `BinarySData\DBItemData.SData`,
		`BinarySData\DBDualLayerClothesData.sdata`, `BinarySData\DBSetItemData.sdata`,
		`BinarySData\DBMonsterData.sdata`, `BinarySData\DBNpcSkillData.sdata`,
	}
)

type shortcutGroup struct {
	name  string
	open  bool // ToggleNode initial state
	files []*archive.File
}

// workspaceTree is the tree of the main window (StudioList / TreeList).
type workspaceTree struct {
	groups   []shortcutGroup
	allFiles *archive.Root
	selected any
}

func (t *workspaceTree) clear() {
	t.groups = nil
	t.allFiles = nil
	t.selected = nil
}

func (t *workspaceTree) build(c *client) {
	t.clear()
	root := c.root
	binary := root.ChildFolder("BinarySData") != nil
	name := "Shortcuts"
	if binary {
		name = "Shortcuts (5.x)"
	}
	g := shortcutGroup{name: name, open: true}
	for _, p := range shortcuts {
		if f := root.FindPath(p); f != nil {
			g.files = append(g.files, f)
		}
	}
	t.groups = append(t.groups, g)
	if binary {
		g6 := shortcutGroup{name: "Shortcuts (6.x)", open: true}
		for _, p := range shortcuts6 {
			if f := root.FindPath(p); f != nil {
				g6.files = append(g6.files, f)
			}
		}
		t.groups = append(t.groups, g6)
	}
	t.allFiles = root
}

func (a *App) drawTree() {
	t := &a.tree
	flags := imgui.TreeNodeFlagsOpenOnArrow | imgui.TreeNodeFlagsOpenOnDoubleClick | imgui.TreeNodeFlagsSpanAvailWidth
	// Client node.
	if !a.client.isClosed() {
		imgui.SetNextItemOpenV(true, imgui.CondOnce)
	}
	open := imgui.TreeNodeExStrV(a.client.name+"##client", flags|a.selFlag(a.client))
	a.treeItemEvents(a.client)
	if open {
		if !a.busy {
			for i := range t.groups {
				g := &t.groups[i]
				if g.open {
					imgui.SetNextItemOpenV(true, imgui.CondOnce)
				}
				if imgui.TreeNodeExStrV(g.name, flags) {
					for _, f := range g.files {
						a.drawFileItem(f, "s"+g.name)
					}
					imgui.TreePop()
				}
			}
			if t.allFiles != nil {
				if imgui.TreeNodeExStrV("All Files", flags) {
					a.drawFolder(&t.allFiles.Folder)
					imgui.TreePop()
				}
			}
		}
		imgui.TreePop()
	}
	// Server node (ServerNode).
	sflags := flags | a.selFlag(a.server)
	if a.server.root == nil {
		sflags |= imgui.TreeNodeFlagsLeaf
	}
	open = imgui.TreeNodeExStrV(a.server.title()+"##server", sflags)
	a.treeItemEvents(a.server)
	if open {
		a.drawServerChildren()
		imgui.TreePop()
	}
	// SQL Server node (SqlSvrNode).
	open = imgui.TreeNodeExStrV(a.sql.title()+"##sql", flags|imgui.TreeNodeFlagsLeaf|a.selFlag(a.sql))
	a.treeItemEvents(a.sql)
	if open {
		imgui.TreePop()
	}
}

func (a *App) selFlag(node any) imgui.TreeNodeFlags {
	if a.tree.selected == node {
		return imgui.TreeNodeFlagsSelected
	}
	return 0
}

func (a *App) drawFolder(d *archive.Folder) {
	flags := imgui.TreeNodeFlagsOpenOnArrow | imgui.TreeNodeFlagsOpenOnDoubleClick | imgui.TreeNodeFlagsSpanAvailWidth
	for _, s := range d.Folders {
		imgui.PushIDStr(s.Name)
		open := imgui.TreeNodeExStrV(s.Name, flags|a.selFlag(s))
		a.treeItemEvents(s)
		if open {
			a.drawFolder(s)
			imgui.TreePop()
		}
		imgui.PopID()
	}
	for i, f := range d.Files {
		imgui.PushIDInt(int32(i))
		a.drawFileItem(f, "")
		imgui.PopID()
	}
}

func (a *App) drawFileItem(f *archive.File, id string) {
	label := f.Name
	if s := f.StateText(); s != "" {
		label += "  (" + s + ")"
	}
	flags := imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsNoTreePushOnOpen | imgui.TreeNodeFlagsSpanAvailWidth | a.selFlag(f)
	imgui.TreeNodeExStrV(label+"##"+id+f.Path, flags)
	a.treeItemEvents(f)
}

// treeItemEvents handles selection, activation (double click / Enter) and
// the context menu of the last tree item.
func (a *App) treeItemEvents(node any) {
	if imgui.IsItemClicked() || imgui.IsItemClickedV(imgui.MouseButtonRight) {
		a.tree.selected = node
	}
	if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
		if f, ok := node.(*archive.File); ok {
			a.reopenDocument(f)
		}
	}
	if imgui.IsItemFocused() && imgui.IsKeyPressedBool(imgui.KeyEnter) {
		if f, ok := node.(*archive.File); ok {
			a.reopenDocument(f)
		}
	}
	if imgui.BeginPopupContextItem() {
		a.drawNodeMenu(node)
		imgui.EndPopup()
	}
}

// drawNodeMenu is the popup menu of a tree node (menus 132, 133, 134).
func (a *App) drawNodeMenu(node any) {
	switch n := node.(type) {
	case *client:
		// Menu 132 "shaiya client items".
		if imgui.MenuItemBool("Open Workspace…") {
			a.cmdOpen(false)
		}
		if imgui.MenuItemBool("Open Archive…") {
			a.cmdOpen(true)
		}
		imgui.Separator()
		if imgui.MenuItemBoolV("Close Workspace", "", false, !a.client.isClosed()) {
			a.closeAll()
		}
		if imgui.MenuItemBoolV("Save as Archive …", "", false, a.canSaveAsArchive()) {
			a.cmdSaveAsArchive()
		}
		imgui.MenuItemBoolV("Generate Patch …", "", false, false)
	case *serverNode:
		n.drawMenu(a)
	case *sqlNode:
		n.drawMenu(a)
	case *archive.File:
		if imgui.MenuItemBool("Open") {
			a.reopenDocument(n)
		}
	}
}

func (a *App) canSaveAsArchive() bool {
	p := content.Current
	return p != nil && !p.IsArchive()
}
