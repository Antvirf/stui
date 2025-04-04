package main

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/view"
)

func main() {
	config.Configure()

	app := view.InitializeApplication()
	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh(config.RefreshInterval)

	if err := app.App.
		SetRoot(app.MainGrid, true).
		EnableMouse(false).
		Run(); err != nil {
		panic(err)
	}
}
