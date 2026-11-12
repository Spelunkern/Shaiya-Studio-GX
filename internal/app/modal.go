package app

import (
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/ui/filedlg"
)

// Message box icons (MB_ICON* values passed to AlertDial::alert).
const (
	iconInfo  = 0x40
	iconError = 0x10
	iconWarn  = 0x30
)

// pushModal queues a modal UI step; draw returns true when it is finished.
// Modals run one at a time, in order (the C++ code nests DoModal calls).
func (a *App) pushModal(draw func() bool) { a.modals = append(a.modals, draw) }

func (a *App) drawModals() {
	if len(a.modals) == 0 {
		return
	}
	if a.modals[0]() {
		a.modals = a.modals[1:]
	}
}

// alert shows a message box (AlertDial::alert); then runs after OK.
func (a *App) alert(text string, icon int, then func()) {
	title := "Shaiya Studio GX"
	id := title + "##alert" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
		done := false
		if imgui.BeginPopupModalV(id, nil, imgui.WindowFlagsAlwaysAutoResize) {
			switch icon {
			case iconError:
				imgui.TextColored(imgui.NewVec4(0.9, 0.2, 0.2, 1), "Error")
			case iconWarn:
				imgui.TextColored(imgui.NewVec4(0.9, 0.7, 0.1, 1), "Warning")
			}
			for _, line := range strings.Split(text, "\n") {
				imgui.TextUnformatted(line)
			}
			if imgui.Button("OK") || imgui.IsKeyPressedBool(imgui.KeyEnter) {
				imgui.CloseCurrentPopup()
				done = true
			}
			imgui.EndPopup()
		}
		if done && then != nil {
			then()
		}
		return done
	})
}

// confirm asks a yes/no question.
func (a *App) confirm(text string, then func(yes bool)) {
	id := "Shaiya Studio GX##confirm" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
		done, yes := false, false
		if imgui.BeginPopupModalV(id, nil, imgui.WindowFlagsAlwaysAutoResize) {
			for _, line := range strings.Split(text, "\n") {
				imgui.TextUnformatted(line)
			}
			if imgui.Button("Yes") {
				done, yes = true, true
			}
			imgui.SameLine()
			if imgui.Button("No") || imgui.IsKeyPressedBool(imgui.KeyEscape) {
				done = true
			}
			if done {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		if done {
			then(yes)
		}
		return done
	})
}

// pickFile shows the file/folder dialog and calls then with the result.
func (a *App) pickFile(d *filedlg.Dialog, then func(path string)) {
	a.pushModal(func() bool {
		done, ok := d.Draw()
		if done && ok {
			then(d.Result())
		}
		return done
	})
}

func (a *App) nextID() int {
	a.idSeq++
	return a.idSeq
}

// Answers of askSave.
const (
	answerYes = iota
	answerNo
	answerCancel
)

// askSave asks whether to save a modified document (Doc::onClose).
func (a *App) askSave(path string, then func(answer int)) {
	id := "Shaiya Studio GX##save" + strconv.Itoa(a.nextID())
	opened := false
	a.pushModal(func() bool {
		if !opened {
			imgui.OpenPopupStr(id)
			opened = true
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
		answer := -1
		if imgui.BeginPopupModalV(id, nil, imgui.WindowFlagsAlwaysAutoResize) {
			imgui.TextUnformatted("Do you want to save the changes made to:")
			imgui.TextUnformatted(path)
			if imgui.Button("Yes") {
				answer = answerYes
			}
			imgui.SameLine()
			if imgui.Button("No") {
				answer = answerNo
			}
			imgui.SameLine()
			if imgui.Button("Cancel") || imgui.IsKeyPressedBool(imgui.KeyEscape) {
				answer = answerCancel
			}
			if answer >= 0 {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
		if answer >= 0 {
			then(answer)
			return true
		}
		return false
	})
}
