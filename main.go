package main

import (
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/view"
	"github.com/rivo/tview"
)

func main() {
	config.Configure()
	app := &view.App{
		App:                    tview.NewApplication(),
		Pages:                  tview.NewPages(),
		RequestTimeout:         config.RequestTimeout * time.Second,
		DebugMultiplier:        config.DebugMultiplier,
		SearchDebounceInterval: config.SearchDebounceInterval * time.Millisecond,
	}

	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh(config.RefreshInterval * time.Second)

	if err := app.App.
		SetRoot(app.MainGrid, true).
		EnableMouse(false).
		Run(); err != nil {
		panic(err)
	}
}
