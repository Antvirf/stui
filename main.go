package main

import (
	"time"

	"github.com/antvirf/stui/internal/view"
	"github.com/rivo/tview"
)

func main() {
	DebugMultiplier := 1 // Default multiplier value for debugging
	app := &view.App{
		App:             tview.NewApplication(),
		Pages:           tview.NewPages(),
		RefreshInterval: 60 * time.Second,
		RequestTimeout:  5 * time.Second, // Must be less than refreshInterval
		DebugMultiplier: DebugMultiplier,
	}

	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh()

	if err := app.App.SetRoot(app.MainGrid, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
