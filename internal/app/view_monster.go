package app

import (
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/table"
)

// monsterView shows Monster.SData / DBMonsterData (MonsterView +
// MonsterList).
type monsterView struct {
	app      *App
	doc      *sdataDoc
	monsters *sdata.Monsters
	items    []*sdata.Monster
	list     *table.List
	showAll  bool
	groups   int // column groups: 1 stats, 2 resistances, 4 attacks
	clip     []*sdata.Monster
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		name := strings.ToLower(f.Name)
		if name != "monster.sdata" && name != "dbmonsterdata.sdata" && name != "monsterex.sdata" {
			return nil
		}
		dbFiles := strings.EqualFold(prefs.Default.String("SData", "Primary", "undefined"), "DBfiles")
		_ = dbFiles
		m := sdata.NewMonsters()
		if name == "monsterex.sdata" {
			m.Extended = true
		}
		v := &monsterView{app: a, monsters: m}
		d := &sdataDoc{file: f, data: m, view: v, hasTranslation: true, detectCP: true, decrypt: true}
		v.doc = d
		return d
	})
}

var monsterBaseColumns = []table.Column{
	table.C("MobID", 60, 16, "mbid"),
	table.C("Name", 200, 4096, "name"),
	table.C("Level", 45, 18, "lvl "),
	table.C("Model", 45, 18, "modl"),
	table.C("Exp", 45, 18, "exp_"),
	table.C("AI", 45, 18, "ai  "),
}

func (v *monsterView) columns() []table.Column {
	cols := append([]table.Column(nil), monsterBaseColumns...)
	if v.groups&1 != 0 {
		cols = append(cols, table.C("Money1", 50, 18, "gld1"), table.C("Money2", 50, 18, "gld2"),
			table.C("Qst Itm", 50, 18, "qust"), table.C("HP", 45, 18, "hp  "), table.C("SP", 45, 18, "sp  "),
			table.C("MP", 45, 18, "mp  "), table.C("Dex", 45, 18, "dex "), table.C("Wis", 45, 18, "wis "),
			table.C("Luc", 45, 18, "luc "), table.C("Height", 45, 18, "day "), table.C("Size", 45, 18, "size"),
			table.C("Elem", 45, 2, "attr"), table.C("Defense", 45, 18, "defs"), table.C("Magic", 45, 18, "magc"),
			table.C("Normal Time", 45, 18, "ntme"), table.C("Normal Step", 45, 18, "nstp"),
			table.C("Chase Time", 45, 18, "ctme"), table.C("Chase Step", 45, 18, "cstp"),
			table.C("Chase Range", 45, 18, "crng"))
	}
	if v.groups&2 != 0 {
		for i := 1; i <= 15; i++ {
			cols = append(cols, table.C("Res.State "+itoa(i), 45, 18, sdata.Tag("rs"+pad2(i))))
		}
		for i := 1; i <= 6; i++ {
			cols = append(cols, table.C("Res.Skill "+itoa(i), 45, 18, sdata.Tag("rk"+pad2(i))))
		}
	}
	if v.groups&4 != 0 {
		for n := 1; n <= 3; n++ {
			p := itoa(n)
			cols = append(cols, table.C("Atk "+p+" Type", 45, 18, sdata.Tag(p+"typ")),
				table.C("Atk "+p+" Time", 45, 18, sdata.Tag(p+"tme")), table.C("Atk "+p+" Range", 45, 18, sdata.Tag(p+"rng")),
				table.C("Atk "+p+" Dmg", 45, 18, sdata.Tag(p+"dmg")), table.C("Atk "+p+" Plus", 45, 18, sdata.Tag(p+"add")),
				table.C("Atk "+p+" Elem", 45, 18, sdata.Tag(p+"atr")), table.C("Atk "+p+" Spec", 45, 18, sdata.Tag(p+"spc")),
				table.C("Atk "+p+" Mode", 45, 18, sdata.Tag(p+"mod")))
		}
	}
	if v.groups == 0 {
		cols = append(cols, table.C("Height", 45, 18, "day "), table.C("Elem", 45, 2, "attr"),
			table.C("Qst Itm", 50, 18, "qust"), table.C("Atk Mode 1", 50, 18, "1mod"),
			table.C("Atk Mode 2", 50, 18, "2mod"), table.C("Atk Mode 3", 50, 18, "3mod"))
	}
	return cols
}

// refresh rebuilds the rows (MonsterList::refresh): all monsters, or the
// valid ones (level > 0, name not starting with "??" or "Error").
func (v *monsterView) refresh() {
	v.items = v.items[:0]
	for _, m := range v.monsters.Monsters {
		if v.showAll {
			v.items = append(v.items, m)
			continue
		}
		if m.Level > 0 && !strings.HasPrefix(m.Name, "??") && !strings.HasPrefix(m.Name, "Error") {
			v.items = append(v.items, m)
		}
	}
	v.list.Reset()
}

func (v *monsterView) loaded(d *sdataDoc) {
	if v.list == nil {
		l := table.New("monsters", v.columns()...)
		l.SortColumn = "mbid"
		l.Rows = func() int { return len(v.items) }
		l.Row = func(i int) table.Node { return v.items[i] }
		l.Sort = func(k1, k2 sdata.Tag, asc bool) bool { return sdata.SortList(v.items, k1, k2, asc) }
		l.Insert = func(index int) int {
			m := sdata.NewMonster(int32(len(v.monsters.Monsters)), v.monsters.Format)
			v.monsters.Monsters = append(v.monsters.Monsters, m)
			if index < 0 || index > len(v.items) {
				index = len(v.items)
			}
			v.items = append(v.items[:index], append([]*sdata.Monster{m}, v.items[index:]...)...)
			return index
		}
		l.Remove = func(index int) bool {
			m := v.items[index]
			for i, x := range v.monsters.Monsters {
				if x == m {
					v.monsters.Monsters = append(v.monsters.Monsters[:i], v.monsters.Monsters[i+1:]...)
					break
				}
			}
			v.items = append(v.items[:index], v.items[index+1:]...)
			return true
		}
		l.Clear = func(index int) bool {
			v.items[index].Reset()
			return true
		}
		l.Activate = func(index int) {
			d.status = "The monster editor is not ported yet."
		}
		l.Copy = func(rows []int) string {
			v.clip = nil
			var b strings.Builder
			for _, r := range rows {
				c := *v.items[r]
				v.clip = append(v.clip, &c)
				line, _ := v.items[r].Line(sdata.TextCSV)
				b.WriteString(line + "\r\n")
			}
			return b.String()
		}
		l.CanPaste = func() bool { return len(v.clip) > 0 }
		l.Paste = func(sel []int) bool {
			if len(v.clip) == 1 && len(sel) == 1 {
				v.items[sel[0]].Assign(v.clip[0])
				return true
			}
			changed := false
			for _, c := range v.clip {
				if m := v.monsters.Find(c.ID); m != nil {
					m.Assign(c)
				} else {
					n := *c
					v.monsters.Monsters = append(v.monsters.Monsters, &n)
					v.items = append(v.items, &n)
				}
				changed = true
			}
			return changed
		}
		l.Changed = func() { d.setModified(true) }
		l.Confirm = func(text string, then func(bool)) {
			v.app.confirm("Are you sure you want to delete the selected monster(s)?\nThe indexes of all subsequent monsters will be changed.", then)
		}
		l.ContextMenu = v.contextMenu
		v.list = l
	}
	v.refresh()
}

// contextMenu is the "Monsters Options" menu (194) plus the edit commands.
func (v *monsterView) contextMenu() {
	if imgui.MenuItemBoolV("Show valid mobs", "", !v.showAll, true) {
		v.showAll = false
		v.refresh()
	}
	if imgui.MenuItemBoolV("Show all mobs", "", v.showAll, true) {
		v.showAll = true
		v.refresh()
	}
	if imgui.BeginMenu("Show Columns") {
		if imgui.MenuItemBoolV("Show All Columns", "", v.groups == 7, true) {
			v.groups = 7
			v.list.Columns = v.columns()
		}
		for i, name := range []string{"Show Base Stats", "Show Resistances", "Show Attacks"} {
			if imgui.MenuItemBoolV(name, "", v.groups&(1<<i) != 0, true) {
				v.groups ^= 1 << i
				v.list.Columns = v.columns()
			}
		}
		imgui.EndMenu()
	}
	imgui.Separator()
	v.list.EditMenu()
}

func (v *monsterView) draw(d *sdataDoc) { v.list.Draw() }

func (v *monsterView) editMenu(d *sdataDoc) {
	v.list.EditMenu()
}

func (v *monsterView) toolsMenu(d *sdataDoc) {}

func (v *monsterView) export(format string) (string, bool) {
	if format == sdata.TextSQL {
		return v.monsters.Text("mbdp"), true
	}
	return v.monsters.Text(format), true
}
