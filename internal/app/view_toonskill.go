package app

import (
	"os"
	"strings"
	"unicode/utf16"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/table"
)

// toonSkillView shows Skill.SData / DBSkillData (ToonSkillView with its
// ToonSkillTab: "Toon Skills", "Noos Skills" and "Raw Skills").
type toonSkillView struct {
	app    *App
	doc    *sdataDoc
	skills *sdata.ToonSkills
	sets   [3][]*sdata.ToonSkill
	lists  [3]*table.List
	class  int // class filter of the Toon Skills list (0 = all)
	tab    int
}

var toonSkillModes = [3]int{1, 2, 0}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		name := strings.ToLower(f.Name)
		if name != "skill.sdata" && name != "dbskilldata.sdata" {
			return nil
		}
		s := sdata.NewToonSkills()
		v := &toonSkillView{app: a, skills: s}
		d := &sdataDoc{file: f, data: s, view: v, hasTranslation: true, detectCP: true, decrypt: true}
		d.afterSave = v.writeScript
		v.doc = d
		return d
	})
}

func (v *toonSkillView) fill(i int) {
	v.sets[i] = v.skills.Filter(toonSkillModes[i], v.class)
	if v.lists[i] != nil {
		v.lists[i].Reset()
	}
}

func (v *toonSkillView) newList(i int, id string, cols []table.Column) *table.List {
	d := v.doc
	l := table.New(id, cols...)
	l.SortColumn, l.PrevSortColumn = "skid", "name"
	l.Rows = func() int { return len(v.sets[i]) }
	l.Row = func(r int) table.Node { return v.sets[i][r] }
	l.Sort = func(k1, k2 sdata.Tag, asc bool) bool { return sdata.SortList(v.sets[i], k1, k2, asc) }
	l.Insert = func(index int) int {
		s := sdata.NewToonSkill(int32(len(v.skills.Skills)), v.skills.Format)
		v.skills.Skills = append(v.skills.Skills, s)
		switch i {
		case 0:
			s.Levels[0].TypeShow = 1
		case 1:
			s.Levels[0].TypeShow = 0
			s.Levels[0].TypeDetail = 82
		}
		v.sets[i] = append(v.sets[i], s)
		if i != 2 {
			v.sets[2] = append(v.sets[2], s)
			v.lists[2].Resort()
		}
		return len(v.sets[i]) - 1
	}
	l.Remove = func(index int) bool {
		// Only the last skill can be deleted (ToonSkillList::removeItem).
		if index+1 != len(v.sets[i]) {
			return false
		}
		s := v.sets[i][index]
		for k := range v.sets {
			for j, x := range v.sets[k] {
				if x == s {
					v.sets[k] = append(v.sets[k][:j], v.sets[k][j+1:]...)
					break
				}
			}
		}
		for j, x := range v.skills.Skills {
			if x == s {
				v.skills.Skills = append(v.skills.Skills[:j], v.skills.Skills[j+1:]...)
				break
			}
		}
		return true
	}
	l.Clear = func(index int) bool {
		v.sets[i][index].Reset()
		return true
	}
	l.Activate = func(index int) { d.status = "The skill editor is not ported yet." }
	l.Copy = func(rows []int) string {
		var b strings.Builder
		for _, r := range rows {
			b.WriteString(v.sets[i][r].Line(sdata.TextCSV))
		}
		return b.String()
	}
	l.Changed = func() { d.setModified(true) }
	l.Confirm = func(text string, then func(bool)) {
		switch {
		case strings.Contains(text, "to delete"):
			text = "Are you sure you want to delete this skill?"
		case strings.Contains(text, "to clear"), strings.Contains(text, "erase"):
			text = "Are you sure you want to erase this skill?"
		}
		v.app.confirm(text, then)
	}
	if i == 0 {
		l.ContextMenu = func() {
			// Menu 192 "Show Skill Options".
			names := []string{"Show All Skills", "Show Fighter / Warrior Skills", "Show Defender / Guardian Skills",
				"Show Ranger / Assassin Skills", "Show Archer / Hunter Skills", "Show Mage / Pagan Skills",
				"Show Priest / Oracle Skills"}
			for c, name := range names {
				if imgui.MenuItemBoolV(name, "", v.class == c, true) {
					v.class = c
					v.fill(0)
				}
			}
			imgui.Separator()
			l.EditMenu()
		}
	}
	return l
}

func (v *toonSkillView) loaded(d *sdataDoc) {
	clr := []table.Column{
		table.C("ID", 50, 16, "skid"), table.C("Name", 120, 4096, "name"), table.C("State", 50, 18, "sttp"),
		table.C("Type", 130, 0, "tpdt"), table.C("Dmg Type", 50, 18, "dmtp"), table.C("Desc", 250, 4096, "desc"),
	}
	// RawSkillList::onCreate.
	raw := []table.Column{table.C("ID", 50, 16, "skid"), table.C("Name", 120, 4096, "name"),
		table.C("Type", 50, 18, "tpdt"), table.C("Icon", 50, 18, "icon"), table.C("Anim", 50, 18, "anim")}
	for k := 0; k < 60; k++ {
		tag, name := sdata.SkillColumnTags[8+k], sdata.SkillColumnNames[8+k]
		if tag == "tpdt" {
			continue
		}
		if tag == "" || name == "" {
			break
		}
		raw = append(raw, table.C(name, 50, 18, tag))
	}
	add := func(from, n int) {
		for k := 0; k < n; k++ {
			raw = append(raw, table.C(sdata.SkillColumnNames[from+k], 50, 18, sdata.SkillColumnTags[from+k]))
		}
	}
	add(88, 6)
	add(68, 6)
	if v.skills.Format&0xFFF >= 0x600 {
		add(74, 14)
	}
	add(94, 6)
	v.lists[0] = v.newList(0, "pers", clr)
	v.lists[1] = v.newList(1, "noos", clr)
	v.lists[2] = v.newList(2, "raws", raw)
	for i := 0; i < 2; i++ {
		v.lists[i].Icon = func(n table.Node, size float32) {
			v.app.drawIcon(v.app.iconLoader().Strip("icon_skills"), n.(*sdata.ToonSkill).IconID(), size)
		}
	}
	for i := range v.sets {
		v.fill(i)
	}
}

func (v *toonSkillView) draw(d *sdataDoc) {
	if imgui.BeginTabBar("toonskilltabs") {
		for i, title := range []string{"Toon Skills", "Noos Skills", "Raw Skills"} {
			if imgui.BeginTabItem(title) {
				v.tab = i
				v.lists[i].Draw()
				imgui.EndTabItem()
			}
		}
		imgui.EndTabBar()
	}
}

func (v *toonSkillView) editMenu(d *sdataDoc) { v.lists[v.tab].EditMenu() }

func (v *toonSkillView) toolsMenu(d *sdataDoc) {
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

func (v *toonSkillView) export(format string) (string, bool) { return v.skills.Text(format), true }

// writeScript writes the SQL script as UTF-16 after a save
// (ToonSkillView::save).
func (v *toonSkillView) writeScript(d *sdataDoc) {
	path, ok := v.app.scriptPath(d.file.Name)
	if !ok {
		return
	}
	text := v.skills.Text(sdata.TextSQL)
	var out []byte
	if text != "" {
		u := utf16.Encode([]rune(text))
		out = append(out, 0xFF, 0xFE)
		for _, c := range u {
			out = append(out, byte(c), byte(c>>8))
		}
	}
	os.WriteFile(path, out, 0o644)
	d.status = "File saved and SQL script generated."
}
