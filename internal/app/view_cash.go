package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/archive"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
)

// cashView shows Cash.SData / DBItemSellData (CashView + ProductTab +
// ProductTile): one tab per bag, products shown as tiles, icons or a
// report.
type cashView struct {
	app      *App
	doc      *sdataDoc
	products *sdata.CashProducts
	bags     []int32
	tab      int32
	mode     int // 0 icon, 1 report, 4 tile (ProductTab::m_tileView)
	selected *sdata.CashProduct
}

func init() {
	registerView(func(a *App, ext string, f *archive.File, reload bool) Document {
		name := strings.ToLower(f.Name)
		if name != "cash.sdata" && name != "dbitemselldata.sdata" {
			return nil
		}
		p := sdata.NewCashProducts(1024)
		v := &cashView{app: a, products: p, mode: 4}
		d := &sdataDoc{file: f, data: p, view: v, hasTranslation: true, detectCP: true, decrypt: true}
		v.doc = d
		return d
	})
}

func (v *cashView) loaded(d *sdataDoc) { v.bags = v.products.Bags() }

func (v *cashView) bagProducts(bag int32) []*sdata.CashProduct {
	var out []*sdata.CashProduct
	for _, p := range v.products.Products {
		if p.Bag == bag {
			out = append(out, p)
		}
	}
	return out
}

func (v *cashView) icons() bool {
	return !v.app.iconLoader().Strip("icon_somo2").Empty()
}

func (v *cashView) draw(d *sdataDoc) {
	if !imgui.BeginTabBar("cashbags") {
		return
	}
	for _, bag := range v.bags {
		if !imgui.BeginTabItem(sdata.BagName(int(bag)) + "##bag" + strconv.Itoa(int(bag))) {
			continue
		}
		v.tab = bag
		v.drawBag(bag)
		imgui.EndTabItem()
	}
	imgui.EndTabBar()
}

func (v *cashView) drawBag(bag int32) {
	items := v.bagProducts(bag)
	set := v.app.iconLoader().Strip("icon_somo2")
	if imgui.BeginChildStrV("##tiles", imgui.NewVec2(0, 0), 0, 0) {
		switch v.mode {
		case 1: // report
			flags := imgui.TableFlagsBorders | imgui.TableFlagsRowBg | imgui.TableFlagsScrollY | imgui.TableFlagsResizable | imgui.TableFlagsSizingFixedFit
			if imgui.BeginTableV("cashreport", 4, flags, imgui.NewVec2(0, 0), 0) {
				imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthFixed, 200, 0)
				imgui.TableSetupColumnV("Cost", imgui.TableColumnFlagsWidthFixed, 50, 0)
				imgui.TableSetupColumnV("Code", imgui.TableColumnFlagsWidthFixed, 80, 0)
				imgui.TableSetupColumnV("Description", imgui.TableColumnFlagsWidthFixed, 250, 0)
				imgui.TableHeadersRow()
				for _, p := range items {
					imgui.TableNextRow()
					imgui.TableNextColumn()
					v.app.drawIcon(set, p.IconID(), imgui.TextLineHeight())
					imgui.SameLine()
					v.selectable(p, p.Name, imgui.SelectableFlagsSpanAllColumns, imgui.NewVec2(0, 0))
					imgui.TableNextColumn()
					imgui.TextUnformatted(strconv.Itoa(int(p.Cost)))
					imgui.TableNextColumn()
					imgui.TextUnformatted(p.Code)
					imgui.TableNextColumn()
					imgui.TextUnformatted(strings.ReplaceAll(p.Desc, "\r\n", " "))
				}
				imgui.EndTable()
			}
		default:
			w := float32(320)
			if v.mode == 0 {
				w = 96
			}
			perRow := int(imgui.ContentRegionAvail().X / (w + 8))
			if perRow < 1 {
				perRow = 1
			}
			for i, p := range items {
				if i%perRow != 0 {
					imgui.SameLine()
				}
				imgui.BeginGroup()
				pos := imgui.CursorPos()
				h := float32(52)
				if v.mode == 0 {
					h = 70
				}
				v.selectable(p, "##tile"+strconv.Itoa(i), 0, imgui.NewVec2(w, h))
				imgui.SetCursorPos(pos)
				if v.mode == 0 {
					imgui.SetCursorPosX(pos.X + (w-32)/2)
					v.app.drawIcon(set, p.IconID(), 32)
					imgui.PushTextWrapPosV(pos.X + w)
					imgui.TextUnformatted(p.Name)
					imgui.PopTextWrapPos()
				} else {
					v.app.drawIcon(set, p.IconID(), 32)
					imgui.SameLine()
					imgui.BeginGroup()
					imgui.TextUnformatted(p.Name)
					imgui.TextDisabled(p.Code)
					imgui.PushTextWrapPosV(pos.X + w)
					imgui.TextUnformatted(firstLine(p.Desc))
					imgui.PopTextWrapPos()
					imgui.EndGroup()
				}
				imgui.SetCursorPos(pos)
				imgui.Dummy(imgui.NewVec2(w, h))
				imgui.EndGroup()
			}
		}
		if imgui.BeginPopupContextWindow() {
			v.editMenu(v.doc)
			imgui.EndPopup()
		}
		if imgui.IsWindowFocused() {
			switch {
			case imgui.IsKeyPressedBool(imgui.KeyInsert):
				v.create()
			case imgui.IsKeyPressedBool(imgui.KeyDelete):
				v.deleteSelected()
			case imgui.IsKeyPressedBool(imgui.KeyEnter) && v.selected != nil:
				v.edit(v.selected)
			}
		}
	}
	imgui.EndChild()
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}

func (v *cashView) selectable(p *sdata.CashProduct, label string, flags imgui.SelectableFlags, size imgui.Vec2) {
	if imgui.SelectableBoolV(label, v.selected == p, flags|imgui.SelectableFlagsAllowDoubleClick, size) {
		v.selected = p
		if imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
			v.edit(p)
		}
	}
	if imgui.IsItemClickedV(imgui.MouseButtonRight) {
		v.selected = p
	}
}

// create adds a product to the current bag through the editor
// (ProductTile::onCommand 0xE12A).
func (v *cashView) create() {
	p := sdata.NewCashProduct(v.products.Version, 0)
	p.Bag = v.tab
	v.app.editProduct(p, v.products, func() {
		v.doc.setModified(true)
		v.products.Products = append(v.products.Products, p)
		v.bags = v.products.Bags()
		v.selected = p
	})
}

func (v *cashView) edit(p *sdata.CashProduct) {
	v.app.editProduct(p, v.products, func() {
		v.doc.setModified(true)
		v.bags = v.products.Bags()
	})
}

func (v *cashView) deleteSelected() {
	p := v.selected
	if p == nil {
		return
	}
	v.app.confirm("Are you sure you want to delete this item ?", func(yes bool) {
		if !yes {
			return
		}
		for i, x := range v.products.Products {
			if x == p {
				v.products.Products = append(v.products.Products[:i], v.products.Products[i+1:]...)
				break
			}
		}
		v.selected = nil
		v.doc.setModified(true)
	})
}

func (v *cashView) editMenu(d *sdataDoc) {
	if imgui.MenuItemBoolV("Copy", "Ctrl+C", false, v.selected != nil) {
		if line, ok := v.selected.Line(sdata.TextCSV); ok {
			imgui.SetClipboardText(line)
		}
	}
	imgui.Separator()
	if imgui.MenuItemBoolV("Create item", "Ins", false, true) {
		v.create()
	}
	if imgui.MenuItemBoolV("Delete item", "Del", false, v.selected != nil) {
		v.deleteSelected()
	}
	imgui.Separator()
	// Menu 130 "ListView".
	for _, m := range []struct {
		name string
		mode int
	}{{"Icon", 0}, {"Report", 1}, {"Tile", 4}} {
		if imgui.MenuItemBoolV(m.name, "", v.mode == m.mode, true) {
			v.mode = m.mode
		}
	}
}

// toolsMenu is the Format menu of the cash view (CashView::updateMenu).
func (v *cashView) toolsMenu(d *sdataDoc) {
	if imgui.BeginMenu("Format") {
		mode := v.products.Version & 0xFFF
		if imgui.MenuItemBoolV("Episo 4 format, 24 items per product, single price", "", mode == 0x400, true) {
			v.products.SetVersion(0x400)
		}
		if imgui.MenuItemBoolV("Episo 7 format, 24 items per product, 2 prices", "", mode == 0x700, true) {
			v.products.SetVersion(v.products.Version&0xF000 | 0x700)
		}
		imgui.EndMenu()
	}
}

func (v *cashView) export(format string) (string, bool) { return v.products.Text(format), true }

// editProduct is the "Cash Product" dialog (ProductEditor); ok runs when
// the product was changed.
func (a *App) editProduct(p *sdata.CashProduct, all *sdata.CashProducts, ok func()) {
	name, code, desc := p.Name, p.Code, strings.ReplaceAll(p.Desc, "\r\n", "\n")
	bag, icon := p.Bag, p.Icon
	cost, cost2 := p.Cost, p.Cost2
	ids, counts := p.ItemID, p.ItemCount
	set := a.iconLoader().Strip("icon_somo2")
	id := "Cash Product##" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowSizeV(imgui.NewVec2(420, 560), imgui.CondAppearing)
		done := false
		if imgui.BeginPopupModalV(id, nil, 0) {
			if imgui.ImageButtonV("##icon", a.iconTexRef(set, int(icon)), imgui.NewVec2(32, 32),
				a.iconUV(set, int(icon), false), a.iconUV(set, int(icon), true), imgui.NewVec4(0, 0, 0, 0), imgui.NewVec4(1, 1, 1, 1)) {
				imgui.OpenPopupStr("##pickicon")
			}
			if imgui.BeginPopup("##pickicon") {
				if sel, ok := a.pickIcon(set); ok {
					icon = int32(sel)
					imgui.CloseCurrentPopup()
				}
				imgui.EndPopup()
			}
			imgui.SameLine()
			imgui.BeginGroup()
			imgui.SetNextItemWidth(200)
			imgui.InputTextWithHint("Product Name", "", &name, 0, nil)
			imgui.SetNextItemWidth(200)
			if imgui.BeginCombo("Product Code", code) {
				for _, o := range all.Products {
					if imgui.SelectableBool(o.Code) {
						code = o.Code
					}
				}
				imgui.EndCombo()
			}
			imgui.SetNextItemWidth(200)
			imgui.InputTextWithHint("##code", "code", &code, 0, nil)
			imgui.EndGroup()
			imgui.TextUnformatted("Product Description:")
			imgui.InputTextMultiline("##desc", &desc, imgui.NewVec2(-1, 74), 0, nil)
			imgui.SetNextItemWidth(125)
			if imgui.BeginCombo("Bag", sdata.BagName(int(bag))) {
				for _, b := range all.Bags() {
					if imgui.SelectableBoolV(sdata.BagName(int(b)), b == bag, 0, imgui.NewVec2(0, 0)) {
						bag = b
					}
				}
				imgui.EndCombo()
			}
			imgui.SameLine()
			imgui.SetNextItemWidth(60)
			imgui.InputInt("Cost", &cost)
			imgui.SameLine()
			if p.Version&0xFFF < 0x700 {
				imgui.BeginDisabled()
			}
			imgui.SetNextItemWidth(60)
			imgui.InputInt("Cost(2)", &cost2)
			if p.Version&0xFFF < 0x700 {
				imgui.EndDisabled()
			}
			if imgui.BeginTableV("items", 3, imgui.TableFlagsSizingStretchSame, imgui.NewVec2(0, 0), 0) {
				for row := 0; row < 8; row++ {
					imgui.TableNextRow()
					for col := 0; col < 3; col++ {
						k := col*8 + row
						imgui.TableNextColumn()
						imgui.PushIDInt(int32(k))
						imgui.TextUnformatted(fmt.Sprintf("# %d", k+1))
						imgui.SameLine()
						imgui.SetNextItemWidth(55)
						imgui.InputScalar("##id", imgui.DataTypeS32, uintptrOf(&ids[k]))
						imgui.SameLine()
						imgui.SetNextItemWidth(30)
						imgui.InputScalar("##n", imgui.DataTypeU8, uintptrOf8(&counts[k]))
						imgui.PopID()
					}
				}
				imgui.EndTable()
			}
			if imgui.Button("OK") {
				msg := ""
				switch {
				case name == "":
					msg = "You must provide a name (up to 50 char.) for the product."
				case code == "":
					msg = "You must provide a code (up to 20 char.) for the product."
				case desc == "":
					msg = "You must provide a description for the product."
				case bag < 0:
					msg = "You must select a bag to insert product into."
				case icon <= 0 && !set.Empty():
					msg = "You must select an icon for the product."
				default:
					for _, o := range all.Products {
						if o != p && strings.EqualFold(o.Code, code) {
							msg = "The given code is already used by another product."
							break
						}
					}
				}
				if msg != "" {
					a.alert(msg, iconError, nil)
				} else {
					p.Name, p.Code, p.Desc = name, code, strings.ReplaceAll(desc, "\n", "\r\n")
					p.Bag = bag
					if !set.Empty() {
						p.Icon = icon
					}
					p.Cost = cost
					if p.Version&0xFFF >= 0x700 {
						p.Cost2 = cost2
					}
					p.Title = fmt.Sprintf("%s - %d Pts", name, cost)
					p.ItemID, p.ItemCount = ids, counts
					p.ItemTotal = 0
					for _, x := range ids {
						if x != 0 {
							p.ItemTotal++
						}
					}
					ok()
					done = true
				}
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
