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

type modal struct {
	draw func() bool
	done bool
}

// pushModal adds a modal UI step; draw returns true when it is finished.
// Modals form a stack: each open modal draws the next ones inside its own
// popup (inner), so a dialog opened from another one stays on top of it,
// like the nested DoModal calls of shStudio.
func (a *App) pushModal(draw func() bool) { a.modals = append(a.modals, &modal{draw: draw}) }

func (a *App) drawModals() {
	a.modalLevel = -1
	a.runModal(0)
	kept := a.modals[:0]
	for _, m := range a.modals {
		if !m.done {
			kept = append(kept, m)
		}
	}
	a.modals = kept
}

func (a *App) runModal(i int) {
	if i >= len(a.modals) {
		return
	}
	prev := a.modalLevel
	a.modalLevel = i
	m := a.modals[i]
	if !m.done && m.draw() {
		m.done = true
	}
	a.modalLevel = prev
}

// inner draws the modals stacked above the current one; every modal calls
// it inside its popup, before EndPopup.
func (a *App) inner() { a.runModal(a.modalLevel + 1) }

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
			a.inner()
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
			a.inner()
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
	d.Inner = a.inner
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
			a.inner()
			imgui.EndPopup()
		}
		if answer >= 0 {
			then(answer)
			return true
		}
		return false
	})
}
