package app

import (
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/content"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/prefs"
	"github.com/Spelunkern/Shaiya-Studio-GX/internal/sdata"
)

// clientData holds the tables shared by the views of a workspace
// (ClientData): they are loaded on first use and dropped when the
// workspace closes.
type clientData struct {
	items    *sdata.Items
	monsters *sdata.Monsters
}

// usesDBFiles reports whether the BinarySData DB files are the primary
// source ("SData" / "Primary" = "DBfiles").
func usesDBFiles() bool {
	return strings.EqualFold(prefs.Default.String("SData", "Primary", "undefined"), "DBfiles")
}

// loadSData loads a shared table from path (ClientData::loadSData). DB
// files also load their language file.
func (a *App) loadSData(data tableData, path string, f sdata.Format) bool {
	p := content.Current
	if p == nil {
		return false
	}
	file := p.Root().FindPath(path)
	if file == nil {
		return false
	}
	raw, err := p.ReadFile(file, true)
	if err != nil {
		return false
	}
	cp := 0
	if p.IsArchive() {
		cp = p.CodePage()
	}
	if data.Read(raw, cp, f) != nil {
		return false
	}
	if f == sdata.Flnt {
		language := prefs.Default.Int("SData", "language", 0)
		if language == 0 {
			a.alert("You are loading a binary-only file ("+file.Name+") and no prefered language was defined.\nUse the settings dialog to define your language.", iconError, nil)
			return false
		}
		prefix := data.(interface{ FilePrefix() string }).FilePrefix()
		suffix := sdata.CountryName(language)
		if strings.Contains(strings.ToLower(prefix), "binarysdata") {
			suffix = sdata.CountryCode(language)
		}
		text, err := readFile(prefix+suffix+".SData", false)
		if err != nil {
			return false
		}
		data.Read(text, cp, sdata.Ltxt)
	}
	file.State |= 1
	return true
}

// clientItems returns the item table of the workspace, loading Item.SData
// or DBItemData.SData (ClientData::loadSData for m_items).
func (a *App) clientItems() *sdata.Items {
	cd := &a.client.data
	if cd.items != nil && len(cd.items.Items) > 0 {
		return cd.items
	}
	items := sdata.NewItems()
	ok := false
	if usesDBFiles() {
		ok = a.loadSData(items, `BinarySData\DBItemData.sdata`, sdata.Flnt)
	} else {
		ok = a.loadSData(items, `Item\item.sdata`, sdata.File)
	}
	if ok {
		cd.items = items
	}
	return items
}
