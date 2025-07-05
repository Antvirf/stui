package main

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
	"github.com/antvirf/stui/internal/view"
)

func main() {
	config.Configure()

	app := view.InitializeApplication()
	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh()

	// Enable log buffering after UI is set up, such that logs
	// are only printed after the UI has exited.
	logger.EnableBuffering()

	err := app.App.
		SetRoot(app.MainFlex, true).
		EnableMouse(true).
		Run()

	logger.LogFlush()

	if err != nil {
		panic(err)
	}
}
