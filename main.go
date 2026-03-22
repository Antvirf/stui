package main

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
	"github.com/antvirf/stui/internal/state"
	"github.com/antvirf/stui/internal/view"
)

func main() {
	config.Configure()

	StuiState, err := state.InitializeStuiState(config.StateDirPath)
	if err != nil {
		logger.Debugf("state failed to initialize due to: %s", err)
	}

	app := view.InitializeApplication(StuiState)
	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh()

	// Enable log buffering after UI is set up, such that logs
	// are only printed after the UI has exited.
	logger.EnableBuffering()

	err = app.App.
		SetRoot(app.MainFlex, true).
		EnableMouse(!config.MouseDisabled).
		Run()

	app.SaveState() // Save app state before exiting
	logger.LogFlush()

	if err != nil {
		panic(err)
	}
}
