package app

import (
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/table"
)

// itemSetView shows SetItem.SData / DBSetItemData (ItemSetView +
// ItemSetList).
type itemSetView struct {
	app  *App
	doc  *sdataDoc
	sets *sdata.ItemSets
	list *table.List
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		name := strings.ToLower(f.Name)
		if name != "setitem.sdata" && name != "dbsetitemdata.sdata" {
			return nil
		}
		s := &sdata.ItemSets{}
		v := &itemSetView{app: a, sets: s}
		d := &sdataDoc{file: f, data: s, view: v, hasTranslation: true, detectCP: true, decrypt: true}
		v.doc = d
		return d
	})
}

func (v *itemSetView) loaded(d *sdataDoc) {
	if v.list != nil {
		v.list.Reset()
		return
	}
	l := table.New("itemsets", table.C("Index", 50, 18, "indx"), table.C("Name", 300, 0, "name"))
	l.SortColumn = "indx"
	l.Rows = func() int { return len(v.sets.Sets) }
	l.Row = func(i int) table.Node { return v.sets.Sets[i] }
	l.Sort = func(k1, k2 sdata.Tag, asc bool) bool { return sdata.SortList(v.sets.Sets, k1, k2, asc) }
	l.Activate = func(i int) { v.app.editSet(v.sets.Sets[i], func() { d.setModified(true) }) }
	l.Copy = func(rows []int) string {
		var b strings.Builder
		for _, r := range rows {
			b.WriteString(v.sets.Sets[r].Line() + "\r\n")
		}
		return b.String()
	}
	l.Changed = func() { d.setModified(true) }
	v.list = l
}

func (v *itemSetView) draw(d *sdataDoc)      { v.list.Draw() }
func (v *itemSetView) editMenu(d *sdataDoc)  { v.list.EditMenu() }
func (v *itemSetView) toolsMenu(d *sdataDoc) {}

// export writes the "----" text of the sets (ItemSets::exportText), the
// only format the view knows.
func (v *itemSetView) export(format string) (string, bool) { return v.sets.Text(), true }

// editSet is the "Item Set Editor" (SetEditor): the name and the 13 items
// with their names and bonus texts.
func (a *App) editSet(set *sdata.ItemSet, modified func()) {
	items := a.clientItems()
	name := set.Name
	changed := false
	id := "Item Set Editor##" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowSizeV(imgui.NewVec2(660, 430), imgui.CondAppearing)
		done := false
		if imgui.BeginPopupModalV(id, nil, 0) {
			imgui.TextUnformatted("Name:")
			imgui.SameLine()
			imgui.SetNextItemWidth(320)
			imgui.InputTextWithHint("##name", "", &name, 0, nil)
			imgui.TextUnformatted("Items:")
			flags := imgui.TableFlagsBorders | imgui.TableFlagsRowBg | imgui.TableFlagsSizingFixedFit
			if imgui.BeginTableV("setitems", 4, flags, imgui.NewVec2(0, imgui.ContentRegionAvail().Y-imgui.FrameHeightWithSpacing()), 0) {
				imgui.TableSetupColumnV("Type", imgui.TableColumnFlagsWidthFixed, 50, 0)
				imgui.TableSetupColumnV("TypeID", imgui.TableColumnFlagsWidthFixed, 50, 0)
				imgui.TableSetupColumnV("[Item's Name]", imgui.TableColumnFlagsWidthFixed, 160, 0)
				imgui.TableSetupColumnV("Bonus", imgui.TableColumnFlagsWidthStretch, 0, 0)
				imgui.TableHeadersRow()
				for i := 0; i < 13; i++ {
					imgui.TableNextRow()
					imgui.TableNextColumn()
					if imgui.SelectableBoolV(strconv.Itoa(int(set.Types[i]))+"##"+strconv.Itoa(i), false,
						imgui.SelectableFlagsSpanAllColumns|imgui.SelectableFlagsAllowDoubleClick, imgui.NewVec2(0, 0)) &&
						imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
						a.editSetItem(set, i, func() { changed = true })
					}
					imgui.TableNextColumn()
					imgui.TextUnformatted(strconv.Itoa(int(set.IDs[i])))
					imgui.TableNextColumn()
					if it := items.Find(int(set.Types[i]), int(set.IDs[i])); it != nil {
						imgui.TextUnformatted(it.Name)
					}
					imgui.TableNextColumn()
					imgui.TextUnformatted(strings.ReplaceAll(set.Bonus[i], "\r\n", " "))
				}
				imgui.EndTable()
			}
			if imgui.Button("OK") {
				if name != set.Name {
					set.Name = name
					changed = true
				}
				done = true
			}
			imgui.SameLine()
			if imgui.Button("Cancel") {
				done = true
			}
			if done {
				imgui.CloseCurrentPopup()
			}
			a.inner()
			imgui.EndPopup()
		}
		if done && changed {
			modified()
		}
		return done
	})
}

// editSetItem is the "Item Editor" of one set slot (ItemSetEditor).
func (a *App) editSetItem(set *sdata.ItemSet, i int, changed func()) {
	t := int32(set.Types[i])
	tid := int32(set.IDs[i])
	bonus := strings.ReplaceAll(set.Bonus[i], "\r\n", "\n")
	id := "Item Editor##" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowSizeV(imgui.NewVec2(270, 260), imgui.CondAppearing)
		done := false
		if imgui.BeginPopupModalV(id, nil, 0) {
			imgui.TextUnformatted("Additional Item:")
			imgui.SetNextItemWidth(60)
			imgui.InputInt("Type", &t)
			imgui.SetNextItemWidth(60)
			imgui.InputInt("TypeID", &tid)
			imgui.TextUnformatted("Bonus description:")
			imgui.InputTextMultiline("##bonus", &bonus, imgui.NewVec2(-1, 90), 0, nil)
			if imgui.Button("OK") {
				set.Types[i] = int16(t)
				set.IDs[i] = int16(tid)
				set.Bonus[i] = strings.ReplaceAll(bonus, "\n", "\r\n")
				changed()
				done = true
			}
			imgui.SameLine()
			if imgui.Button("Cancel") {
				done = true
			}
			if done {
				imgui.CloseCurrentPopup()
			}
			a.inner()
			imgui.EndPopup()
		}
		return done
	})
}
