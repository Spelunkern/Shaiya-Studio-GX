package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/codepage"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/table"
)

// npcSkillView shows NpcSkill.SData / DBNpcSkillData (NpcSkillView with
// its NpcSkillTab: "Mobs Skills" and "Raw Skills").
type npcSkillView struct {
	app    *App
	doc    *sdataDoc
	skills *sdata.MobsSkills
	set    []*sdata.MobsSkill // skills with a type (MobsSkillSet)
	lists  [2]*table.List
	tab    int
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		name := strings.ToLower(f.Name)
		if name != "npcskill.sdata" && name != "dbnpcskilldata.sdata" {
			return nil
		}
		s := sdata.NewMobsSkills()
		v := &npcSkillView{app: a, skills: s}
		d := &sdataDoc{file: f, data: s, view: v, hasTranslation: true, detectCP: true, decrypt: true}
		d.afterSave = v.writeScript
		v.doc = d
		return d
	})
}

func (v *npcSkillView) skillList(id string, cols []table.Column) *table.List {
	d := v.doc
	l := table.New(id, cols...)
	l.SortColumn = "skid"
	l.Rows = func() int { return len(v.set) }
	l.Row = func(i int) table.Node { return v.set[i] }
	l.Sort = func(k1, k2 sdata.Tag, asc bool) bool { return sdata.SortList(v.set, k1, k2, asc) }
	l.Insert = func(index int) int {
		s := sdata.NewMobsSkill(v.skills.Format, int32(int16(len(v.skills.Skills)+1)))
		v.skills.Skills = append(v.skills.Skills, s)
		v.set = append(v.set, s)
		return len(v.set) - 1
	}
	l.Remove = func(index int) bool {
		s := v.set[index]
		v.set = append(v.set[:index], v.set[index+1:]...)
		for i, x := range v.skills.Skills {
			if x == s {
				v.skills.Skills = append(v.skills.Skills[:i], v.skills.Skills[i+1:]...)
				break
			}
		}
		return true
	}
	l.Clear = func(index int) bool {
		v.set[index].Reset()
		return true
	}
	l.Activate = func(index int) { d.status = "The NPC skill editor is not ported yet." }
	l.Copy = func(rows []int) string {
		var b strings.Builder
		for _, r := range rows {
			b.WriteString(v.set[r].Line(sdata.TextCSV))
		}
		return b.String()
	}
	l.Changed = func() { d.setModified(true) }
	l.Confirm = func(text string, then func(bool)) {
		if strings.Contains(text, "to delete") {
			text = "Are you sure you want to delete this skill?\nThe indexes of all subsequent skills will be changed."
		}
		v.app.confirm(text, then)
	}
	l.ContextMenu = func() {
		// Menu 190 "Show Skill Options".
		imgui.MenuItemBoolV("Show All Skills", "", true, true)
		imgui.Separator()
		l.EditMenu()
	}
	return l
}

func (v *npcSkillView) loaded(d *sdataDoc) {
	v.set = v.skills.Visible()
	clr := []table.Column{
		table.C("ID", 50, 16, "skid"), table.C("Name", 120, 4096, "name"), table.C("State", 50, 18, "sttp"),
		table.C("Type", 130, 0, "tpdt"), table.C("Dmg Type", 50, 18, "dmtp"), table.C("Desc", 250, 4096, "desc"),
	}
	raw := []table.Column{table.C("ID", 50, 16, "skid"), table.C("Name", 120, 4096, "name"),
		table.C("Type", 50, 18, "tpdt"), table.C("Icon", 50, 18, "icon")}
	add := func(from, n int) {
		for i := 0; i < n; i++ {
			raw = append(raw, table.C(sdata.SkillColumnNames[from+i], 50, 18, sdata.SkillColumnTags[from+i]))
		}
	}
	add(19, 2)
	add(40, 5)
	add(46, 6)
	add(57, 11)
	add(88, 6)
	add(68, 6)
	if v.skills.Format&0xFFF >= 0x600 {
		add(74, 14)
	}
	v.lists[0] = v.skillList("mobs", clr)
	v.lists[1] = v.skillList("raw", raw)
}

func (v *npcSkillView) draw(d *sdataDoc) {
	if imgui.BeginTabBar("npcskilltabs") {
		for i, title := range []string{"Mobs Skills", "Raw Skills"} {
			if imgui.BeginTabItem(title) {
				v.tab = i
				v.lists[i].Draw()
				imgui.EndTabItem()
			}
		}
		imgui.EndTabBar()
	}
}

func (v *npcSkillView) editMenu(d *sdataDoc) { v.lists[v.tab].EditMenu() }

// toolsMenu is the Format menu (NpcSkillView::updateMenu).
func (v *npcSkillView) toolsMenu(d *sdataDoc) {
	if imgui.BeginMenu("Format") {
		mode := v.skills.Format & 0xFFF
		for _, f := range []struct {
			text string
			mode int
		}{
			{"Episo 4 format (3 levels per skill, 3 abilities per level)", 0x400},
			{"Episo 5 format (9 levels per skill, 3 abilities per level)", 0x500},
			{"Episo 6 format (15 levels per skill, 10 abilities per level)", 0x600},
		} {
			if imgui.MenuItemBoolV(f.text, "", mode == f.mode, true) {
				v.skills.SetVersion(f.mode)
				d.setModified(true)
			}
		}
		imgui.EndMenu()
	}
}

func (v *npcSkillView) export(format string) (string, bool) { return v.skills.Text(format), true }

// writeScript writes the SQL script of the table next to the server files
// after a save (NpcSkillView::save).
func (v *npcSkillView) writeScript(d *sdataDoc) {
	path, ok := v.app.scriptPath(d.file.Name)
	if !ok {
		return
	}
	text := v.skills.Text(sdata.TextSQL)
	os.WriteFile(path, codepage.Encode(text, d.file.CodePage), 0o644)
	d.status = "File saved and SQL script generated."
}

// scriptPath returns <server>\scripts\<name>.sql, or <server>\<name>.sql
// when there is no scripts folder (SDataView::getScriptPath).
func (a *App) scriptPath(name string) (string, bool) {
	dir := a.serverPath()
	if dir == "" {
		return "", false
	}
	if st, err := os.Stat(filepath.Join(dir, "scripts")); err == nil && st.IsDir() {
		dir = filepath.Join(dir, "scripts")
	}
	return filepath.Join(dir, strings.TrimSuffix(name, filepath.Ext(name))+".sql"), true
}

// serverPath returns the PSM_Client folder of the server repository: two
// levels above its Bin\Data folder (StudioFrame::getServerPath), or "".
func (a *App) serverPath() string {
	s := a.server
	if s == nil || !s.connected || s.path == "" {
		return ""
	}
	return filepath.Clean(s.path)
}
