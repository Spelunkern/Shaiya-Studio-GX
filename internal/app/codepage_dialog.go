package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/codepage"
)

// lastChosenCodepage is g_lastChosenCodepage.
var lastChosenCodepage = -1

var codepageRows = []struct {
	cp   int
	name string
}{
	{932, "Japanese (Shift_JIS)"},
	{936, "Chinese (Simplified) (GB2312)"},
	{949, "Korean (EUC-KR)"},
	{1250, "Central European"},
	{1251, "Cyrillic"},
	{1252, "Western European (Latin 1)"},
}

// chooseCodepage asks the code page used to decode the strings of fileName
// (ChooseCodepageDialog::run); then receives -1 on cancel.
func (a *App) chooseCodepage(fileName string, then func(cp int)) {
	cp := lastChosenCodepage
	if cp < 0 {
		cp = 0
	}
	text := strconv.Itoa(cp)
	what := "the file "
	switch {
	case strings.HasSuffix(fileName, `\`) || strings.HasSuffix(fileName, "/"):
		what = "the workspace "
	case strings.HasSuffix(strings.ToLower(fileName), ".saf"):
		what = "the archive "
	}
	id := "Codepage Selection##" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
		imgui.SetNextWindowSizeV(imgui.NewVec2(380, 340), imgui.CondAppearing)
		done, result := false, -1
		if imgui.BeginPopupModalV(id, nil, 0) {
			imgui.TextWrapped("Please select the codepage to be used to decode the sentences of\n" + what + fileName)
			flags := imgui.TableFlagsBorders | imgui.TableFlagsRowBg | imgui.TableFlagsScrollY
			if imgui.BeginTableV("cp", 3, flags, imgui.NewVec2(0, 140), 0) {
				imgui.TableSetupColumnV("CP", imgui.TableColumnFlagsWidthFixed, 40, 0)
				imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthFixed, 180, 0)
				imgui.TableSetupColumnV("AGE Clients", imgui.TableColumnFlagsWidthStretch, 0, 0)
				imgui.TableHeadersRow()
				for _, r := range codepageRows {
					imgui.TableNextRow()
					imgui.TableNextColumn()
					if imgui.SelectableBoolV(strconv.Itoa(r.cp), r.cp == cp, imgui.SelectableFlagsSpanAllColumns, imgui.NewVec2(0, 0)) {
						cp = r.cp
						text = strconv.Itoa(cp)
					}
					imgui.TableNextColumn()
					imgui.TextUnformatted(r.name)
					imgui.TableNextColumn()
				}
				imgui.EndTable()
			}
			imgui.SetNextItemWidth(80)
			if imgui.InputTextWithHint("Codepage", "", &text, imgui.InputTextFlagsCharsDecimal, nil) && len(text) >= 3 {
				if v, err := strconv.Atoi(text); err == nil {
					cp = v
				}
			}
			imgui.SameLine()
			imgui.TextUnformatted(codepageName(cp))
			if imgui.Button("OK") {
				done, result = true, cp
			}
			imgui.SameLine()
			if imgui.Button("Cancel") || imgui.IsKeyPressedBool(imgui.KeyEscape) {
				done = true
			}
			if done {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		if done {
			if result >= 0 {
				lastChosenCodepage = result
			}
			then(result)
		}
		return done
	})
}

func codepageName(cp int) string {
	for _, r := range codepageRows {
		if r.cp == cp {
			return r.name
		}
	}
	if codepage.Known(cp) {
		return fmt.Sprintf("Code page %d", cp)
	}
	return fmt.Sprintf("%d (Invalid)", cp)
}
