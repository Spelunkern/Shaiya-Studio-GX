package app

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
)

// task is a long operation that runs in a goroutine behind a modal
// progress dialog (ArchiveDialog, BrowseFoldersDialog in shStudio). It
// implements archive.Notifier.
type task struct {
	title   string
	total   int64
	pos     atomic.Int64
	cancel  atomic.Bool
	mu      sync.Mutex
	status  string
	file    string
	done    atomic.Bool
	err     error
	opened  bool
	finish  func(err error) // runs on the UI thread when the work ends
	stopped bool
}

func (t *task) SetStatus(text string) bool {
	t.mu.Lock()
	t.status = text
	t.mu.Unlock()
	return !t.cancel.Load()
}

func (t *task) Progress(name string, units int64) bool {
	t.mu.Lock()
	t.file = name
	t.mu.Unlock()
	t.pos.Add(units)
	return !t.cancel.Load()
}

func (t *task) SetRange(n int64) { t.total = n }

// startTask runs work in a goroutine and shows the progress dialog until it
// ends; finish then runs on the UI thread.
func (a *App) startTask(title string, total int64, work func(t *task) error, finish func(err error)) {
	t := &task{title: title, total: total, finish: finish}
	a.busy = true
	go func() {
		t.err = work(t)
		t.done.Store(true)
		a.wake()
	}()
	a.pushModal(func() bool { return a.drawTask(t) })
}

func (a *App) drawTask(t *task) bool {
	if t.done.Load() {
		if imgui.IsPopupOpenStr(t.title) {
			imgui.CloseCurrentPopup()
		}
		a.busy = false
		if t.finish != nil {
			t.finish(t.err)
		}
		return true
	}
	if !t.opened {
		imgui.OpenPopupStr(t.title)
		t.opened = true
	}
	imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
	imgui.SetNextWindowSizeV(imgui.NewVec2(420, 0), imgui.CondAppearing)
	if imgui.BeginPopupModalV(t.title, nil, imgui.WindowFlagsNoResize) {
		t.mu.Lock()
		status, file := t.status, t.file
		t.mu.Unlock()
		imgui.TextUnformatted(status)
		imgui.TextUnformatted(file)
		frac := float32(0)
		if t.total > 0 {
			frac = float32(t.pos.Load()) / float32(t.total)
			if frac > 1 {
				frac = 1
			}
		}
		imgui.ProgressBarV(frac, imgui.NewVec2(-1, 0), fmt.Sprintf("%d%%", int(frac*100)))
		if t.cancel.Load() {
			imgui.BeginDisabled()
			imgui.Button("Cancel")
			imgui.EndDisabled()
		} else if imgui.Button("Cancel") {
			t.cancel.Store(true)
		}
		a.inner()
		imgui.EndPopup()
	}
	a.wake()
	return false
}
