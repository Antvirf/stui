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

	err := app.App.
		SetRoot(app.MainFlex, true).
		EnableMouse(false).
		Run()
	if err != nil {
		panic(err)
	}
}
