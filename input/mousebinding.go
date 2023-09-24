package input

import (
	"time"

	"github.com/seyys/sticky-display/desktop"
	"github.com/seyys/sticky-display/store"
)

var (
	workspace *desktop.Workspace // Stores last active workspace
)

func BindMouse(tr *desktop.Tracker) {
	poll(100, func() {
		store.PointerUpdate(store.X)

		// Update systray icon
		ws := tr.ActiveWorkspace()
		if ws != workspace {
			workspace = ws
		}
	})
}

func poll(t time.Duration, fun func()) {
	fun()
	go func() {
		for range time.Tick(t * time.Millisecond) {
			_, err := store.X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			fun()
		}
	}()
}
